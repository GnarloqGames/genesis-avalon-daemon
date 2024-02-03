package worker

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunTask(t *testing.T) {

	tests := []struct {
		name           string
		duration       string
		interrupt      bool
		status         Status
		expectedStatus Status
	}{
		{
			name:           "done",
			duration:       "500ms",
			interrupt:      false,
			status:         StatusPending,
			expectedStatus: StatusDone,
		},
		{
			name:           "interrupted",
			duration:       "30s",
			interrupt:      true,
			status:         StatusPending,
			expectedStatus: StatusInterrupted,
		},
		{
			name:           "error",
			duration:       "invalid",
			interrupt:      false,
			status:         StatusPending,
			expectedStatus: StatusFailed,
		},
		{
			name:           "started",
			duration:       "500ms",
			interrupt:      false,
			status:         StatusDone,
			expectedStatus: StatusDone,
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

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go process.Start(wg)
			if tt.interrupt {
				time.Sleep(1 * time.Second)
				process.Stop()
			}
			wg.Wait()

			assert.Equal(t, tt.expectedStatus, process.Status)
		}
		t.Run(tt.name, tf)
	}
}
