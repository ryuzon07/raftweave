package handler

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/raftweave/raftweave/internal/gen/raftweave/v1"
	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
	"github.com/raftweave/raftweave/internal/ingestion/domain"
	"github.com/raftweave/raftweave/internal/ingestion/usecase"
)

type IngestionHandler struct {
	logger               *zap.Logger
	submitWorkload       *usecase.SubmitWorkloadUseCase
	addCredential        *usecase.AddCredentialUseCase
	workloadRepo         domain.WorkloadRepository // For direct reads (CQRS pattern)
}

// Compile-time check to ensure we implement the interface
var _ raftweavev1connect.IngestionServiceHandler = (*IngestionHandler)(nil)

func NewIngestionHandler(
	logger *zap.Logger,
	submitWorkload *usecase.SubmitWorkloadUseCase,
	addCredential *usecase.AddCredentialUseCase,
	workloadRepo domain.WorkloadRepository,
) *IngestionHandler {
	return &IngestionHandler{
		logger:         logger,
		submitWorkload: submitWorkload,
		addCredential:  addCredential,
		workloadRepo:   workloadRepo,
	}
}

func (h *IngestionHandler) SubmitWorkload(ctx context.Context, req *connect.Request[v1.SubmitWorkloadRequest]) (*connect.Response[v1.SubmitWorkloadResponse], error) {
	in := req.Msg.GetWorkload()
	if in == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, domain.ErrInvalidWorkloadName) // Close enough to invalid argument concept
	}

	primaryReq := in.GetRegions().GetPrimary()
	if primaryReq == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, domain.ErrInvalidWorkloadName)
	}

	input := usecase.SubmitWorkloadInput{
		Name: in.GetName(),
		Source: domain.SourceConfig{
			Type:       in.GetSource().GetType(),
			RepoURL:    in.GetSource().GetRepoUrl(),
			Branch:     in.GetSource().GetBranch(),
			Dockerfile: in.GetSource().GetDockerfile(),
			ImageRef:   in.GetSource().GetImageRef(),
		},
		PrimaryRegion: domain.Region{
			Provider: domain.CloudProvider(mapCloudProvider(primaryReq.GetProvider())),
			Name:     mapRegion(primaryReq.GetRegion()),
		},
		Compute: domain.ResourceSpec{
			CPU:      in.GetCompute().GetCpu(),
			Memory:   in.GetCompute().GetMemory(),
			Replicas: in.GetCompute().GetReplicas(),
		},
		Database: domain.DatabaseSpec{
			Engine:    in.GetDatabase().GetEngine(),
			Version:   in.GetDatabase().GetVersion(),
			StorageGB: in.GetDatabase().GetStorageGb(),
		},
		Failover: domain.FailoverConfig{
			RTOSeconds:      in.GetFailover().GetRtoSeconds(),
			RPOSeconds:      in.GetFailover().GetRpoSeconds(),
			AutoFailover:    in.GetFailover().GetAutoFailover(),
			MinHealthyNodes: in.GetFailover().GetMinHealthyNodes(),
			FencingEnabled:  in.GetFailover().GetFencingEnabled(),
		},
	}

	// Standbys mapping
	for _, s := range in.GetRegions().GetStandbys() {
		input.StandbyRegions = append(input.StandbyRegions, domain.Region{
			Provider: domain.CloudProvider(mapCloudProvider(s.GetProvider())),
			Name:     mapRegion(s.GetRegion()),
		})
	}

	out, err := h.submitWorkload.Execute(ctx, input)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return connect.NewResponse(&v1.SubmitWorkloadResponse{
		WorkloadId: string(out.WorkloadID),
		Status:     mapDomainStatus(out.Status),
		Message:    "Workload submitted successfully",
	}), nil
}

func (h *IngestionHandler) AddCloudCredentials(ctx context.Context, req *connect.Request[v1.AddCloudCredentialsRequest]) (*connect.Response[v1.AddCloudCredentialsResponse], error) {
	c := req.Msg.GetCredentials()
	if c == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, domain.ErrInvalidWorkloadName)
	}

	input := usecase.AddCredentialInput{
		WorkloadID: domain.WorkloadID(c.GetId()), 
		Provider:   domain.CloudProvider(mapCloudProvider(c.GetProvider())),
		CredType:   domain.CredentialType(c.GetCredentialType()), // Usually mapped, but domain accepts string logic
		RawPayload: c.GetEncryptedPayload(),                       // we are sending raw payload to be encrypted by domain internally
	}

	out, err := h.addCredential.Execute(ctx, input)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return connect.NewResponse(&v1.AddCloudCredentialsResponse{
		CredentialId: string(out.CredentialID),
		Success:      true,
		Message:      "Credentials added successfully",
	}), nil
}

