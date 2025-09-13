# MultiView Monitor

k9s 스타일의 CLI UI로 FFmpeg 프로세스와 HLS 패키징을 실시간 모니터링하는 도구입니다.

## 특징

- **실시간 모니터링**: 1초 간격으로 자동 업데이트
- **유연한 채널 설정**: 1-99개 채널까지 설정 가능 (기본 24개)
- **설정 가능한 경로**: HLS 패키지 폴더 경로 커스터마이징
- **전체 화면 UI**: 터미널 창을 꽉 채우는 반응형 레이아웃
- **직관적인 UI**: 메인 화면과 상세 화면 간의 부드러운 전환

## 빌드 방법

```bash
go build -o multiview-monitor ./cmd/monitor
```

## 사용 방법

### 1. 기본 실행
```bash
./multiview-monitor
```

### 2. 설정 파일 생성 및 사용
```bash
# 기본 설정 파일 생성
./multiview-monitor --generate-config

# 설정 파일 편집 후 실행
./multiview-monitor -f multiview-monitor.yaml

# 커스텀 설정 파일 사용
./multiview-monitor -f my-custom-config.yaml
```

### 3. 명령행 옵션으로 실행 (설정 파일 오버라이드)
```bash
# HLS 폴더 경로 변경
./multiview-monitor -p /data/streaming

# 채널 수 변경 (12개 채널)
./multiview-monitor -c 12

# 설정 파일 + 명령행 옵션 조합
./multiview-monitor -f myconfig.yaml -c 16 -p /custom/path
```

### 명령행 옵션
#### 설정 관리
- `-f, --config`: 설정 파일 경로 지정
- `--generate-config`: 기본 설정 파일 생성
- `-h, --help`: 도움말 표시

#### 모니터링 옵션 (설정 파일 설정 오버라이드)
- `-p, --hls-path`: HLS 패키지 기본 경로 (기본값: `/output`)
- `-c, --channels`: 모니터링할 채널 수 (기본값: `24`)
- `-s, --start-port`: FFmpeg 시작 포트 번호 (기본값: `8001`)

### 설정 우선순위
```
명령행 옵션 > 설정 파일 > 기본값
```

### 설정 파일 자동 탐지
다음 경로에서 설정 파일을 자동으로 찾습니다:
1. `./multiview-monitor.yaml`
2. `./configs/multiview-monitor.yaml`  
3. `~/.multiview-monitor.yaml`

### 키보드 조작

#### 메인 화면
- `↑/↓`: 채널 선택
- `Tab`: FFmpeg/HLS 패널 간 전환
- `Enter`: 선택된 채널의 상세 정보 보기
- `q`: 프로그램 종료

#### 상세 화면
- `Esc`: 메인 화면으로 돌아가기
- `↑/↓`: 스크롤
- `PgUp/PgDn`: 페이지 단위 스크롤
- `q`: 프로그램 종료

## 화면 구성

### 메인 화면 (전체 터미널 창 사용)
```
┌─ MultiView Monitor - /output (24 channels) ────────────────────────────────────────────────────────────┐
│ FFmpeg Processes (24)                        │ HLS Packages (24)                                      │
│ ┌─────┬────────┬────────┬────────┬─────────┐ │ ┌─────┬─────────────┬──────────┬──────┬──────┬────────┐ │
│ │ Ch  │ Port   │ PID    │ Status │ Command │ │ │ Ch  │ Path        │ Latest   │ M3U8 │ Segs │ Size   │ │
│ ├─────┼────────┼────────┼────────┼─────────┤ │ ├─────┼─────────────┼──────────┼──────┼──────┼────────┤ │
│ │ch01 │ :8001  │ 1234   │ RUN    │ ffmpeg..│ │ │ch01 │ channel01   │ seg_1234 │live  │ 1234 │ 2.5GB  │ │
│ │ch02 │ :8002  │ 5678   │ RUN    │ ffmpeg..│ │ │ch02 │ channel02   │ seg_1235 │index │ 1235 │ 2.6GB  │ │
│ │ch03 │ :8003  │ -      │ STOP   │ Not run │ │ │ch03 │ channel03   │ N/A      │ None │ 0    │ 0 B    │ │
│ │...  │ ...    │ ...    │ ...    │ ...     │ │ │...  │ ...         │ ...      │ ...  │ ...  │ ...    │ │
│ └─────┴────────┴────────┴────────┴─────────┘ │ └─────┴─────────────┴──────────┴──────┴──────┴────────┘ │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────┘
Status: 23/24 Running  Packages: 24/24  Updated: 14:32:15  [Tab] Switch  [↑↓] Select  [Enter] Details  [q] Quit
```

### 상세 화면
- FFmpeg 프로세스 정보 (포트, PID, 상태, 명령어)
- HLS 패키지 정보 (경로, 세그먼트 수, 파일 크기)
- M3U8 파일 내용 미리보기
- 최신 세그먼트 목록

## 요구사항

- Go 1.25 이상
- macOS/Linux (터미널 환경)

## 설정

### 기본 설정
- **채널 수**: 24개
- **FFmpeg 포트**: 8001-8024
- **HLS 출력 경로**: `/output/channel01` - `/output/channel24`

### 설정 파일 예시

**기본 생성되는 multiview-monitor.yaml:**
```yaml
# HLS package monitoring settings
hls:
  base_path: "/output"
  channel_dir_pattern: "channel%02d"

# FFmpeg process monitoring settings  
ffmpeg:
  start_port: 8001
  port_increment: 1

# Channel configuration
channels:
  count: 24
  id_format: "ch%02d"
  name_format: "Channel %02d"

# UI configuration
ui:
  refresh_interval: 1
  fullscreen: true
  theme: "dark"

# Logging configuration
logging:
  file: "monitor.log"
  level: "info"
```

**커스터마이징 예시:**
```bash
# 16개 채널, /data/hls 경로, 9001부터 포트 사용
# multiview-monitor.yaml 파일에서:
hls:
  base_path: "/data/hls"
channels:
  count: 16
ffmpeg:
  start_port: 9001

# 결과: 
# - 채널: ch01-ch16  
# - 포트: 9001-9016
# - 경로: /data/hls/channel01 - /data/hls/channel16
```

## 로그

프로그램 실행 중 발생하는 로그는 `monitor.log` 파일에 기록됩니다.