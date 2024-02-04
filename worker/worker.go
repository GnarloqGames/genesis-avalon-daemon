package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrInterrupted error = fmt.Errorf("interrupted")

type System struct {
	processes *ProcessStore
	inbox     chan Task
	stop      chan struct{}
	wg        *sync.WaitGroup
}

type Task interface {
	GetID() uuid.UUID
	GetDuration() time.Duration
	Run(ctx context.Context, stop chan struct{}) error
}

func NewSystem() *System {
	system := &System{
		processes: &ProcessStore{
			mx:    &sync.Mutex{},
			store: make(map[uuid.UUID]*Process),
		},
		inbox: make(chan Task),
		stop:  make(chan struct{}),
		wg:    &sync.WaitGroup{},
	}

	go func() {
		for {
			select {
			case <-system.stop:
				handleStopSignal(system)

				return
			case task := <-system.inbox:
				handleInboxTask(system, task)
			}
		}
	}()

	return system
}

func handleStopSignal(system *System) {
	slog.Info("stopping worker system")

	system.processes.Range(func(pp *Process) {
		pp.Stop()
	})
}

func handleInboxTask(system *System, task Task) {
	slog.Info("received task", "id", task.GetID())
	process := newProcess(task)
	system.processes.Add(process)

	system.wg.Add(1)
	go process.Start(system.wg)
}

func (s *System) Inbox() chan Task {
	return s.inbox
}

func (s *System) Stop() {
	s.stop <- struct{}{}
	s.wg.Wait()
}
