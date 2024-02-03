package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type TestTask struct {
	ID      uuid.UUID
	Stopped bool
	Wg      *sync.WaitGroup
}

func (tt *TestTask) GetID() uuid.UUID { return tt.ID }
func (tt *TestTask) Run(ctx context.Context, stop chan struct{}) error {
	<-stop
	tt.Stopped = true
	tt.Wg.Done()
	return nil
}

var _ Task = (*TestTask)(nil)

func TestHandleStopSignal(t *testing.T) {
	taskID := uuid.New()

	tt := &TestTask{
		ID: taskID,
		Wg: &sync.WaitGroup{},
	}
	p := newProcess(tt)
	p.Status = StatusInProgress

	store := NewStore()
	store.Add(p)

	system := &System{
		processes: store,
	}

	tt.Wg.Add(1)
	go tt.Run(context.Background(), p.stop) //nolint:errcheck
	handleStopSignal(system)
	tt.Wg.Wait()
	assert.True(t, tt.Stopped)
}

func TestSystem(t *testing.T) {
	system := NewSystem()

	shortID := uuid.New()
	longID := uuid.New()
	system.Inbox() <- &BuildTask{
		Name:     "test",
		Duration: "500ms",
		ID:       shortID,
	}

	system.Inbox() <- &BuildTask{
		Name:     "test",
		Duration: "10s",
		ID:       longID,
	}

	time.Sleep(1 * time.Second)

	system.Stop()

	system.wg.Wait()

	assert.Equal(t, 2, len(system.processes.store))

	// short task assertions
	assert.NotNil(t, system.processes.store[shortID])
	assert.Equal(t, StatusDone, system.processes.store[shortID].Status)

	assert.NotNil(t, system.processes.store[longID])
	assert.Equal(t, StatusInterrupted, system.processes.store[longID].Status)
}

func TestStatusStringer(t *testing.T) {
	tests := []struct {
		status         Status
		expectedString string
	}{
		{
			status:         StatusInProgress,
			expectedString: LabelStatusInProgress,
		},
		{
			status:         StatusPending,
			expectedString: LabelStatusPending,
		},
		{
			status:         StatusDone,
			expectedString: LabelStatusDone,
		},
		{
			status:         StatusFailed,
			expectedString: LabelStatusFailed,
		},
		{
			status:         StatusInterrupted,
			expectedString: LabelStatusInterrupted,
		},
		{
			status:         Status(100),
			expectedString: "",
		},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			assert.Equal(t, tt.expectedString, tt.status.String())
		}

		label := tt.expectedString
		if label == "" {
			label = "invalid"
		}
		t.Run(label, tf)
	}
}
