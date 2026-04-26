# RaftWeave Deployment Guide — Authentication & Connectivity

This guide covers the necessary steps to resolve deployment errors and ensure the frontend connects successfully to the authentication service.

## 1. Token Storage Strategy

**No manual storage is required.** RaftWeave is configured with a secure, cookie-based authentication system:

- **Cookies:** The backend sets `HttpOnly`, `Secure`, and `SameSite=Lax` cookies (`raftweave_at` for access and `raftweave_rt` for refresh).
- **Security:** `HttpOnly` prevents JavaScript (and thus XSS attacks) from accessing the tokens.
- **Frontend:** The Angular `AuthInterceptor` is configured with `withCredentials: true`, which automatically includes these cookies in all API requests.

## 2. Resolving `CreateContainerConfigError`

The pods are failing because the required Kubernetes Secrets are missing. You must create them manually (or via your secret manager) in the `raftweave` namespace.

### Step 1: Create Namespace
```bash
kubectl create namespace raftweave
```

### Step 2: Create Core Secrets
Replace the placeholders with your actual values.

```bash
# Database Credentials
kubectl create secret generic raftweave-postgres \
  -n raftweave \
  --from-literal=host=raftweave-postgres \
  --from-literal=port=5432 \
  --from-literal=user=raftweave \
  --from-literal=password=REPLACE_WITH_DB_PASSWORD \
  --from-literal=dbname=raftweave

# Redis Credentials
kubectl create secret generic raftweave-redis \
  -n raftweave \
  --from-literal=host=raftweave-redis \
  --from-literal=port=6379 \
  --from-literal=password=REPLACE_WITH_REDIS_PASSWORD

# Encryption & JWT Keys
# Generate a private key: openssl genrsa -out jwt.pem 2048
kubectl create secret generic raftweave-auth-secrets \
  -n raftweave \
  --from-literal=jwt_private_key_pem="$(cat jwt.pem)" \
  --from-literal=encryption_key=$(openssl rand -hex 32)

# OAuth Credentials (from GitHub/Google Developer Consoles)
kubectl create secret generic raftweave-oauth-secrets \
  -n raftweave \
  --from-literal=github_client_id=YOUR_GITHUB_ID \
  --from-literal=github_client_secret=YOUR_GITHUB_SECRET \
  --from-literal=google_client_id=YOUR_GOOGLE_ID \
  --from-literal=google_client_secret=YOUR_GOOGLE_SECRET

# Crypto Keys for Ingestion
kubectl create secret generic raftweave-crypto \
  -n raftweave \
  --from-literal=aes-key-v1=$(openssl rand -hex 32)
```

## 3. Apply the Updated Manifests

I have updated the Ingress and Deployment manifests to fix routing and integration. Apply them using:

```bash
kubectl apply -k deploy/kubernetes/base
```

## 4. Verification

1. **Check Pod Status:**
   ```bash
   kubectl get pods -n raftweave
   ```
   All pods should transition to `Running`.

2. **Check Frontend Connectivity:**
   - Open the dashboard in your browser.
   - The "Authentication Service" status should now show **Connected**.
   - You should be able to click "Login" and be redirected to GitHub/Google.

## 5. OAuth Redirect Configuration

Ensure your OAuth application (GitHub/Google) has the following callback URL configured:
`https://dashboard.raftweave.io/auth/github/callback` (or google)
