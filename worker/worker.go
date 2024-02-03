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
	Run(ctx context.Context, stop chan struct{}) error
}

type BuildTask struct {
	ID       uuid.UUID
	Name     string
	Duration string
}

func (b *BuildTask) GetID() uuid.UUID {
	return b.ID
}

func (b *BuildTask) Run(ctx context.Context, stop chan struct{}) error {
	dur, err := time.ParseDuration(b.Duration)
	if err != nil {
		return err
	}

	timer := time.NewTimer(dur)
	slog.Info("starting building", "name", b.Name, "duration", dur.String())
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
