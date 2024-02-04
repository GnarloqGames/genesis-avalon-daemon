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

func (b *BuildTask) Run(ctx context.Context, stop chan struct{}) error {
	timer := time.NewTimer(b.Duration)

	slog.Info("starting building", "name", b.Name, "duration", b.Duration.String())

Listener:
	for {
		select {
		case <-stop:
			return ErrInterrupted
		case <-timer.C:
			break Listener
		}
	}

	slog.Info("building complete", "name", b.Name)
	return nil
}
