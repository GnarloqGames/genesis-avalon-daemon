package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/GnarloqGames/genesis-avalon-kit/database/couchbase"
	"github.com/GnarloqGames/genesis-avalon-kit/registry"
	"github.com/google/uuid"
)

type BuildTask struct {
	ID       uuid.UUID
	Name     string
	Duration time.Duration
	Owner    string
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

func (b *BuildTask) UpdateDB(status Status) error {
	db, err := couchbase.Get()
	if err != nil {
		return err
	}

	item := registry.Building{
		ID:     b.ID.String(),
		Owner:  b.Owner,
		Name:   b.Name,
		Status: status.String(),
	}

	return db.Upsert(item)
}

func (b *BuildTask) Run(ctx context.Context) error {
	slog.Info("building complete", "name", b.Name)

	return nil
}
