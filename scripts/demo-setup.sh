#!/bin/bash

# MultiView Monitor 데모 환경 자동 구성 스크립트
# 완전한 테스트 환경을 한 번에 설정합니다.

set -e

DEMO_DIR="/tmp/multiview-demo"
CHANNELS=6
START_PORT=8001

echo "🚀 MultiView Monitor 데모 환경 설정 시작"
echo "📁 데모 디렉토리: $DEMO_DIR"
echo "📺 채널 수: $CHANNELS"
echo "🔌 시작 포트: $START_PORT"
echo ""

# 기존 데모 환경 정리
if [ -d "$DEMO_DIR" ]; then
    echo "🧹 기존 데모 환경 정리 중..."
    pkill -f "multiview-demo" 2>/dev/null || true
    rm -rf "$DEMO_DIR"
fi

# 디렉토리 구조 생성
echo "📂 디렉토리 구조 생성 중..."
mkdir -p "$DEMO_DIR"/{hls-output,scripts,logs}

for i in $(seq -f "%02g" 1 $CHANNELS); do
    mkdir -p "$DEMO_DIR/hls-output/channel$i"
done

cd "$DEMO_DIR"

# FFmpeg 시뮬레이터 스크립트 생성
echo "🎬 FFmpeg 시뮬레이터 생성 중..."
cat > scripts/ffmpeg-simulator.sh << 'EOF'
#!/bin/bash
# FFmpeg 시뮬레이터 - 실제 FFmpeg처럼 보이는 더미 프로세스

CHANNEL=$1
PORT=$2
LOG_FILE="logs/ffmpeg-ch$CHANNEL.log"

if [ -z "$CHANNEL" ] || [ -z "$PORT" ]; then
    echo "Usage: $0 <channel> <port>"
    exit 1
fi

# 로그 파일 생성
mkdir -p logs
echo "$(date): Starting FFmpeg simulator for channel $CHANNEL on port $PORT" > "$LOG_FILE"

# 주기적으로 로그 업데이트
(
    while true; do
        echo "$(date): [ffmpeg] Processing frame $(($RANDOM % 10000 + 1000)) for channel $CHANNEL" >> "$LOG_FILE"
        sleep $((RANDOM % 5 + 3))
    done
) &

# 실제 FFmpeg 명령어처럼 보이는 프로세스 이름으로 실행
exec -a "ffmpeg -f rtsp -i rtsp://192.168.1.100:$PORT/stream$CHANNEL -c:v libx264 -preset fast -b:v 2000k -c:a aac -f hls -hls_time 6 -hls_list_size 10 -hls_flags delete_segments $DEMO_DIR/hls-output/channel$CHANNEL/index.m3u8" \
    bash -c "
        echo 'FFmpeg simulator for channel $CHANNEL started with PID $$' >> '$LOG_FILE'
        trap 'echo \"$(date): FFmpeg simulator for channel $CHANNEL stopped\" >> \"$LOG_FILE\"' EXIT
        while true; do
            sleep 1
        done
    "
EOF

chmod +x scripts/ffmpeg-simulator.sh

# HLS 생성기 스크립트 생성
echo "📺 HLS 생성기 생성 중..."
cat > scripts/hls-generator.sh << 'EOF'
#!/bin/bash
# HLS 파일 생성기 - 실제 스트리밍 세그먼트 파일 시뮬레이션

CHANNEL=$1
OUTPUT_DIR="$DEMO_DIR/hls-output/channel$CHANNEL"
LOG_FILE="logs/hls-ch$CHANNEL.log"

if [ -z "$CHANNEL" ]; then
    echo "Usage: $0 <channel>"
    exit 1
fi

echo "$(date): Starting HLS generator for channel $CHANNEL" > "$LOG_FILE"

# 시작 세그먼트 번호 (채널별로 다르게)
SEGMENT=$(($(echo $CHANNEL | sed 's/^0*//') * 1000 + RANDOM % 100))

while true; do
    # M3U8 헤더 생성
    cat > "$OUTPUT_DIR/index.m3u8" << EOM
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:$SEGMENT
#EXT-X-PLAYLIST-TYPE:VOD

