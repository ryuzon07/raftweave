# Testing RaftWeave Authentication System

This guide covers how to verify and test the RaftWeave authentication and authorization system.

## 1. Automated Tests

### Unit & Integration Tests
Run all tests in the auth system:
```bash
go test -v ./internal/auth/... ./internal/middleware/...
```

> [!NOTE]
> Postgres and Redis tests require Docker. If you are on Windows and encounter `rootless Docker is not supported` errors, these tests are best run in a Linux-based CI environment or WSL2.

### Coverage Report
Generate a coverage report to identify untested paths:
```bash
go test -coverprofile=coverage.out ./internal/auth/...
go tool cover -html=coverage.out
```

---

## 2. Manual RPC Testing (grpcurl)

If the server is running locally (default `:8080`), you can use `grpcurl` to interact with the AuthService.

### Request an Email OTP
```bash
grpcurl -plaintext -d '{"email": "user@example.com"}' \
  localhost:8080 auth.v1.AuthService/RequestOTP
```
*Output: Returns a `challenge_id`.*

### Verify OTP & Get Tokens
Replace `<challenge_id>` and `<code>` (check logs if using `NoopMailer`):
```bash
grpcurl -plaintext -d '{"challenge_id": "<id>", "code": "123456"}' \
  localhost:8080 auth.v1.AuthService/VerifyOTP
```
*Output: Returns `access_token` and `refresh_token`.*

### Test Authenticated Request (GetMe)
```bash
grpcurl -plaintext -H "Authorization: Bearer <access_token>" \
  localhost:8080 auth.v1.AuthService/GetMe
```

### Refresh Session
```bash
grpcurl -plaintext -d '{"refresh_token": "<refresh_token>"}' \
  localhost:8080 auth.v1.AuthService/RefreshToken
```

---

## 3. OAuth Flow Testing

Since OAuth requires a browser redirect, follow these steps:

1. **Start the server** with valid `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET`.
2. **Navigate to the Login URL** in your browser:
   `http://localhost:8080/auth/github/login`
3. **Complete the GitHub Auth** on their site.
4. **Capture the Token**: The server will redirect you back to the dashboard (or a callback URL) with the access token set in a `Secure; HttpOnly` cookie named `raftweave_at`.

---

## 4. Security & Robustness Testing

### Algorithm Confusion Attack
Try to validate a token signed with `HS256` (Symmetric) instead of `RS256` (Asymmetric). The system should reject it:
```bash
# This is covered by internal/auth/adapter/jwt/issuer_test.go
go test -v ./internal/auth/adapter/jwt -run TestValidate_AlgorithmConfusion
```

### Rate Limiting
Attempt to request an OTP 10 times in 1 minute for the same email. You should receive a `CodeResourceExhausted` (429) error.

### Session Hijacking (Fingerprint Mismatch)
1. Login on one device (or use a specific User-Agent).
2. Attempt to use the `refresh_token` with a different `User-Agent` or from a different IP.
3. The system should revoke **all** user sessions and trigger a security alert.

### RBAC Enforcement
1. Create a token for a user with `VIEWER` role in a workspace.
2. Attempt to call an admin-only endpoint (e.g., a hypothetical `DeleteWorkspace`) with that token.
3. The system should return `CodePermissionDenied` (403).

---

## 5. Load Testing (Optional)
Use `ghz` for load testing the auth endpoints:
```bash
ghz --insecure --proto ./api/proto/auth/v1/auth.proto \
    --call auth.v1.AuthService/GetMe \
    -H "Authorization: Bearer <token>" \
    localhost:8080
```
