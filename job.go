package rigel

import "time"

type _Job struct {
	Queue string
	Job   []byte
}

type _FailedJob struct {
	Job      string    `json:"job"`
	Err      string    `json:"error"`
	FailedAt time.Time `json:"failed_at"`
}
