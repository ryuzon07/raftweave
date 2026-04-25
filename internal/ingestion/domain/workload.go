package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

type WorkloadID string

type WorkloadStatus string

const (
	WorkloadStatusPending   WorkloadStatus = "PENDING"
	WorkloadStatusBuilding  WorkloadStatus = "BUILDING"
	WorkloadStatusDeploying WorkloadStatus = "DEPLOYING"
	WorkloadStatusRunning   WorkloadStatus = "RUNNING"
	WorkloadStatusFailing   WorkloadStatus = "FAILING"
	WorkloadStatusFailed    WorkloadStatus = "FAILED"
	WorkloadStatusFailover  WorkloadStatus = "FAILOVER_IN_PROGRESS"
)

type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderAzure CloudProvider = "azure"
	CloudProviderGCP   CloudProvider = "gcp"
)

type Region struct {
	Provider CloudProvider
	Name     string
}

type ResourceSpec struct {
	CPU             string
	Memory          string
	Replicas        int32
	Port            int32
	HealthCheckPath string
	Env             map[string]string
}

type DatabaseSpec struct {
	Engine    string
	Version   string
	StorageGB int32
	Tier      string
}

type FailoverConfig struct {
	RTOSeconds      int32
	RPOSeconds      int32
	AutoFailover    bool
	MinHealthyNodes int32
	FencingEnabled  bool
}

type ComplianceConfig struct {
	DataResidency []DataResidencyRule
}

type DataResidencyRule struct {
	Country string
	Regions []Region
}

type SourceConfig struct {
	Type       string
	RepoURL    string
	Branch     string
	Dockerfile string
	ImageRef   string
}

type Workload struct {
	ID             WorkloadID
	Name           string
	Status         WorkloadStatus
	Source         SourceConfig
	PrimaryRegion  Region
	StandbyRegions []Region
	Compute        ResourceSpec
	Database       DatabaseSpec
	Failover       FailoverConfig
	Compliance     ComplianceConfig
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{2,62}$`)

func NewWorkload(name string, source SourceConfig, primary Region, standbys []Region, compute ResourceSpec, db DatabaseSpec, failover FailoverConfig) (*Workload, error) {
	w := &Workload{
		ID:             WorkloadID(uuid.New().String()),
		Name:           name,
		Status:         WorkloadStatusPending,
		Source:         source,
		PrimaryRegion:  primary,
		StandbyRegions: standbys,
		Compute:        compute,
		Database:       db,
		Failover:       failover,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := w.Validate(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Workload) Validate() error {
	var errs ValidationErrors

	if !nameRegex.MatchString(w.Name) {
		errs = append(errs, ValidationError{Field: "Name", Message: ErrInvalidWorkloadName.Error()})
	}

	if w.Source.Type != "git" && w.Source.Type != "image" {
		errs = append(errs, ValidationError{Field: "Source.Type", Message: ErrInvalidSource.Error()})
	}

	if w.Failover.RTOSeconds < 10 || w.Failover.RTOSeconds > 300 {
		errs = append(errs, ValidationError{Field: "Failover.RTOSeconds", Message: ErrInvalidRTO.Error()})
	}

	if w.Failover.RPOSeconds < 1 || w.Failover.RPOSeconds > 60 {
		errs = append(errs, ValidationError{Field: "Failover.RPOSeconds", Message: ErrInvalidRPO.Error()})
	}

	if len(w.StandbyRegions) == 0 {
		errs = append(errs, ValidationError{Field: "StandbyRegions", Message: "must have at least 1 standby region"})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (w *Workload) IsActive() bool {
	return w.Status == WorkloadStatusRunning
}

func (w *Workload) CanFailover() bool {
	return w.Status == WorkloadStatusRunning || w.Status == WorkloadStatusFailing
}