EOM

    # 최근 8개 세그먼트 유지
    START_SEG=$((SEGMENT > 8 ? SEGMENT - 7 : 1))
    
    for i in $(seq $START_SEG $SEGMENT); do
        SEGFILE="segment_$(printf %06d $i).ts"
        DURATION=$(echo "scale=1; 5.8 + ($RANDOM % 5) / 10" | bc -l 2>/dev/null || echo "6.0")
        
        # 세그먼트 파일 생성 (실제 크기의 더미 데이터)
        if [ ! -f "$OUTPUT_DIR/$SEGFILE" ]; then
            SIZE=$((RANDOM % 200 + 800))  # 800-1000KB
            dd if=/dev/urandom of="$OUTPUT_DIR/$SEGFILE" bs=1024 count=$SIZE 2>/dev/null
            echo "$(date): Generated $SEGFILE (${SIZE}KB)" >> "$LOG_FILE"
        fi
        
        echo "#EXTINF:$DURATION," >> "$OUTPUT_DIR/index.m3u8"
        echo "$SEGFILE" >> "$OUTPUT_DIR/index.m3u8"
    done
    
    # 다양한 품질의 M3U8 파일들 생성
    for quality in "720p" "480p" "360p"; do
        bitrate=$(case $quality in
            "720p") echo "2500" ;;
            "480p") echo "1500" ;;
            "360p") echo "800" ;;
        esac)
        
        cat > "$OUTPUT_DIR/${quality}.m3u8" << EOM
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:$SEGMENT

EOM
        
        # 품질별 세그먼트 참조
        for i in $(seq $START_SEG $SEGMENT); do
            SEGFILE="segment_${quality}_$(printf %06d $i).ts"
            DURATION=$(echo "scale=1; 5.8 + ($RANDOM % 5) / 10" | bc -l 2>/dev/null || echo "6.0")
            
            if [ ! -f "$OUTPUT_DIR/$SEGFILE" ]; then
                SIZE=$((bitrate * 6 / 8 / 1000 + RANDOM % 100))  # 대략적인 크기
                dd if=/dev/urandom of="$OUTPUT_DIR/$SEGFILE" bs=1024 count=$SIZE 2>/dev/null
            fi
            
            echo "#EXTINF:$DURATION," >> "$OUTPUT_DIR/${quality}.m3u8"
            echo "$SEGFILE" >> "$OUTPUT_DIR/${quality}.m3u8"
        done
    done
    
    # 오래된 세그먼트 파일 정리 (10개 이상 유지)
    find "$OUTPUT_DIR" -name "segment_*.ts" -type f | sort | head -n -10 | xargs rm -f 2>/dev/null || true
    
    SEGMENT=$((SEGMENT + 1))
    
    # 랜덤한 간격으로 세그먼트 생성 (5-7초)
    SLEEP_TIME=$((RANDOM % 3 + 5))
    sleep $SLEEP_TIME
done
EOF

chmod +x scripts/hls-generator.sh

# 메인 시작 스크립트 생성
echo "🎯 메인 시작 스크립트 생성 중..."
cat > scripts/start-demo.sh << EOF
#!/bin/bash
# 데모 환경 시작 스크립트

echo "🚀 MultiView Monitor 데모 환경 시작"
echo ""

# 기존 프로세스 정리
pkill -f "ffmpeg-simulator.sh" 2>/dev/null || true
pkill -f "hls-generator.sh" 2>/dev/null || true
sleep 2

cd "$DEMO_DIR"

# 각 채널에 대해 FFmpeg 시뮬레이터와 HLS 생성기 시작
for i in \$(seq -f "%02g" 1 $CHANNELS); do
    PORT=\$((START_PORT - 1 + \$(echo \$i | sed 's/^0*//')))
    
    echo "📺 채널 \$i 시작 (포트: \$PORT)"
    
    # FFmpeg 시뮬레이터 시작
    ./scripts/ffmpeg-simulator.sh \$i \$PORT &
    
    # HLS 생성기 시작  
    ./scripts/hls-generator.sh \$i &
    
    sleep 0.5
done

echo ""
echo "✅ 데모 환경 시작 완료!"
echo ""
echo "📊 실행 중인 프로세스:"
ps aux | grep -E "(ffmpeg|hls-generator)" | grep -v grep | head -10

echo ""
echo "📁 생성된 파일들:"
find hls-output -name "*.m3u8" | head -6

echo ""
echo "🔧 MultiView Monitor 실행 명령어:"
echo "  ./multiview-monitor -p $DEMO_DIR/hls-output -c $CHANNELS -s $START_PORT"
echo ""
echo "⏹️  데모 중지: $DEMO_DIR/scripts/stop-demo.sh"
EOF

chmod +x scripts/start-demo.sh

# 정리 스크립트 생성
echo "🧹 정리 스크립트 생성 중..."
cat > scripts/stop-demo.sh << EOF
#!/bin/bash
# 데모 환경 정리 스크립트

