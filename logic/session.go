package logic

import (
	"yact/config"
)

// Session holds what is needed to deliver one LLM answer later, possibly after a batch job has finished.
type Session struct {
	Linked   LinkedFiles       `json:"linked"`
	Mode     config.OutputMode `json:"mode"`
	Response string            `json:"response"`
}
