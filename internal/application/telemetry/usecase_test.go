package telemetry

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domain "telemetry-collector/internal/domain/telemetry"

	"github.com/golang/mock/gomock"
)

func TestProcessUseCaseExecuteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockTelemetryRepository(ctrl)
	uc := NewProcessUseCase(repo)

	in := Input{
		GPUID:          "gpu-1",
		HostID:         "host-1",
		Timestamp:      time.Now().UTC(),
		GPUUtilization: 50,
		MemoryUsedMB:   4096,
		TemperatureC:   60,
	}

	repo.EXPECT().Save(gomock.Any(), gomock.AssignableToTypeOf(domain.Telemetry{})).DoAndReturn(
		func(_ context.Context, got domain.Telemetry) error {
			if got.GPUID != in.GPUID || got.HostID != in.HostID {
				t.Fatalf("unexpected telemetry mapping: %+v", got)
			}
			return nil
		},
	)

	if err := uc.Execute(context.Background(), in); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessUseCaseExecuteValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockTelemetryRepository(ctrl)
	uc := NewProcessUseCase(repo)

	err := uc.Execute(context.Background(), Input{
		GPUID:          "",
		HostID:         "host-1",
		Timestamp:      time.Now().UTC(),
		GPUUtilization: 50,
		MemoryUsedMB:   4096,
		TemperatureC:   60,
	})
	if !domain.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestProcessUseCaseExecuteRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockTelemetryRepository(ctrl)
	uc := NewProcessUseCase(repo)

	repoErr := errors.New("db unavailable")
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(repoErr)

	err := uc.Execute(context.Background(), Input{
		GPUID:          "gpu-1",
		HostID:         "host-1",
		Timestamp:      time.Now().UTC(),
		GPUUtilization: 50,
		MemoryUsedMB:   4096,
		TemperatureC:   60,
	})
	if err == nil {
		t.Fatal("expected wrapped repository error")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
	if !strings.Contains(err.Error(), "save telemetry") {
		t.Fatalf("expected context in error message, got %v", err)
	}
}
