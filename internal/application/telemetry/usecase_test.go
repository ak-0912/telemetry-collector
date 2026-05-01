package telemetry

import (
	"context"
	"errors"
	"strings"
	"testing"

	domain "telemetry-collector/internal/domain/telemetry"

	"github.com/golang/mock/gomock"
)

func TestProcessUseCaseExecuteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockTelemetryRepository(ctrl)
	uc := NewProcessUseCase(repo)

	in := Input{
		MetricName:          "gpu.temperature",
		GPUID:               "gpu-1",
		Device:              "nvidia0",
		UUID:                "d083db3f-88d3-4714-bcff-e0a4e95d709f",
		ModelName:           "A100",
		HostName:            "host-1",
		Value:               60,
		LabelsRaw:           "{}",
		ProcessedAtUnixNano: 1735689600000000000,
	}

	repo.EXPECT().Save(gomock.Any(), gomock.AssignableToTypeOf(domain.Telemetry{})).DoAndReturn(
		func(_ context.Context, got domain.Telemetry) error {
			if got.GPUID != in.GPUID || got.HostName != in.HostName || got.UUID != in.UUID {
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
		MetricName:          "gpu.temperature",
		GPUID:               "",
		Device:              "nvidia0",
		UUID:                "d083db3f-88d3-4714-bcff-e0a4e95d709f",
		ModelName:           "A100",
		HostName:            "host-1",
		Value:               60,
		LabelsRaw:           "{}",
		ProcessedAtUnixNano: 1735689600000000000,
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
		MetricName:          "gpu.temperature",
		GPUID:               "gpu-1",
		Device:              "nvidia0",
		UUID:                "d083db3f-88d3-4714-bcff-e0a4e95d709f",
		ModelName:           "A100",
		HostName:            "host-1",
		Value:               60,
		LabelsRaw:           "{}",
		ProcessedAtUnixNano: 1735689600000000000,
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
