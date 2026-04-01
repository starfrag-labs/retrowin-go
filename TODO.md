# TODO

## MEDIUM - 플로우/기능 문제

### M2. POST /user API가 불필요
- **파일**: `internal/handler/user.go`
- **문제**: OIDC 콜백에서 자동 생성됨. 별도 생성 API가 중복이며, provider/providerId를 클라이언트가 임의로 지정 가능
- **해결**: 해당 엔드포인트 제거 또는 관리자 전용으로 제한

### M4. CompleteUpload의 race condition
- **파일**: `internal/handler/upload.go`
- **문제**: 기존 파일 확인 → unlink → delete 사이에 다른 요청이 끼어들 수 있음
- **해결**: 트랜잭션 또는 낙관적 락 도입

### M5. HandleCallback이 쿠키를 설정하지 않음
- **파일**: `internal/handler/auth.go`
- **문제**: 세션 ID를 JSON body로 반환하지만, 브라우저가 자동 전송하려면 쿠키 필요
- **해결**: ogen response hook 또는 미들웨어로 `retrowin_session` 쿠키 설정
