# Secret Rotation Procedures — RaftWeave Auth

## JWT Key Rotation

### When to Rotate
- Routine: Every 90 days
- Emergency: Immediately if private key compromise is suspected

### Steps

1. **Generate new RSA-4096 key pair:**
   ```bash
   openssl genpkey -algorithm RSA -out new_jwt_private.pem -pkeyopt rsa_keygen_bits:4096
   openssl rsa -in new_jwt_private.pem -pubout -out new_jwt_public.pem
   ```

2. **Update Kubernetes Secret with new private key:**
   ```bash
   kubectl create secret generic raftweave-auth-secrets \
     --from-file=jwt_private_key_pem=new_jwt_private.pem \
     --from-literal=encryption_key="$(cat encryption.key)" \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

3. **Keep old public key in JWKS endpoint for 15 minutes** (token TTL grace period):
   - The auth service should serve both old and new public keys at `/.well-known/jwks.json`
   - Downstream services will validate tokens with either key

4. **Rolling restart auth deployment:**
   ```bash
   kubectl rollout restart deployment/raftweave-auth -n raftweave-production
   kubectl rollout status deployment/raftweave-auth -n raftweave-production
   ```

5. **Remove old public key after 15 minutes:**
   - All tokens signed with the old key will have expired
   - Remove the old public key from the JWKS endpoint

6. **Verify:**
   ```bash
   # Complete a full auth flow
   curl -s https://api.raftweave.io/auth.v1.AuthService/GetMe \
     -H "Authorization: Bearer <new-token>" | jq .
   ```

---

## Encryption Key Rotation (AES-256-GCM)

### When to Rotate
- Routine: Every 180 days
- Emergency: Immediately if key compromise is suspected

### Steps

1. **Generate new 32-byte AES key:**
   ```bash
   openssl rand -hex 32 > new_encryption.key
   ```

2. **Run migration job to re-encrypt all `github_token_enc` values:**
   ```bash
   kubectl apply -f deploy/kubernetes/auth/jobs/reencrypt-tokens.yaml
   # The job reads old key, decrypts, re-encrypts with new key, updates DB
   kubectl wait --for=condition=complete job/reencrypt-tokens -n raftweave-production
   ```

3. **Update Kubernetes Secret:**
   ```bash
   kubectl create secret generic raftweave-auth-secrets \
     --from-file=jwt_private_key_pem=jwt_private.pem \
     --from-literal=encryption_key="$(cat new_encryption.key)" \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

4. **Rolling restart auth deployment:**
   ```bash
   kubectl rollout restart deployment/raftweave-auth -n raftweave-production
   ```

5. **Verify by listing repos:**
   ```bash
   curl -s https://api.raftweave.io/auth.v1.AuthService/ListUserRepos \
     -H "Authorization: Bearer <token>" | jq '.repos | length'
   ```

---

## OAuth Secret Rotation (GitHub / Google)

### When to Rotate
- Routine: Every 365 days
- Emergency: Immediately if client secret is leaked

### Steps

1. **Generate new client secret in the provider's developer console:**
   - GitHub: Settings → Developer settings → OAuth Apps → Generate new client secret
   - Google: Google Cloud Console → APIs & Services → Credentials → Reset secret

2. **Update Kubernetes Secret:**
   ```bash
   kubectl create secret generic raftweave-oauth-secrets \
     --from-literal=github_client_id="<id>" \
     --from-literal=github_client_secret="<new-secret>" \
     --from-literal=google_client_id="<id>" \
     --from-literal=google_client_secret="<new-secret>" \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

3. **Rolling restart (no migration needed — OAuth tokens are short-lived):**
   ```bash
   kubectl rollout restart deployment/raftweave-auth -n raftweave-production
   ```

4. **Verify OAuth flows:**
   - Complete a GitHub login → confirm token received
   - Complete a Google login → confirm token received

---

## SMTP Credential Rotation

### Steps

1. **Update credentials in your email provider dashboard.**

2. **Update Kubernetes Secret:**
   ```bash
   kubectl create secret generic raftweave-smtp \
     --from-literal=host="<smtp-host>" \
     --from-literal=port="587" \
     --from-literal=username="<new-user>" \
     --from-literal=password="<new-pass>" \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

3. **Rolling restart:**
   ```bash
   kubectl rollout restart deployment/raftweave-auth -n raftweave-production
   ```

4. **Verify by requesting an OTP:**
   ```bash
   grpcurl -d '{"email":"test@example.com"}' \
     api.raftweave.io:443 auth.v1.AuthService/RequestOTP
   ```

---

## Emergency Response

If any secret is compromised:

1. **Rotate the compromised secret immediately** using the steps above
2. **If JWT private key:** Revoke ALL active sessions (`RevokeAll` for each affected user)
3. **If encryption key:** Re-encrypt all tokens, then revoke GitHub tokens for affected users
4. **Audit logs:** Check CloudWatch/Datadog for unauthorized access during exposure window
5. **Notify affected users** via security alert email
6. **Post-incident review:** Document timeline, impact, and prevention measures
