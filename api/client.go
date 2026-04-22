package api

import (
	"yact/config"
	"yact/logic"
)

type Client interface {
	Init(cfg *config.Config)
	GetModelName() string
	Call(transaction logic.Transaction, think bool, systemPrompt string) (logic.Transaction, error)
}
