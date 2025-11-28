# My System Updater (나만의 시스템 업데이터)

## 프로젝트 소개
이 프로젝트는 리눅스(Fedora/GNOME 환경) 시스템을 위한 TUI(Terminal User Interface) 기반 시스템 업데이터입니다.
`dnf`, `flatpak`과 같은 시스템 패키지 매니저 업데이트뿐만 아니라, 사용자가 정의한 커스텀 명령어(npm, 스크립트 등)를 통합하여 한 번에 관리하고 실행할 수 있도록 설계되었습니다.

Golang과 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 프레임워크를 사용하여 모던하고 직관적인 터미널 인터페이스를 제공합니다.

## 주요 기능
1.  **통합 업데이트**:
    *   시스템 기본 업데이트: `dnf update`, `flatpak update` (기본 내장)
    *   커스텀 업데이트: 사용자가 추가한 임의의 명령어 실행
2.  **커스텀 메뉴 관리 (설정)**:
    *   TUI 내에서 직접 커스텀 업데이트 명령어를 추가, 수정, 삭제 가능
    *   설정은 `~/.config/my-updater/config.json`에 영구 저장됨
3.  **직관적인 UI**:
    *   진행 상황 바 (Progress Bar) 및 스피너 제공
    *   실시간 로그 출력 확인

## 설치 및 실행 방법

### 요구 사항
*   Linux (Fedora 권장)
*   Go 1.19 이상
*   `sudo` 권한 (시스템 업데이트를 위해 필요)

### 빌드
```bash
go build -o my-updater cmd/my-updater/main.go
```

### 실행
시스템 업데이트 권한이 필요하므로 `sudo`와 함께 실행하거나, 제공된 `.desktop` 파일을 통해 터미널에서 실행해야 합니다.

```bash
sudo ./my-updater
```

## 프로젝트 구조 (Architecture)
"20년차 시니어 엔지니어"의 철학을 담아 견고하고 확장 가능한 구조로 설계되었습니다.

*   `cmd/`: 애플리케이션 진입점
*   `internal/config`: 설정 파일 관리 및 Persistence 계층
*   `internal/domain`: 비즈니스 로직 및 데이터 모델 (Task 정의)
*   `internal/ui`: Bubble Tea 모델 및 뷰 계층 (MVVM 패턴 유사)
    *   `menu`: 메인 메뉴
    *   `updater`: 업데이트 실행 화면
    *   `settings`: 설정 관리 화면

## 라이선스
MIT License
