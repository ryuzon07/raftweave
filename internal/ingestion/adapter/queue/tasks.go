package queue

const (
	TaskTypeBuild     = "build:image"
	TaskTypeProvision = "provisioner:deploy"
)

type BuildJobPayload struct {
	WorkloadID   string `json:"workload_id"`
	WorkloadName string `json:"workload_name"`
	RepoURL      string `json:"repo_url"`
	Branch       string `json:"branch"`
	CommitSHA    string `json:"commit_sha"`
	CommitMsg    string `json:"commit_msg"`
	TriggeredAt  string `json:"triggered_at"`
}

type ProvisionJobPayload struct {
	WorkloadID   string `json:"workload_id"`
	WorkloadName string `json:"workload_name"`
	Action       string `json:"action"` // PROVISION | DEPROVISION | SCALE_UP
	TriggeredAt  string `json:"triggered_at"`
}
