# MultiView Monitor - Claude 개발 가이드

이 문서는 Claude Code를 사용하여 MultiView Monitor를 개발할 때 알아야 할 중요한 정보들을 담고 있습니다.

## 프로젝트 개요

### 목적
k9s 스타일의 CLI UI로 FFmpeg 프로세스와 HLS 패키징을 실시간 모니터링하는 도구

### 핵심 기능
- **실시간 모니터링**: 1초 간격으로 FFmpeg 프로세스와 HLS 패키지 상태 추적
- **24채널 지원**: 기본 24개 채널 (설정으로 변경 가능)
- **두 화면 모드**: 메인 목록 화면 + 상세 정보 화면
- **전체 화면 UI**: 터미널 창을 꽉 채우는 반응형 레이아웃
- **설정 파일 지원**: YAML 기반 설정 관리

## 아키텍처 구조

### 디렉토리 구조
```
monitorMultiview/
├── cmd/monitor/               # 메인 애플리케이션 엔트리포인트
│   └── main.go
├── internal/
│   ├── config/                # 설정 관리
│   │   └── channels.go
│   ├── monitor/               # 모니터링 로직
│   │   ├── ffmpeg.go         # FFmpeg 프로세스 모니터링
│   │   ├── hls.go            # HLS 패키징 모니터링
│   │   └── m3u8.go           # M3U8 파일 파서
│   └── ui/                    # TUI 인터페이스
│       ├── mainview.go       # 메인 목록 화면
│       ├── detailview.go     # 상세 정보 화면
│       └── styles.go         # UI 스타일 정의
├── configs/                   # 설정 파일들
│   └── multiview-monitor.yaml
├── examples/                  # 예시 설정 파일들
│   ├── production-config.yaml
│   └── development-config.yaml
├── README.md
└── go.mod
```

### 핵심 컴포넌트

#### 1. 설정 관리 (internal/config/channels.go)
- **GlobalConfig**: 전역 설정 객체
- **YAML 지원**: gopkg.in/yaml.v3 사용
- **설정 우선순위**: 명령행 옵션 > 설정 파일 > 기본값
- **자동 탐지**: 여러 경로에서 설정 파일 자동 검색

#### 2. 모니터링 시스템 (internal/monitor/)
- **FFmpegMonitor**: `ps aux` 명령어로 FFmpeg 프로세스 탐지
- **HLSMonitor**: 파일시스템 스캔으로 HLS 패키지 모니터링  
- **M3U8Parser**: M3U8 파일 내용 파싱 및 세그먼트 정보 추출

#### 3. UI 시스템 (internal/ui/)
- **Bubbletea**: TUI 프레임워크
- **Lipgloss**: 스타일링
- **Table/Viewport**: 데이터 표시 컴포넌트

## 기술 스택

### 주요 의존성
```go
// UI 프레임워크
github.com/charmbracelet/bubbletea v1.3.9
github.com/charmbracelet/lipgloss v1.1.0  
github.com/charmbracelet/bubbles v0.21.0

// 설정 관리
gopkg.in/yaml.v3 v3.0.1
```

### Go 버전 요구사항
- **Go 1.25 이상** (최신 Bubbletea 호환성)
- 이전 버전 사용시 의존성 버전 다운그레이드 필요

## 개발 시 주의사항

### 1. 코드 수정 시 고려사항

#### Bubbletea API 변경사항
- **높이/너비 동적 변경**: 새 버전에서는 테이블 재생성 필요
- **컬럼 너비 변경**: `WithColumns()` 메소드 대신 테이블 재생성
- **색상 렌더링**: `lipgloss.Color`에서 `lipgloss.NewStyle().Foreground()` 방식으로 변경

#### 설정 시스템 연동
- **GlobalConfig 참조**: 모든 하드코딩된 값들을 GlobalConfig로 교체
- **채널 수 동적화**: 24개 고정값 대신 `config.GlobalConfig.Channels.Count` 사용
- **경로 동적화**: 하드코딩된 경로 대신 설정 기반 경로 사용

#### 프로세스 모니터링
- **포트 스캔**: 설정된 포트 범위 내에서만 스캔
- **채널 매핑**: 포트 번호로 채널 ID 역추적 로직
- **파일 경로**: 설정된 패턴으로 디렉토리 구조 생성

### 2. 빌드 및 테스트

#### 빌드 명령어
```bash
go build -o multiview-monitor ./cmd/monitor
```

#### 테스트 시나리오
```bash
# 기본 실행
./multiview-monitor

# 설정 파일 생성
./multiview-monitor --generate-config

# 커스텀 설정 실행
./multiview-monitor -f myconfig.yaml

# 명령행 오버라이드
./multiview-monitor -f config.yaml -c 12 -p /custom/path

# 도움말 확인
./multiview-monitor --help
```

