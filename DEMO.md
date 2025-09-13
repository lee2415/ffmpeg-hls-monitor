# MultiView Monitor 데모 가이드

이 문서는 MultiView Monitor를 실제로 체험해볼 수 있는 완전한 테스트 환경 구성 방법을 제공합니다.

## 🚀 원클릭 데모 시작

### 1. 자동 데모 환경 구성
```bash
# 데모 환경 자동 구성 (권장)
./scripts/demo-setup.sh

# 데모 시작
cd /tmp/multiview-demo
./scripts/start-demo.sh

# MultiView Monitor 실행
./multiview-monitor -f demo-config.yaml
```

### 2. 즉시 확인 가능한 것들
- ✅ **6개 FFmpeg 프로세스** 실시간 모니터링
- ✅ **HLS 파일 생성** 실시간 추적
- ✅ **M3U8 파일 내용** 상세 분석
- ✅ **세그먼트 파일** 자동 생성 및 관리
- ✅ **화면 전환** (메인 ↔ 상세뷰)

## 🎯 데모 시나리오별 가이드

### 시나리오 1: 기본 모니터링 체험
```bash
# 1. 데모 환경 시작
./scripts/demo-setup.sh
cd /tmp/multiview-demo
./scripts/start-demo.sh

# 2. MultiView Monitor 실행
./multiview-monitor -f demo-config.yaml

# 3. 확인 포인트
# - FFmpeg 패널: 6개 프로세스 RUN 상태
# - HLS 패널: m3u8 파일들과 세그먼트 정보
# - Tab키로 패널 전환
# - Enter로 채널 상세 보기
```

### 시나리오 2: 장애 상황 시뮬레이션
```bash
# 데모 환경에서 프로세스 중단 테스트
pkill -f "channel03"

# MultiView Monitor에서 확인:
# - channel03이 STOP 상태로 변경
# - 해당 채널의 파일 생성 중단

# 5초 후 복구
cd /tmp/multiview-demo
./scripts/ffmpeg-simulator.sh 03 8003 &
./scripts/hls-generator.sh 03 &

# 복구 확인: channel03이 다시 RUN 상태로 변경
```

### 시나리오 3: 설정 변경 테스트
```bash
# 다른 경로와 채널 수로 테스트
mkdir -p /tmp/custom-test/output
./multiview-monitor -p /tmp/custom-test/output -c 12 -s 9001

# 설정 파일 커스터마이징
cp /tmp/multiview-demo/demo-config.yaml my-config.yaml
# my-config.yaml 편집 후
./multiview-monitor -f my-config.yaml
```

### 시나리오 4: Docker 환경 테스트
```bash
# Docker 컨테이너로 격리된 환경 테스트
docker-compose up -d

# 컨테이너 프로세스 확인
docker ps

# MultiView Monitor로 모니터링
./multiview-monitor -p ./test-data -c 8 -s 9001

# 정리
docker-compose down -v
```

## 🔧 수동 테스트 환경 구성

자동 스크립트 대신 수동으로 구성하고 싶다면:

### 1. 기본 디렉토리 준비
```bash
mkdir -p /tmp/manual-test/{hls-output,scripts}
cd /tmp/manual-test

# 채널 디렉토리 생성
for i in {01..06}; do
    mkdir -p hls-output/channel$i
done
```

### 2. FFmpeg 시뮬레이터 생성
```bash
# 간단한 FFmpeg 더미 프로세스
for i in {1..6}; do
    CHANNEL=$(printf "%02d" $i)
    PORT=$((8000 + i))
    
    # 백그라운드에서 FFmpeg 프로세스 시뮬레이션
    exec -a "ffmpeg -i rtmp://source:$PORT/stream -c:v libx264 -f hls hls-output/channel$CHANNEL/index.m3u8" \
        sleep infinity &
done
```

### 3. HLS 파일 생성
```bash
# 기본 M3U8 파일 생성
for i in {01..06}; do
    cat > hls-output/channel$i/index.m3u8 << EOF
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:1000

#EXTINF:6.0,
segment_001000.ts
#EXTINF:6.0,
segment_001001.ts
#EXTINF:6.0,
segment_001002.ts
EOF

    # 더미 세그먼트 파일들 생성
    for j in {1000..1002}; do
        dd if=/dev/urandom of="hls-output/channel$i/segment_$(printf %06d $j).ts" \
           bs=1024 count=$((RANDOM % 300 + 200)) 2>/dev/null
    done
done
```

### 4. MultiView Monitor 실행
```bash
./multiview-monitor -p /tmp/manual-test/hls-output -c 6 -s 8001
```

## 📊 기대되는 결과

### 메인 화면에서 확인할 것들
```
┌─ MultiView Monitor - /tmp/multiview-demo/hls-output (6 channels) ─┐
│ FFmpeg Processes (6)           │ HLS Packages (6)                 │
│ ┌─────┬────────┬──────┬─────┐  │ ┌─────┬─────────┬──────────┬───┐ │
│ │ Ch  │ Port   │ PID  │ St  │  │ │ Ch  │ Latest  │ M3U8     │Seg│ │
│ │ch01 │ :8001  │ 1234 │ RUN │  │ │ch01 │ seg_123 │ 4 files  │123│ │
│ │ch02 │ :8002  │ 5678 │ RUN │  │ │ch02 │ seg_124 │ 4 files  │124│ │
│ │...  │ ...    │ ...  │ ... │  │ │...  │ ...     │ ...      │...│ │
└─────────────────────────────────┴───────────────────────────────────┘
```

### 상세 화면에서 확인할 것들 (Enter로 진입)
- FFmpeg 프로세스 상세 정보
- M3U8 파일 전체 내용
- 최신 세그먼트 파일 목록
- 실시간 통계 (파일 크기, 세그먼트 수)

## 🛠️ 문제 해결

### 프로세스가 보이지 않는 경우
```bash
ps aux | grep ffmpeg
# 시뮬레이터가 실행 중인지 확인

# 다시 시작
cd /tmp/multiview-demo
./scripts/start-demo.sh
```

### HLS 파일이 생성되지 않는 경우
```bash
ls -la /tmp/multiview-demo/hls-output/channel*/
# 파일 존재 확인

tail -f /tmp/multiview-demo/hls-output/channel01/index.m3u8
# M3U8 파일 변화 실시간 모니터링
```

### 권한 문제
```bash
sudo chown -R $USER:$USER /tmp/multiview-demo
chmod +x /tmp/multiview-demo/scripts/*.sh
```

## 🧪 고급 테스트

### 성능 테스트
```bash
# 24채널로 확장 테스트
./multiview-monitor -p /tmp/multiview-demo/hls-output -c 24 -s 8001
```

### 네트워크 지연 시뮬레이션
```bash
# HLS 생성 지연 (파일 생성을 일시적으로 중단)
pkill -f "hls-generator"
sleep 30
# 30초 후 재시작하여 지연 복구 확인
```

### 메모리 사용량 모니터링
```bash
# MultiView Monitor 실행하면서 리소스 사용량 확인
top -p $(pgrep multiview-monitor)
```

## 🎉 데모 완료 후 정리

```bash
cd /tmp/multiview-demo
./scripts/stop-demo.sh

# 또는 완전 삭제
rm -rf /tmp/multiview-demo
```

이 데모 환경을 통해 MultiView Monitor의 모든 기능을 실제 상황과 유사하게 체험할 수 있습니다!