echo "⏹️  MultiView Monitor 데모 환경 중지"

# 모든 시뮬레이터 프로세스 종료
pkill -f "ffmpeg-simulator.sh" 2>/dev/null || true
pkill -f "hls-generator.sh" 2>/dev/null || true
pkill -f "multiview-demo" 2>/dev/null || true

echo "✅ 모든 데모 프로세스가 중지되었습니다"

# 로그 파일 정리 옵션
echo ""
read -p "📝 로그 파일을 삭제하시겠습니까? [y/N] " -n 1 -r
echo
if [[ \$REPLY =~ ^[Yy]\$ ]]; then
    rm -rf logs/*
    echo "📝 로그 파일이 삭제되었습니다"
fi

# HLS 파일 정리 옵션  
read -p "📺 생성된 HLS 파일들을 삭제하시겠습니까? [y/N] " -n 1 -r
echo
if [[ \$REPLY =~ ^[Yy]\$ ]]; then
    find hls-output -name "*.ts" -delete 2>/dev/null || true
    find hls-output -name "*.m3u8" -delete 2>/dev/null || true
    echo "📺 HLS 파일들이 삭제되었습니다"
fi

echo ""
echo "🎯 데모를 다시 시작하려면: $DEMO_DIR/scripts/start-demo.sh"
EOF

chmod +x scripts/stop-demo.sh

# 설정 파일 생성
echo "⚙️  설정 파일 생성 중..."
cat > demo-config.yaml << EOF
# MultiView Monitor 데모 설정 파일

hls:
  base_path: "$DEMO_DIR/hls-output"
  channel_dir_pattern: "channel%02d"

ffmpeg:
  start_port: $START_PORT
  port_increment: 1

channels:
  count: $CHANNELS
  id_format: "ch%02d"
  name_format: "Channel %02d"

ui:
  refresh_interval: 1
  fullscreen: true
  theme: "dark"

logging:
  file: "$DEMO_DIR/logs/monitor.log"
  level: "info"

app:
  name: "MultiView Demo Monitor"
  version: "1.0.0-demo"
  description: "Demo environment for MultiView Monitor"
EOF

# README 파일 생성
cat > README-DEMO.md << EOF
# MultiView Monitor 데모 환경

이 디렉토리는 MultiView Monitor를 테스트하기 위한 완전한 데모 환경입니다.

## 빠른 시작

1. **데모 환경 시작:**
   \`\`\`bash
   $DEMO_DIR/scripts/start-demo.sh
   \`\`\`

2. **MultiView Monitor 실행:**
   \`\`\`bash
   ./multiview-monitor -f $DEMO_DIR/demo-config.yaml
   # 또는
   ./multiview-monitor -p $DEMO_DIR/hls-output -c $CHANNELS -s $START_PORT
   \`\`\`

3. **데모 환경 중지:**
   \`\`\`bash
   $DEMO_DIR/scripts/stop-demo.sh
   \`\`\`

## 포함된 내용

- **$CHANNELS개 FFmpeg 시뮬레이터**: 실제 FFmpeg 프로세스처럼 동작
- **HLS 파일 생성기**: 실시간으로 m3u8와 ts 파일들 생성
- **로그 시스템**: 각 채널별 로그 파일 생성
- **설정 파일**: 데모용 최적화된 설정

## 생성되는 파일들

- \`hls-output/channel01-06/\`: HLS 스트리밍 파일들
- \`logs/\`: FFmpeg 및 HLS 생성기 로그
- \`demo-config.yaml\`: 데모용 설정 파일

## 테스트 가능한 기능

- ✅ FFmpeg 프로세스 모니터링
- ✅ HLS 패키지 상태 확인
- ✅ M3U8 파일 내용 분석
- ✅ 실시간 세그먼트 추적
- ✅ 설정 파일 기능
- ✅ 화면 전환 (메인 ↔ 상세)

데모 환경에서 MultiView Monitor의 모든 기능을 안전하게 테스트할 수 있습니다!
EOF

echo ""
echo "🎉 데모 환경 설정 완료!"
echo ""
echo "📍 데모 위치: $DEMO_DIR"
echo "📺 채널 수: $CHANNELS개"
echo "🔌 포트 범위: $START_PORT-$((START_PORT + CHANNELS - 1))"
echo ""
echo "🚀 다음 단계:"
echo "  1. cd $DEMO_DIR"
echo "  2. ./scripts/start-demo.sh"
echo "  3. ./multiview-monitor -f demo-config.yaml"
echo ""
echo "📖 자세한 내용: $DEMO_DIR/README-DEMO.md"