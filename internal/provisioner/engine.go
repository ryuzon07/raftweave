package provisioner

import "context"

// TranslationEngine translates cloud-agnostic workload specs into
// provider-specific resource definitions.
type TranslationEngine interface {
	Translate(ctx context.Context, workloadName string, spec *ResourceSpec, provider string, region string) (*ProviderResourceDef, error)
}

// ProviderResourceDef is the provider-specific resource definition.
type ProviderResourceDef struct {
	Provider    string
	Region      string
	ComputeDef  interface{}
	DatabaseDef interface{}
}
