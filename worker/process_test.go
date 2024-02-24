package worker

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	short = 500 * time.Millisecond
	long  = 15 * time.Second
)

type TestTask struct {
	Name  string
	ID    uuid.UUID
	Error error
}

func (tt *TestTask) GetID() uuid.UUID             { return tt.ID }
func (tt *TestTask) GetDuration() time.Duration   { return 5 * time.Second }
func (tt *TestTask) GetName() string              { return tt.Name }
func (tt *TestTask) UpdateDB(status Status) error { return nil }
func (tt *TestTask) Run(ctx context.Context) error {
	slog.Debug("running test task")
	return tt.Error
}

var _ Task = (*TestTask)(nil)

func TestRunTask(t *testing.T) {

	tests := []struct {
		task              Task
		interrupt         bool
		status            Status
		expectedStatus    Status
		expectedRemaining bool
		invalidDeadline   bool
	}{
		{
			task: &BuildTask{
				Name:     "done",
				Duration: short,
			},
			interrupt:         false,
			status:            StatusPending,
			expectedStatus:    StatusDone,
			expectedRemaining: false,
			invalidDeadline:   false,
		},
		{
			task: &BuildTask{
				Name:     "interrupted",
				Duration: long,
			},
			interrupt:         true,
			status:            StatusPending,
			expectedStatus:    StatusInterrupted,
			expectedRemaining: true,
			invalidDeadline:   false,
		},
		{
			task: &BuildTask{
				Name:     "started",
				Duration: short,
			},
			interrupt:         false,
			status:            StatusDone,
			expectedStatus:    StatusDone,
			expectedRemaining: false,
			invalidDeadline:   false,
		},
		{
			task: &TestTask{
				Name:  "error",
				ID:    uuid.New(),
				Error: fmt.Errorf("test error"),
			},
			interrupt:         false,
			status:            StatusPending,
			expectedStatus:    StatusFailed,
			expectedRemaining: false,
			invalidDeadline:   false,
		},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			process := newProcess(tt.task)
			process.Status = tt.status

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go process.Start(wg)
			if tt.interrupt {
				time.Sleep(1 * time.Second)
				process.Stop()
			}
			wg.Wait()

			assert.Equal(t, tt.expectedStatus, process.Status)

			if tt.expectedRemaining {
				assert.Greater(t, process.Remaining, time.Duration(0))
			}
		}
		t.Run(tt.task.GetName(), tf)
	}
}

func TestInvalidDeadline(t *testing.T) {
	task := &BuildTask{
		ID:       uuid.New(),
		Name:     "test",
		Duration: 30 * time.Second,
	}
	process := newProcess(task)
	process.Status = StatusPending

	buf := bytes.NewBuffer([]byte(""))
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	patches := gomonkey.ApplyPrivateMethod(process, "calculateDeadline", func(_ *Process) time.Time {
		return time.Now().Add(-1 * 1 * time.Hour)
	})
	defer patches.Reset()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go process.Start(wg)
	time.Sleep(1 * time.Second)
	process.Stop()
	wg.Wait()

	assert.Equal(t, StatusInterrupted, process.Status)
	assert.Contains(t, buf.String(), "task was interrupted but remaining time less than 0")
	assert.Equal(t, process.Remaining, time.Duration(0))
}
