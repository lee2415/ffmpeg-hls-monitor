# MultiView Monitor 테스트 환경 구성 가이드

이 문서는 MultiView Monitor를 실제로 테스트해볼 수 있는 환경을 구성하는 방법을 설명합니다.

## 빠른 시작 (Docker 사용)

### 1. 전체 환경 한 번에 실행
```bash
# 테스트 환경 시작
docker-compose up -d

# MultiView Monitor 실행
./multiview-monitor -p /tmp/hls-output -c 6 -s 8001

# 정리
docker-compose down
```

### 2. 수동 환경 구성 (추천)

테스트를 위한 가짜 FFmpeg 프로세스와 HLS 파일들을 생성합니다.

## 수동 테스트 환경 구성

### 1. 디렉토리 준비
```bash
# 테스트 디렉토리 생성
mkdir -p /tmp/multiview-test/{hls-output,scripts}
cd /tmp/multiview-test

# 채널별 디렉토리 생성 (6개 채널)
for i in {01..06}; do
    mkdir -p hls-output/channel$i
done
```

### 2. FFmpeg 시뮬레이터 스크립트 생성

FFmpeg 프로세스를 시뮬레이션하는 스크립트들을 만듭니다:

```bash
# scripts/ffmpeg-sim.sh 생성
cat > scripts/ffmpeg-sim.sh << 'EOF'
#!/bin/bash
# FFmpeg 시뮬레이터 - 실제 FFmpeg처럼 동작하는 더미 프로세스

CHANNEL=$1
PORT=$2

if [ -z "$CHANNEL" ] || [ -z "$PORT" ]; then
    echo "Usage: $0 <channel> <port>"
    exit 1
fi

echo "Starting FFmpeg simulator for channel $CHANNEL on port $PORT"

# 실제 FFmpeg 명령어처럼 보이는 프로세스 이름 설정
exec -a "ffmpeg -f rtsp -i rtsp://source:$PORT/stream -c:v libx264 -c:a aac -f hls -hls_time 6 -hls_list_size 5 /tmp/multiview-test/hls-output/channel$CHANNEL/index.m3u8" \
    sleep infinity
EOF

chmod +x scripts/ffmpeg-sim.sh
```

### 3. HLS 파일 생성기 스크립트

실제 HLS 세그먼트 파일들을 생성하는 스크립트:

```bash
# scripts/hls-generator.sh 생성
cat > scripts/hls-generator.sh << 'EOF'
#!/bin/bash
# HLS 파일 생성기 - 실제 HLS 스트리밍을 시뮬레이션

CHANNEL=$1
OUTPUT_DIR="/tmp/multiview-test/hls-output/channel$CHANNEL"

if [ -z "$CHANNEL" ]; then
    echo "Usage: $0 <channel>"
    exit 1
fi

echo "Generating HLS files for channel $CHANNEL"

# 세그먼트 카운터
SEGMENT=1

while true; do
    # M3U8 파일 생성
    cat > "$OUTPUT_DIR/index.m3u8" << EOM
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:$SEGMENT

EOM

    # 최근 5개 세그먼트 표시
    for i in $(seq $((SEGMENT-4 > 1 ? SEGMENT-4 : 1)) $SEGMENT); do
        SEGFILE="segment_$(printf %06d $i).ts"
        
        # 세그먼트 파일 생성 (더미 데이터)
        if [ ! -f "$OUTPUT_DIR/$SEGFILE" ]; then
            dd if=/dev/urandom of="$OUTPUT_DIR/$SEGFILE" bs=1024 count=$((RANDOM % 500 + 200)) 2>/dev/null
        fi
        
        echo "#EXTINF:6.0," >> "$OUTPUT_DIR/index.m3u8"
        echo "$SEGFILE" >> "$OUTPUT_DIR/index.m3u8"
    done
    
    # 추가 M3U8 파일들 생성 (다양한 품질)
    for quality in "720p" "480p" "360p"; do
        cp "$OUTPUT_DIR/index.m3u8" "$OUTPUT_DIR/${quality}.m3u8"
        sed -i '' "s/segment_/segment_${quality}_/g" "$OUTPUT_DIR/${quality}.m3u8" 2>/dev/null || \
        sed -i "s/segment_/segment_${quality}_/g" "$OUTPUT_DIR/${quality}.m3u8" 2>/dev/null
    done
    
    SEGMENT=$((SEGMENT + 1))
    sleep 6  # 6초마다 새 세그먼트 생성
done
EOF

chmod +x scripts/hls-generator.sh
```

### 4. 테스트 환경 시작 스크립트

모든 것을 한 번에 시작하는 스크립트:

