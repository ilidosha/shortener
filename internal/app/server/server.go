package server

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

// Execute — точка входа в команду app
func Execute(_ []string, serverAddress string, baseURL string) error {
	app, err := newServerApp(serverAddress, baseURL)
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
	params params
	rest   Rest
}

type params struct {
	ServerAddress string
	BaseUrl       string
}

// newServerApp собирает зависимости и формирует приложение готовое для запуска
func newServerApp(serverAddress, baseURL string) (*serverApp, error) {
	app := &serverApp{
		rest: Rest{},
		params: params{
			ServerAddress: serverAddress,
			BaseUrl:       baseURL,
		},
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
	app.rest.Run(app.params.ServerAddress, app.params.BaseUrl)

	return nil
}
