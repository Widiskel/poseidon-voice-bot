package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/widiskel/poseidon-voice-bot/internal/app/worker"
	"github.com/widiskel/poseidon-voice-bot/internal/integrations/gmail"
	"github.com/widiskel/poseidon-voice-bot/internal/model"
	"github.com/widiskel/poseidon-voice-bot/internal/utils"
)

type App struct{}

func New() *App { return &App{} }

func (app *App) Run() error {
	accounts, err := utils.LoadAccounts("accounts/accounts.json")
	if err != nil {
		return err
	}

	if err := setupGmailTokens(accounts); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(accounts))

	for idx, email := range accounts {
		sess := &model.Session{
			AccIdx: idx,
			Email:  email,
			Point:  0,
		}
		go func(s *model.Session) {
			defer wg.Done()
			worker.Run(s)
		}(sess)
	}

	wg.Wait()
	return nil
}

func setupGmailTokens(emails []string) error {
	for _, email := range emails {
		tokenPath := filepath.Join("accounts", fmt.Sprintf("%s-data.json", email))
		if _, err := os.Stat(tokenPath); err == nil {
			continue
		}

		_, err := gmail.NewService(context.Background(), "configs/credentials.json", tokenPath, email)

		if err != nil {
			return fmt.Errorf("gmail oauth for %s failed: %w", email, err)
		}
	}

	for _, email := range emails {
		tokenPath := filepath.Join("accounts", fmt.Sprintf("%s-data.json", email))
		if _, err := os.Stat(tokenPath); err != nil {
			return fmt.Errorf("missing token for %s after setup", email)
		}
	}
	return nil
}
