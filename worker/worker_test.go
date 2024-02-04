package worker

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSystem(t *testing.T) {
	system := NewSystem()

	shortID := uuid.New()
	longID := uuid.New()
	system.Inbox() <- &BuildTask{
		Name:     "test",
		Duration: 500 * time.Millisecond,
		ID:       shortID,
	}

	system.Inbox() <- &BuildTask{
		Name:     "test",
		Duration: 10 * time.Second,
		ID:       longID,
	}

	time.Sleep(1 * time.Second)

	system.Stop()

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
