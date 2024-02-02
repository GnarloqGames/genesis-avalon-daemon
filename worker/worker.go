package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type System struct {
	processes *ProcessStore
	inbox     chan Task
	stop      chan struct{}
}

type Task interface {
	GetID() uuid.UUID
	Run(ctx context.Context, signals chan signal) error
}

type BuildTask struct {
	ID       uuid.UUID
	Name     string
	Duration string
}

func (b *BuildTask) GetID() uuid.UUID {
	return b.ID
}

func (b *BuildTask) Run(ctx context.Context, signals chan signal) error {
	slog.Warn("hello")
	return nil
}

type signal interface {
	status() Status
	task() uuid.UUID
	timestamp() time.Time
}

type start struct {
	now time.Time
}

func (s start) status() Status {
	return StatusInProgress
}

func NewSystem() *System {
	system := &System{
		processes: &ProcessStore{
			mx:    &sync.Mutex{},
			store: make(map[uuid.UUID]*Process),
		},
		inbox: make(chan Task),
		stop:  make(chan struct{}),
	}

	go func() {
		for {
			select {
			case <-system.stop:
				slog.Info("stopping worker system")
				return
			case task := <-system.inbox:
				slog.Info("received task", "id", task.GetID())
				process := newProcess(task)
				process.Start()
			}
		}
	}()

	return system
}

func (s *System) Inbox() chan Task {
	return s.inbox
}

func (s *System) Stop() chan struct{} {
	return s.stop
}
