package domain

import (
	"context"
	"tt-newbotsacc/pkg/browser"
)

type TaskPayload struct {
	Action    string `json:"action"`
	TargetURL string `json:"target_url"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type ActionHandler interface {
	Execute(ctx context.Context, b *browser.Browser, payload []byte) error
}