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

	timer := time.NewTimer(p.Task.GetDuration())

	slog.Info("starting building", "name", p.Task.GetName(), "duration", p.Task.GetDuration().String())

Listener:
	for {
		select {
		case <-p.stop:
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
		case <-timer.C:
			slog.Info("timer expired")
			break Listener
		}
	}

	if err := p.Task.Run(context.Background()); err != nil {
		p.SetStatus(StatusFailed)
	} else {
		p.SetStatus(StatusDone)
	}
}

func (p *Process) Stop() {
	// Only send stop signals to tasks that are still running
	if p.Status == StatusInProgress {
		slog.Info("stopping process", "id", p.ID, "name", p.Task.GetName())
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
