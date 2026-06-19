package worker

import "errors"

var (
	errNoWorker  = errors.New("no online worker for agent")
	errSendFull  = errors.New("worker send buffer full")
)
