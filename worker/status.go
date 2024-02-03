package worker

type Status uint

const (
	StatusPending Status = iota
	StatusInProgress
	StatusFailed
	StatusDone
	StatusInterrupted

	LabelStatusPending     string = "pending"
	LabelStatusInProgress  string = "running"
	LabelStatusDone        string = "done"
	LabelStatusFailed      string = "failed"
	LabelStatusInterrupted string = "interrupted"
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return LabelStatusPending
	case StatusInProgress:
		return LabelStatusInProgress
	case StatusDone:
		return LabelStatusDone
	case StatusFailed:
		return LabelStatusFailed
	case StatusInterrupted:
		return LabelStatusInterrupted
	}

	return ""
}
