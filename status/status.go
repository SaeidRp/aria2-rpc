package status

type Status string

const (
	Active   Status = "active"
	Waiting  Status = "waiting"
	Paused   Status = "paused"
	Error    Status = "error"
	Complete Status = "complete"
	Removed  Status = "removed"
)
