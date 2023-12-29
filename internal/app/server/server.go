package server

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

// ServerCommand опции для команды app
type ServerCommand struct {
	Address string `long:"address" short:"a" env:"SERVER_ADDRESS" default:"localhost:8080" description:"server address"`
	BaseURL string `long:"url" short:"b" env:"BASE_URL" default:"localhost:8080" description:"server address"`
}

// Execute — точка входа в команду app
func (cmd *ServerCommand) Execute(_ []string) error {

	app, err := cmd.newServerApp()
	if err != nil {
		log.Fatal().Err(err).Msg("не удалось создать сервер")
	}

	// Server run context
	serverCtx, serverCancel := context.WithCancel(context.Background())
	go func() {
		// Listen for syscall signals for process to interrupt/quit
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig
		serverCancel()
	}()

	// Run the app
	if err = app.run(serverCtx); err != nil {
		log.Fatal().Err(err).Msg("не удалось запустить приложение")
	}

	// Wait for app context to be stopped
	<-serverCtx.Done()
	return nil
}

// serverApp содержит все активные объекты: http-app и фоновые сервисы
type serverApp struct {
	*ServerCommand
	rest Rest
}

// newServerApp собирает зависимости и формирует приложение готовое для запуска
func (cmd *ServerCommand) newServerApp() (*serverApp, error) {
	app := &serverApp{
		ServerCommand: cmd,
		rest:          Rest{},
	}

	log.Info().Msg("зависимости построены")
	return app, nil
}

// Запускает приложение
func (app *serverApp) run(ctx context.Context) error { // nolint:unparam // error понадобится позже
	// При отмене контекста останавливаем http-сервер
	go func() {
		<-ctx.Done()
		log.Info().Msg("получен сигнал остановки сервиса")
		app.rest.Shutdown()
	}()

	// Запускаем http-сервер
	app.rest.Run(app.Address, app.BaseURL)

	return nil
}
