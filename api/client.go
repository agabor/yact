package api

import (
	"yact/config"
)

type Role string

const (
	RoleTypeAssistant Role = "Assistant"
	RoleTypeUser      Role = "User"
)

type Message struct {
	Role    Role
	Content string
}

type Client interface {
	Init(cfg *config.Config)
	GetModelName() string
	Call(messages []Message, systemPrompt string) (Message, error)
}
