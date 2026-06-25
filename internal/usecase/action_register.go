package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"tt-newbotsacc/internal/domain"
	"tt-newbotsacc/pkg/browser"
)

type RegisterAction struct{}

func NewRegisterAction() *RegisterAction {
	return &RegisterAction{}
}

func (a *RegisterAction) Execute(ctx context.Context, b *browser.Browser, payload []byte) error {
	var p domain.TaskPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	if err := b.Navigate(p.TargetURL); err != nil {
		return err
	}

	page := b.GetPage()
	if page == nil {
		return fmt.Errorf("page is not initialized")
	}

	emailInput, err := page.Element("input[type=email]")
	if err != nil {
		return fmt.Errorf("failed to find email field: %w", err)
	}

	if err := emailInput.Input(p.Email); err != nil {
		return err
	}

	passInput, err := page.Element("input[type=password]")
	if err != nil {
		return fmt.Errorf("failed to find password field: %w", err)
	}
	if err := passInput.Input(p.Password); err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	submitBtn, err := page.Element("button[type=submit]")
	if err != nil {
		return fmt.Errorf("failed to find submit button: %w", err)
	}
	if err := submitBtn.Click("left", 1); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	return nil
}