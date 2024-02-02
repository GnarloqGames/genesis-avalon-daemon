package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Process struct {
	ID     uuid.UUID
	Status Status
	Task   Task

	signals chan signal
}

func newProcess(task Task) *Process {
	process := &Process{
		ID:   uuid.New(),
		Task: task,

		signals: make(chan signal),
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

func (p *Process) Start() {
	go func() {
		for signal := range p.signals {
			p.SetStatus(signal.status())
			slog.Info("received signal",
				"task_id", signal.task(),
				"status", signal.status(),
				"timestamp", signal.timestamp().Format(time.RFC1123),
			)
		}
	}()

	p.SetStatus(StatusInProgress)
	if err := p.Task.Run(context.Background(), p.signals); err != nil {
		p.SetStatus(StatusFailed)
	} else {
		p.SetStatus(StatusDone)
	}
}

type ProcessStore struct {
	mx    *sync.Mutex
	store map[uuid.UUID]*Process
}