```bash
# scripts/start-test-env.sh 생성
cat > scripts/start-test-env.sh << 'EOF'
#!/bin/bash
# MultiView Monitor 테스트 환경 시작

echo "Starting MultiView Monitor test environment..."

# 기존 프로세스 정리
pkill -f "ffmpeg-sim.sh"
pkill -f "hls-generator.sh"

# FFmpeg 시뮬레이터 시작 (6개 채널)
for i in {1..6}; do
    CHANNEL=$(printf "%02d" $i)
    PORT=$((8000 + i))
    
    echo "Starting FFmpeg simulator for channel $CHANNEL on port $PORT"
    ./scripts/ffmpeg-sim.sh $CHANNEL $PORT &
    
    echo "Starting HLS generator for channel $CHANNEL"
    ./scripts/hls-generator.sh $CHANNEL &
    
    sleep 0.5
done

echo ""
echo "Test environment started!"
echo ""
echo "Running processes:"
ps aux | grep -E "(ffmpeg|hls-generator)" | grep -v grep

echo ""
echo "Generated files:"
find hls-output -name "*.m3u8" | head -10

echo ""
echo "Now run MultiView Monitor:"
echo "  ./multiview-monitor -p /tmp/multiview-test/hls-output -c 6 -s 8001"
echo ""
echo "To stop test environment:"
echo "  ./scripts/stop-test-env.sh"
EOF

chmod +x scripts/start-test-env.sh
```

### 5. 정리 스크립트

```bash
# scripts/stop-test-env.sh 생성
cat > scripts/stop-test-env.sh << 'EOF'
#!/bin/bash
# MultiView Monitor 테스트 환경 정리

echo "Stopping MultiView Monitor test environment..."

# 모든 시뮬레이터 프로세스 종료
pkill -f "ffmpeg-sim.sh"
pkill -f "hls-generator.sh"
pkill -f "sleep infinity"

# 생성된 파일들 정리 (선택사항)
read -p "Delete generated HLS files? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf hls-output/channel*/
    echo "HLS files deleted"
fi

echo "Test environment stopped"
EOF

chmod +x scripts/stop-test-env.sh
```

## 실행 방법

### 1. 테스트 환경 시작
```bash
cd /tmp/multiview-test
./scripts/start-test-env.sh
```

### 2. MultiView Monitor 실행
```bash
# 별도 터미널에서 실행
./multiview-monitor -p /tmp/multiview-test/hls-output -c 6 -s 8001
```

### 3. 확인사항
- **FFmpeg 패널**: 6개 프로세스가 RUN 상태로 표시
- **HLS 패널**: 6개 채널의 m3u8 파일들과 세그먼트 정보
- **상세 화면**: Enter로 채널 선택 시 M3U8 내용과 세그먼트 목록 확인

### 4. 정리
```bash
./scripts/stop-test-env.sh
```

## Docker를 사용한 테스트 환경

더 격리된 환경에서 테스트하려면:

```bash
# docker-compose.yml 생성 (별도 파일)
# 실행
docker-compose up -d

# MultiView Monitor 테스트
./multiview-monitor -p /tmp/docker-hls -c 8 -s 9001

# 정리
docker-compose down -v
```

## 고급 테스트 시나리오

### 1. 프로세스 장애 시뮬레이션
```bash
# 특정 채널 중단
pkill -f "channel03"

# 5초 후 재시작
sleep 5
./scripts/ffmpeg-sim.sh 03 8003 &
./scripts/hls-generator.sh 03 &
```

### 2. 네트워크 문제 시뮬레이션
```bash
# HLS 파일 생성 중단
pkill -f "hls-generator"

# 30초 후 재시작
sleep 30
for i in {1..6}; do
    CHANNEL=$(printf "%02d" $i)
    ./scripts/hls-generator.sh $CHANNEL &
done
```

### 3. 대용량 테스트
```bash
# 24채널로 확장 테스트
./multiview-monitor -p /tmp/multiview-test/hls-output -c 24 -s 8001
```

## 문제 해결

### 프로세스가 보이지 않는 경우
```bash
ps aux | grep ffmpeg
# ffmpeg 시뮬레이터가 실행 중인지 확인
```

### HLS 파일이 생성되지 않는 경우
```bash
ls -la /tmp/multiview-test/hls-output/channel*/
# 파일 생성 확인

tail -f /tmp/multiview-test/hls-output/channel01/index.m3u8
# M3U8 파일 변화 모니터링
```

### 권한 문제
```bash
chmod +x scripts/*.sh
sudo chown -R $USER:$USER /tmp/multiview-test
```

이 테스트 환경을 통해 MultiView Monitor의 모든 기능을 실제와 유사한 상황에서 확인할 수 있습니다!