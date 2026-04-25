package postgres

const (
	queryInsertWorkload = `
		INSERT INTO workloads (id, name, descriptor_json, status, primary_region, primary_provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	queryFindWorkloadByID = `
		SELECT id, name, descriptor_json, status, primary_region, primary_provider, created_at, updated_at
		FROM workloads WHERE id = $1
	`

	queryFindWorkloadByName = `
		SELECT id, name, descriptor_json, status, primary_region, primary_provider, created_at, updated_at
		FROM workloads WHERE name = $1
	`

	queryUpdateWorkloadStatus = `
		UPDATE workloads SET status = $1, updated_at = $2 WHERE id = $3
	`

	queryListWorkloads = `
		SELECT id, name, descriptor_json, status, primary_region, primary_provider, created_at, updated_at
		FROM workloads ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	queryDeleteWorkload = `DELETE FROM workloads WHERE id = $1`

	queryInsertCredential = `
		INSERT INTO credentials (id, workload_id, provider, credential_type, encrypted_payload, key_version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	queryFindCredentialByWorkloadAndProvider = `
		SELECT id, workload_id, provider, credential_type, encrypted_payload, key_version, created_at, rotated_at
		FROM credentials WHERE workload_id = $1 AND provider = $2
	`

	queryDeleteCredential = `DELETE FROM credentials WHERE id = $1`
)
