# MultiView Monitor 설치 가이드

## 빠른 설치

### Linux
```bash
# AMD64 (Intel/AMD)
curl -L https://github.com/your-repo/multiview-monitor/releases/latest/download/multiview-monitor_1.0.0_linux_amd64.tar.gz | tar -xz
sudo mv multiview-monitor_1.0.0_linux_amd64/multiview-monitor /usr/local/bin/
multiview-monitor --help

# ARM64 (Raspberry Pi 4, AWS Graviton 등)
curl -L https://github.com/your-repo/multiview-monitor/releases/latest/download/multiview-monitor_1.0.0_linux_arm64.tar.gz | tar -xz
sudo mv multiview-monitor_1.0.0_linux_arm64/multiview-monitor /usr/local/bin/
```

### macOS
```bash
# Intel Mac
curl -L https://github.com/your-repo/multiview-monitor/releases/latest/download/multiview-monitor_1.0.0_darwin_amd64.tar.gz | tar -xz
sudo mv multiview-monitor_1.0.0_darwin_amd64/multiview-monitor /usr/local/bin/

# Apple Silicon (M1/M2/M3)
curl -L https://github.com/your-repo/multiview-monitor/releases/latest/download/multiview-monitor_1.0.0_darwin_arm64.tar.gz | tar -xz
sudo mv multiview-monitor_1.0.0_darwin_arm64/multiview-monitor /usr/local/bin/
```

### Windows
1. [Windows 릴리즈](https://github.com/your-repo/multiview-monitor/releases/latest/download/multiview-monitor_1.0.0_windows_amd64.zip) 다운로드
2. 원하는 폴더에 압축 해제
3. `multiview-monitor.exe` 실행

## 상세 설치 가이드

### 1. GitHub Releases에서 다운로드

1. [Releases 페이지](https://github.com/your-repo/multiview-monitor/releases) 방문
2. 최신 버전에서 플랫폼에 맞는 파일 다운로드:
   - `multiview-monitor_1.0.0_linux_amd64.tar.gz` - Linux Intel/AMD
   - `multiview-monitor_1.0.0_linux_arm64.tar.gz` - Linux ARM  
   - `multiview-monitor_1.0.0_darwin_amd64.tar.gz` - macOS Intel
   - `multiview-monitor_1.0.0_darwin_arm64.tar.gz` - macOS Apple Silicon
   - `multiview-monitor_1.0.0_windows_amd64.zip` - Windows

### 2. 압축 해제 및 설치

**Unix 계열 (Linux, macOS):**
```bash
# 압축 해제
tar -xzf multiview-monitor_*.tar.gz
cd multiview-monitor_*

# 자동 설치 스크립트 사용
sudo ./install.sh

# 또는 수동 설치
sudo cp multiview-monitor /usr/local/bin/
sudo chmod +x /usr/local/bin/multiview-monitor

# 설치 확인
multiview-monitor --version
```

**Windows:**
```cmd
# 압축 해제 후 원하는 위치에 배치
# PATH에 추가하거나 직접 실행
multiview-monitor.exe --help
```

### 3. 검증

설치가 완료되면 다음 명령어로 확인:
```bash
multiview-monitor --help
multiview-monitor --generate-config
```

## 제거

### Unix 계열
```bash
sudo rm /usr/local/bin/multiview-monitor
rm -rf ~/.multiview-monitor.yaml  # 설정 파일 (선택사항)
rm -rf monitor.log                # 로그 파일 (선택사항)
```

### Windows
설치 폴더에서 `multiview-monitor.exe` 파일 삭제

## 업그레이드

새 버전으로 업그레이드하려면 동일한 설치 과정을 반복하면 됩니다. 기존 바이너리가 자동으로 교체됩니다.

설정 파일은 하위 호환성을 유지하므로 별도 작업이 필요하지 않습니다.

## 문제 해결

### 권한 에러
```bash
# macOS에서 "cannot be opened because the developer cannot be verified" 에러
sudo spctl --master-disable  # Gatekeeper 임시 비활성화
# 또는
xattr -d com.apple.quarantine multiview-monitor
```

### PATH 설정
```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
export PATH="$PATH:/usr/local/bin"
source ~/.bashrc  # 또는 ~/.zshrc
```

### 실행 권한
```bash
chmod +x multiview-monitor
```

## Docker 사용

Docker 환경에서 실행하려면:
```dockerfile
FROM alpine:latest
COPY multiview-monitor /usr/local/bin/
RUN chmod +x /usr/local/bin/multiview-monitor
ENTRYPOINT ["multiview-monitor"]
```

```bash
docker build -t multiview-monitor .
docker run -it --rm multiview-monitor --help
```