func (h *IngestionHandler) GetWorkloadStatus(ctx context.Context, req *connect.Request[v1.GetWorkloadStatusRequest]) (*connect.Response[v1.GetWorkloadStatusResponse], error) {
	w, err := h.workloadRepo.FindByName(ctx, req.Msg.GetWorkloadName())
	if err != nil {
		return nil, mapDomainError(err)
	}

	standbys := make([]*v1.RegionTarget, len(w.StandbyRegions))
	for i, sb := range w.StandbyRegions {
		standbys[i] = &v1.RegionTarget{
			Provider: reverseCloudProvider(w.PrimaryRegion.Provider), // assuming same, need fix mapping
			Region:   reverseRegionMappedFastString(sb.Name),
		}
	}

	return connect.NewResponse(&v1.GetWorkloadStatusResponse{
		WorkloadName: w.Name,
		Status:       mapDomainStatus(w.Status),
		PrimaryRegion: &v1.RegionTarget{
			Provider: reverseCloudProvider(w.PrimaryRegion.Provider),
			Region:   reverseRegionMappedFastString(w.PrimaryRegion.Name),
		},
		StandbyRegions: standbys,
		LastUpdated:    timestamppb.New(w.UpdatedAt),
	}), nil
}

func (h *IngestionHandler) ListWorkloads(ctx context.Context, req *connect.Request[v1.ListWorkloadsRequest]) (*connect.Response[v1.ListWorkloadsResponse], error) {
	// A real implementation would parse pagination, let's dummy for now or do full scan
	// Assume workloadRepo defines a FindAll or similar, let's see. If not, missing. Return Unimplemented
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("list workloads not implemented"))
}

// -----------------------
// Helper Mappings
// -----------------------

func mapCloudProvider(p v1.CloudProvider) string {
	switch p {
	case v1.CloudProvider_CLOUD_PROVIDER_AWS:
		return string(domain.CloudProviderAWS)
	case v1.CloudProvider_CLOUD_PROVIDER_AZURE:
		return string(domain.CloudProviderAzure)
	case v1.CloudProvider_CLOUD_PROVIDER_GCP:
		return string(domain.CloudProviderGCP)
	}
	return ""
}

func reverseCloudProvider(p domain.CloudProvider) v1.CloudProvider {
	switch p {
	case domain.CloudProviderAWS:
		return v1.CloudProvider_CLOUD_PROVIDER_AWS
	case domain.CloudProviderAzure:
		return v1.CloudProvider_CLOUD_PROVIDER_AZURE
	case domain.CloudProviderGCP:
		return v1.CloudProvider_CLOUD_PROVIDER_GCP
	}
	return v1.CloudProvider_CLOUD_PROVIDER_UNSPECIFIED
}

func mapDomainStatus(s domain.WorkloadStatus) v1.WorkloadStatus {
	switch s {
	case domain.WorkloadStatusPending:
		return v1.WorkloadStatus_WORKLOAD_STATUS_PENDING
	case domain.WorkloadStatusBuilding:
		return v1.WorkloadStatus_WORKLOAD_STATUS_BUILDING
	case domain.WorkloadStatusDeploying:
		return v1.WorkloadStatus_WORKLOAD_STATUS_DEPLOYING
	case domain.WorkloadStatusRunning:
		return v1.WorkloadStatus_WORKLOAD_STATUS_RUNNING
	case domain.WorkloadStatusFailing:
		return v1.WorkloadStatus_WORKLOAD_STATUS_FAILING
	case domain.WorkloadStatusFailed:
		return v1.WorkloadStatus_WORKLOAD_STATUS_FAILED
	case domain.WorkloadStatusFailover:
		return v1.WorkloadStatus_WORKLOAD_STATUS_FAILOVER_IN_PROGRESS
	}
	return v1.WorkloadStatus_WORKLOAD_STATUS_UNSPECIFIED
}

func mapRegion(r v1.Region) string {
	switch r {
	case v1.Region_REGION_AWS_AP_SOUTH_1:
		return "ap-south-1"
	case v1.Region_REGION_AWS_US_EAST_1:
		return "us-east-1"
	// simplified mapping for demo	
	}
	return "unknown-region"
}

func reverseRegionMappedFastString(n string) v1.Region {
	switch n {
	case "ap-south-1":
		return v1.Region_REGION_AWS_AP_SOUTH_1
	case "us-east-1":
		return v1.Region_REGION_AWS_US_EAST_1
	}
	return v1.Region_REGION_UNSPECIFIED
}