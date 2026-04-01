# TODO

## CRITICAL - 기능적 결함

### C1. `domainError()` 미구현
- **파일**: `internal/handler/handler.go:77-79`
- **문제**: `domainError()`가 에러를 그대로 반환하기만 함. `errors.Error`의 `StatusCode`를 HTTP 상태 코드로 변환하지 않음
- **영향**: 모든 도메인 에러가 500으로 반환될 가능성. ogen의 `ErrorHandler`가 처리하는지 확인 필요
- **해결**: `errors.Error` 타입 체크 후 적절한 HTTP 상태 코드 매핑, 또는 ogen의 convenient_errors 활용

### C2. `GetDownloadUrl` ExpiresAt 잘못됨
- **파일**: `internal/handler/upload.go:122`
- **문제**: `toTimestamp(in.Mtime())` 사용 — inode 수정 시간이 아닌 실제 presigned URL 만료 시간을 반환해야 함
- **해결**: `objectSvc.GetDownloadURL`에서 URL과 만료 시간을 함께 반환하도록 변경

### C3. `gc.go` 파일 손상
- **파일**: `internal/application/storage/gc.go`
- **문제**: GarbageCollector 코드 이후에 storage.go 코드가 섞여 들어감
- **해결**: gc.go에서 storage.go 내용 제거, 각 파일의 책임 분리

### C4. `HandleCallback` 에러 처리 단순화
- **파일**: `internal/handler/auth.go:39-41`
- **문제**: 모든 콜백 에러를 `HandleCallbackUnauthorized`로 반환. state 불일치, 코드 만료 등 구분 안 됨
- **해결**: 에러 타입별로 적절한 응답 반환 (BadRequest, Unauthorized 등)

---

## HIGH - 보안/설계 문제

### H1. 시스템 접근 인가(Authorization) 없음
- **파일**: `internal/handler/system.go`, `internal/handler/fs.go`, `internal/handler/upload.go`
- **문제**: 인증은 되어 있으나 인가가 없음. 모든 인증된 사용자가 모든 시스템의 파일/데이터에 접근 가능
- **영향**: 사용자 A가 사용자 B의 시스템 파일 읽기/수정/삭제 가능
- **해결**: 시스템 접근 시 해당 사용자가 시스템 멤버인지 확인하는 인가 미들웨어/체크 추가

### H2. 핸들러 간 에러 처리 불일치
- **파일**: `internal/handler/system_user.go`, `internal/handler/system_group.go`
- **문제**: 일부 핸들러는 `h.domainError(err)`를 사용하고, 다른 핸들러는 `return nil, err`를 직접 사용
- **해결**: 모든 핸들러에서 일관된 에러 처리 방식 사용

### H3. UID/GID 기반 사용자 조회 비효율
- **파일**: `internal/handler/system_user.go:52-67`, `internal/handler/system_group.go:84-99`
- **문제**: 전체 사용자를 `Find`로 가져온 후 루프로 UID 매칭. DB 레벨 필터링 아님
- **해결**: `FindBySystemAndUID(systemID, uid)` 같은 리포지토리 메서드 추가

---

## MEDIUM - 플로우/기능 문제

### M1. 로그아웃 시 쿠키 미삭제 (가장 시급)
- **파일**: `internal/cmd/retrowin-server/server.go:190-208`
- **문제**: `logoutMiddleware`가 `/auth/logout` 경로에서만 동작. ogen이 `/auth/logout` 대신 다른 경로 패턴을 사용할 수 있음. 실제로 쿠키가 삭제되는지 확인 필요
- **해결**: ogen 라우팅 경로와 middleware 경로 일치 확인. 또는 핸들러에서 직접 쿠키 설정

### M2. POST /user API가 불필요
- **파일**: `internal/handler/user.go`
- **문제**: OIDC 콜백에서 `auth.UserService.FindOrCreate`로 자동 생성됨. 별도 생성 API가 중복이며, provider/providerId를 클라이언트가 임의로 지정 가능 = 보안 위험
- **해결**: 해당 엔드포인트 제거 또는 관리자 전용으로 제한

### M3. 파일 삭제 시 S3 오브젝트 미삭제
- **파일**: `internal/handler/fs.go:178-206` (Unlink), `internal/handler/upload.go:85` (CompleteUpload 덮어쓰기)
- **문제**: inode와 디렉토리 엔트리는 삭제하지만, S3에 실제 업로드된 오브젝트는 삭제하지 않음
- **해결**: inode의 `ObjectContent`를 파싱하여 해당 object ID로 `objectSvc.Delete` 호출

### M4. CompleteUpload의 race condition
- **파일**: `internal/handler/upload.go:78-88`
- **문제**: 기존 파일 확인 → unlink → delete 사이에 다른 요청이 끼어들 수 있음
- **해결**: 트랜잭션 또는 낙관적 락 도입

### M5. ListSystems에 필터링 없음
- **파일**: `internal/handler/system.go:41-45`
- **문제**: `Find(ctx, system.Filter{})` — 빈 필터로 전체 시스템 반환. 사용자의 시스템만 반환해야 함
- **해결**: 사용자 ID 기반 필터링 추가

### M6. `HandleCallback`이 쿠키를 설정하지 않음
- **파일**: `internal/handler/auth.go:43-48`
- **문제**: 세션 ID를 JSON body로 반환하지만, 보통 쿠키로 설정해야 브라우저가 자동으로 전송함
- **영향**: SPA/브라우저 클라이언트가 쿠키를 수동으로 설정해야 함
- **해결**: `http.Cookie` 헤더로 `retrowin_session` 설정

---

## LOW - 개선 사항

### L1. `InitiateUpload` size 음수 검증 없음
- **파일**: `internal/handler/upload.go`
- **해결**: `req.Size > 0` 검증 추가

### L2. symlink target 검증 없음
- **파일**: `internal/handler/fs.go`
- **해결**: target path 길이 및 포맷 검증

### L3. chmod mode 값 범위 검증 없음
- **파일**: `internal/handler/fs.go:159`
- **해결**: permission bits (0-0o777) 범위 검증. file type bits 포함 시 거부

### L4. GC(Garbage Collector) 스케줄링 없음
- **파일**: `internal/application/storage/gc.go`
- **문제**: `GarbageCollector` 구현은 있지만 FX DI에 등록되지 않았고, 주기적 실행 로직 없음
- **해결**: FX lifecycle에 주기적 GC 실행 등록

### L5. 덮어쓰기 시 기존 S3 오브젝트 삭제 누락
- **파일**: `internal/handler/upload.go:84-87`
- **문제**: 기존 inode는 삭제하지만 연결된 S3 오브젝트는 삭제하지 않음 (M3과 관련)
