package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type BuildTask struct {
	ID       uuid.UUID
	Name     string
	Duration time.Duration
}

func (b *BuildTask) GetID() uuid.UUID {
	return b.ID
}

func (b *BuildTask) GetDuration() time.Duration {
	return b.Duration
}

func (b *BuildTask) GetName() string {
	return b.Name
}

func (b *BuildTask) Run(ctx context.Context) error {
	slog.Info("building complete", "name", b.Name)
	return nil
}