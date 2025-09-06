package main

import (
	"time"

	"github.com/widiskel/poseidon-voice-bot/internal/app"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/logger"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/spinner"
)

func main() {
	_ = logger.Init("logs/app.log")
	defer logger.Close()

	spinner.StartUISystem()
	defer spinner.StopUISystem()

	if err := app.New().Run(); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)
}
