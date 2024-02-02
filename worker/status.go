package worker

type Status uint

const (
	StatusPending Status = iota
	StatusInProgress
	StatusFailed
	StatusDone
	StatusInterrupted
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusInProgress:
		return "running"
	case StatusDone:
		return "done"
	case StatusFailed:
		return "failed"
	case StatusInterrupted:
		return "interrupted"
	}

	return ""
}
