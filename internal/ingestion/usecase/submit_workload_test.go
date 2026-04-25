package usecase_test

import (
	"context"
	"testing"

	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raftweave/raftweave/internal/ingestion"
	"github.com/raftweave/raftweave/internal/ingestion/domain"
	"github.com/raftweave/raftweave/internal/ingestion/usecase"
)

type mockWorkloadRepo struct { mock.Mock }
func (m *mockWorkloadRepo) Save(ctx context.Context, w *domain.Workload) error { return m.Called(ctx, w).Error(0) }
func (m *mockWorkloadRepo) FindByID(ctx context.Context, id domain.WorkloadID) (*domain.Workload, error) {
	args := m.Called(ctx, id)
	if w := args.Get(0); w != nil {
		return w.(*domain.Workload), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockWorkloadRepo) FindByName(ctx context.Context, name string) (*domain.Workload, error) {
	args := m.Called(ctx, name)
	if w := args.Get(0); w != nil {
		return w.(*domain.Workload), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockWorkloadRepo) FindAll(ctx context.Context, limit, offset int) ([]*domain.Workload, error) { return nil, nil }
func (m *mockWorkloadRepo) UpdateStatus(ctx context.Context, id domain.WorkloadID, status domain.WorkloadStatus) error { return nil }
func (m *mockWorkloadRepo) Delete(ctx context.Context, id domain.WorkloadID) error { return m.Called(ctx, id).Error(0) }

type mockJobEnqueuer struct { mock.Mock }
func (m *mockJobEnqueuer) EnqueueBuildJob(ctx context.Context, id domain.WorkloadID, event *domain.WebhookEvent) error { return nil }
func (m *mockJobEnqueuer) EnqueueProvisionJob(ctx context.Context, id domain.WorkloadID) error { return m.Called(ctx, id).Error(0) }

// Simple mock tests for SubmitWorkload
func TestSubmitWorkload_Success(t *testing.T) {
	repo := new(mockWorkloadRepo)
	enq := new(mockJobEnqueuer)
	logger, _ := zap.NewDevelopment()
	meter := metricnoop.MeterProvider{}.Meter("test")
	deps := usecase.Dependencies{
		WorkloadRepo:   repo,
		JobEnqueuer:    enq,
		Tracer:         tracenoop.Tracer{},
		Logger:         logger,
		Metrics:        ingestion.NewIngestionMetrics(meter),
	}

	repo.On("FindByName", mock.Anything, "new-workload").Return((*domain.Workload)(nil), domain.ErrWorkloadNotFound)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Workload")).Return(nil)
	enq.On("EnqueueProvisionJob", mock.Anything, mock.AnythingOfType("domain.WorkloadID")).Return(nil)

	uc := usecase.NewSubmitWorkloadUseCase(deps)

	input := usecase.SubmitWorkloadInput{
		Name: "new-workload",
		Source: domain.SourceConfig{Type: "git", RepoURL: "https://git.com", Branch: "main"},
		PrimaryRegion: domain.Region{Name: "us-east-1", Provider: "aws"},
		StandbyRegions: []domain.Region{{Name: "us-west-2", Provider: "aws"}},
		Failover: domain.FailoverConfig{RTOSeconds: 30, RPOSeconds: 15},
	}

	out, err := uc.Execute(context.Background(), input)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, domain.WorkloadStatusPending, out.Status)
}

func TestSubmitWorkload_DuplicateName(t *testing.T) {
	repo := new(mockWorkloadRepo)
	logger, _ := zap.NewDevelopment()
	meter := metricnoop.MeterProvider{}.Meter("test")
	deps := usecase.Dependencies{
		WorkloadRepo: repo,
		Tracer:       tracenoop.Tracer{},
		Logger:       logger,
		Metrics:      ingestion.NewIngestionMetrics(meter),
	}

	existing := &domain.Workload{Name: "duplicate-name"}
	repo.On("FindByName", mock.Anything, "duplicate-name").Return(existing, nil)

	uc := usecase.NewSubmitWorkloadUseCase(deps)

	input := usecase.SubmitWorkloadInput{
		Name: "duplicate-name",
		Source: domain.SourceConfig{Type: "git", RepoURL: "https://git.com", Branch: "main"},
		PrimaryRegion: domain.Region{Name: "us-east-1", Provider: "aws"},
		StandbyRegions: []domain.Region{{Name: "us-west-2", Provider: "aws"}},
		Failover: domain.FailoverConfig{RTOSeconds: 30, RPOSeconds: 15},
	}
	out, err := uc.Execute(context.Background(), input)
	assert.ErrorIs(t, err, domain.ErrWorkloadAlreadyExists)
	assert.Nil(t, out)
}
