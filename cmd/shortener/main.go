package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"shortener/internal/app/server"
)

// Options — опции приложения
type Options struct {
	ServerAddress string `long:"address" short:"a" env:"SERVER_ADDRESS" default:"localhost:8080" description:"server address"`
	BaseURL       string `long:"url" short:"b" env:"BASE_URL" default:"localhost:8080" description:"server address"`
	Debug         bool   `long:"dbg" env:"DEBUG" description:"debug mode" required:"false"`
}

var revision = "unknown"

func main() {
	var opts Options
	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		setupLog(opts.Debug)

		log.Info().Str("revision", revision).Send()

		err := server.Execute(args, opts.ServerAddress, opts.BaseURL)
		if err != nil {
			log.Info().Err(err).Msg("ошибка выполнения команды")
		}

		return err
	}

	if _, err := p.Parse(); err != nil {
		//var flagsErr *flags.Error
		//if errors.Is(err, flagsErr) && flagsErr.Type == flags.ErrHelp {
		//	os.Exit(0)
		//} else {
		os.Exit(1)
		//}
	}
}

func setupLog(debug bool) {
	if debug {
		// В дебаге выводим плоские строки
		cw := zerolog.ConsoleWriter{Out: os.Stdout}
		log.Logger = zerolog.New(cw).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		return
	}

	// Иначе логируем в stdout в json'е
	log.Logger = zerolog.New(os.Stdout).With().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	zerolog.MessageFieldName = "m"
	zerolog.ErrorFieldName = "e"
	zerolog.LevelFieldName = "l"
	zerolog.TimestampFieldName = "t"
}
