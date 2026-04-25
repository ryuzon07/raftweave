# RaftWeave — Kubernetes Secrets Management

> **⚠️ NEVER store secrets in this repository.**

RaftWeave uses the [External Secrets Operator (ESO)](https://external-secrets.io/) to synchronize secrets from cloud-native secret managers into Kubernetes Secrets at runtime.

## Required Kubernetes Secrets

These secrets **must exist** in the `raftweave` namespace before deploying. They are created automatically by ESO from the configured `SecretStore`.

| Secret Name | Keys | Description |
|---|---|---|
| `raftweave-postgres` | `host`, `port`, `user`, `password`, `dbname` | PostgreSQL connection credentials |
| `raftweave-redis` | `host`, `port`, `password` | Redis connection credentials |
| `raftweave-crypto` | `aes-key-v1`, `aes-key-v2` | AES-256-GCM encryption keys (v2 optional, for rotation) |
| `raftweave-oauth` | `github-client-id`, `github-client-secret`, `google-client-id`, `google-client-secret` | OAuth provider credentials |
| `raftweave-cloud-aws` | `access-key-id`, `secret-access-key` | AWS IAM credentials for provisioning |
| `raftweave-cloud-azure` | `client-id`, `client-secret`, `tenant-id` | Azure Service Principal credentials |
| `raftweave-cloud-gcp` | `service-account-json` | GCP Service Account key JSON |

## Cloud Provider Secret Backends

| Environment | Backend | Region |
|---|---|---|
| **Staging** | AWS Secrets Manager | `ap-south-1` |
| **Production** | AWS Secrets Manager (primary) + Azure Key Vault (failover) | Multi-region |

## Key Rotation

Encryption keys support versioned rotation:

1. Add new key as `aes-key-v2` in the secret manager
2. Update the `ENCRYPTION_CURRENT_KEY_VERSION` config to `v2`
3. New encryptions use `v2`, old data can still be decrypted with `v1`
4. Run the rotation job to re-encrypt all credentials with `v2`
5. Remove `aes-key-v1` after confirming all data is re-encrypted

## Local Development

For local development with Docker Compose, secrets are sourced from `.env` files. See `deploy/.env.example`.

## Setting Up ESO

```bash
# Install External Secrets Operator
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets --create-namespace

# Create a ClusterSecretStore pointing to AWS Secrets Manager
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: ap-south-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets
            namespace: external-secrets
EOF
```
