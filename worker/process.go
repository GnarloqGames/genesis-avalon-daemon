package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Process struct {
	ID        uuid.UUID
	Status    Status
	Task      Task
	Remaining time.Duration
	stop      chan struct{}
}

func newProcess(task Task) *Process {
	process := &Process{
		ID:   task.GetID(),
		Task: task,
		stop: make(chan struct{}),
	}

	return process
}

func (p *Process) SetStatus(status Status) {
	slog.Info("process status changed",
		"process_id", p.ID,
		"old_status", p.Status,
		"new_status", status,
		"task_id", p.Task.GetID(),
	)
	p.Status = status
}

func (p *Process) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	// Only start a task if it hasn't been completed yet
	if p.Status != StatusPending {
		return
	}

	originalDeadline := p.calculateDeadline()

	p.SetStatus(StatusInProgress)
	if err := p.Task.Run(context.Background(), p.stop); err != nil {
		if err == ErrInterrupted {
			p.SetStatus(StatusInterrupted)

			remaining := time.Until(originalDeadline)
			if remaining < 0 {
				slog.Warn("task was interrupted but remaining time less than 0",
					"task_id", p.ID,
					"status", p.Status.String(),
					"remaining", remaining.String(),
				)
				return
			}

			slog.Info("calculated remaining time", "status", p.Status.String(), "remaining", remaining)
			p.Remaining = remaining

			return
		} else {
			p.SetStatus(StatusFailed)
			return
		}
	} else {
		p.SetStatus(StatusDone)
		return
	}
}

func (p *Process) Stop() {
	// Only send stop signals to tasks that are still running
	if p.Status == StatusInProgress {
		p.stop <- struct{}{}
	}
}

type ProcessStore struct {
	mx    *sync.Mutex
	store map[uuid.UUID]*Process
}

func NewStore() *ProcessStore {
	return &ProcessStore{
		mx:    &sync.Mutex{},
		store: make(map[uuid.UUID]*Process),
	}
}

func (p *ProcessStore) Add(process *Process) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.store[process.ID] = process
}

func (p *ProcessStore) Range(action func(pp *Process)) {
	p.mx.Lock()
	defer p.mx.Unlock()

	for _, process := range p.store {
		action(process)
	}
}

func (p *Process) calculateDeadline() time.Time {
	return time.Now().Add(p.Task.GetDuration())
}