### 3. 에러 처리 패턴

#### 설정 파일 에러
- **파일 없음**: 기본값으로 폴백, 경고 메시지
- **YAML 파싱 에러**: 명확한 에러 메시지와 함께 종료
- **설정 검증 실패**: 구체적인 에러 내용과 유효 범위 표시

#### 모니터링 에러
- **프로세스 스캔 실패**: 빈 결과 반환, 로그 기록
- **파일 접근 실패**: N/A 표시, 계속 모니터링
- **M3U8 파싱 실패**: 에러 메시지 표시, 다른 파일들 계속 처리

### 4. 성능 최적화 고려사항

#### UI 렌더링
- **테이블 재생성**: 최소한으로 제한 (윈도우 리사이즈시만)
- **데이터 업데이트**: 1초 간격, 백그라운드에서 수행
- **메모리 사용**: 채널 수에 비례하여 증가

#### 파일 시스템 모니터링
- **디렉토리 스캔**: 존재하지 않는 디렉토리 사전 체크
- **파일 개수**: 많은 세그먼트 파일 처리시 성능 고려
- **캐싱**: 변경되지 않은 데이터 캐싱으로 성능 향상

## 확장 가능성

### 추가 가능한 기능들

#### 모니터링 확장
- **알림 시스템**: 프로세스 다운, 파일 생성 중단시 알림
- **로그 통합**: 실시간 로그 뷰어 추가
- **메트릭 수집**: Prometheus/Grafana 연동
- **원격 모니터링**: 네트워크를 통한 원격 서버 모니터링

#### UI 개선
- **테마 시스템**: 라이트/다크 테마 전환
- **레이아웃 커스터마이징**: 패널 크기 조절
- **필터링**: 특정 채널만 표시
- **정렬**: 다양한 기준으로 정렬

#### 설정 고도화
- **환경별 프로파일**: dev/staging/prod 프로파일
- **동적 설정 리로드**: 실행 중 설정 변경
- **설정 검증**: 더 상세한 유효성 검사

### 코드 확장 패턴

#### 새로운 모니터 추가
```go
// internal/monitor/newmonitor.go
type NewMonitor struct {
    // ...
}

func (m *NewMonitor) GetData() []NewData {
    // 구현
}
```

#### UI 컴포넌트 추가
```go
// internal/ui/newview.go  
type NewViewModel struct {
    // Bubbletea 모델 구현
}

func (m *NewViewModel) Update(msg tea.Msg) (*NewViewModel, tea.Cmd) {
    // 업데이트 로직
}

func (m *NewViewModel) View() string {
    // 렌더링 로직
}
```

#### 설정 확장
```go
// internal/config/channels.go
type Config struct {
    // 기존 필드들...
    NewFeature NewFeatureConfig `yaml:"new_feature"`
}

type NewFeatureConfig struct {
    Enabled bool   `yaml:"enabled"`
    Option  string `yaml:"option"`
}
```

## 문제 해결 가이드

### 자주 발생하는 문제들

#### 1. Go 버전 호환성
**증상**: `cannot range over datalen (variable of type int)` 에러
**해결**: Go 1.25 이상으로 업그레이드 또는 의존성 다운그레이드

#### 2. 테이블 렌더링 문제
**증상**: 테이블이 화면에 맞지 않음
**해결**: `recreateTablesWithSize()` 함수에서 너비 계산 로직 확인

#### 3. 설정 파일 로딩 실패
**증상**: `Error loading config file` 메시지
**해결**: YAML 문법 검사, 파일 권한 확인

#### 4. 프로세스 탐지 실패
**증상**: FFmpeg 프로세스가 표시되지 않음
**해결**: `ps aux` 명령어 결과 확인, 포트 범위 설정 검토

### 디버깅 팁

#### 로그 확인
```bash
tail -f monitor.log  # 실시간 로그 모니터링
```

#### 설정 확인
```bash
./multiview-monitor --help  # 현재 설정 확인
```

#### 개발 모드 실행
```go
// 개발시 디버그 정보 출력
fmt.Printf("Debug: %+v\n", someStruct)
```

## 마무리

이 프로젝트는 실용적인 CLI 도구로서 확장성과 유지보수성을 고려하여 설계되었습니다. 새로운 기능을 추가하거나 수정할 때는 위의 가이드라인을 참고하여 일관성 있는 코드를 작성해 주세요.

추가 질문이나 문제가 있을 때는 이 문서를 참조하여 해결하거나, 필요시 이 문서를 업데이트하여 다음 개발자에게 도움이 되도록 해주세요.