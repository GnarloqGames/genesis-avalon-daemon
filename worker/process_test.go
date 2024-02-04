package worker

import (
	"bytes"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestRunTask(t *testing.T) {

	tests := []struct {
		name              string
		duration          time.Duration
		interrupt         bool
		status            Status
		expectedStatus    Status
		expectedRemaining bool
		invalidDeadline   bool
	}{
		{
			name:              "done",
			duration:          500 * time.Millisecond,
			interrupt:         false,
			status:            StatusPending,
			expectedStatus:    StatusDone,
			expectedRemaining: false,
			invalidDeadline:   false,
		},
		{
			name:              "interrupted",
			duration:          30 * time.Second,
			interrupt:         true,
			status:            StatusPending,
			expectedStatus:    StatusInterrupted,
			expectedRemaining: true,
			invalidDeadline:   false,
		},
		{
			name:              "interrupted_invalid_duration",
			duration:          30 * time.Second,
			interrupt:         true,
			status:            StatusPending,
			expectedStatus:    StatusInterrupted,
			expectedRemaining: false,
			invalidDeadline:   true,
		},
		{
			name:              "started",
			duration:          500 * time.Millisecond,
			interrupt:         false,
			status:            StatusDone,
			expectedStatus:    StatusDone,
			expectedRemaining: false,
			invalidDeadline:   false,
		},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			task := &BuildTask{
				Name:     tt.name,
				Duration: tt.duration,
			}
			process := newProcess(task)
			process.Status = tt.status

			buf := bytes.NewBuffer([]byte(""))

			logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))
			slog.SetDefault(logger)

			if tt.invalidDeadline {
				patches := gomonkey.ApplyPrivateMethod(process, "calculateDeadline", func(_ *Process) time.Time {
					return time.Now().Add(-1 * 1 * time.Hour)
				})
				defer patches.Reset()
			}

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go process.Start(wg)
			if tt.interrupt {
				time.Sleep(1 * time.Second)
				process.Stop()
			}
			wg.Wait()

			assert.Equal(t, tt.expectedStatus, process.Status)

			if tt.invalidDeadline {
				assert.Contains(t, buf.String(), "task was interrupted but remaining time less than 0")
			}

			if tt.expectedRemaining {
				assert.Greater(t, process.Remaining, time.Duration(0))
			}
		}
		t.Run(tt.name, tf)
	}
}
