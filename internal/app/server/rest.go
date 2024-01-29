package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"shortener/internal/app/shortener"
	"time"
)

// Rest реализует REST API сервис
type Rest struct {
	httpServer *http.Server
	storage    shortener.Storage
	baseURL    string
}

// Run запускает REST-сервис
func (rest *Rest) Run(serverAddress, baseURL string) {
	rest.httpServer = &http.Server{
		Addr:              serverAddress,
		Handler:           rest.routes(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	rest.storage = shortener.Storage{
		Records: make(map[string]string),
	}
	rest.baseURL = baseURL

	err := rest.httpServer.ListenAndServe()
	log.Info().Err(err).Msg("http-сервер остановлен")
}

// Shutdown останавливает REST-сервис
func (rest *Rest) Shutdown() {
	// Ожидаем остановки http-сервера секунду
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if rest.httpServer != nil {
		err := rest.httpServer.Shutdown(ctx)
		if err != nil {
			log.Error().Err(err).Msg("ошибка при отключении http-сервера")
		}
	}
}

func (rest *Rest) routes() chi.Router {
	router := chi.NewRouter()

	// Глобальные мидлвари
	router.Use(hlog.NewHandler(log.Logger))
	router.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		var event *zerolog.Event
		if status < 400 {
			// Ответы с успешными кодами логируем только в Debug
			event = hlog.FromRequest(r).Debug()
		} else {
			event = hlog.FromRequest(r).Info()
		}

		event.
			Str("source", "http").
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	router.Use(hlog.RequestIDHandler("request_id", "X-Request-Id"))
	router.Use(hlog.RemoteAddrHandler("remote_addr"))
	router.Use(middleware.Recoverer)

	// Публичный API
	router.Route("/", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/", rest.ShortenURL)
		})
		r.Group(func(r chi.Router) {
			r.Get("/{short}", rest.ReturnURL)
		})
	})

	return router
}

func (rest *Rest) ShortenURL(w http.ResponseWriter, r *http.Request) { //nolint:bodyClose // cause body is closed in func
	responseData, err := io.ReadAll(r.Body)
	longURL := string(responseData)

	defer func(Body io.ReadCloser) {
		errBodyClose := Body.Close()
		if errBodyClose != nil {
			log.Error().Err(errBodyClose).Msg("Cannot close request body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}(r.Body)

	if err != nil {
		log.Error().Err(err).Msg("Cannot read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortened, errAppend := rest.storage.Append(longURL)
	if errAppend != nil {
		log.Error().Err(errAppend).Msg("Cannot append to map")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Info().Msgf("Получен запрос на запись, коротки урл: %v, длинный: %v", shortened, longURL)

	w.WriteHeader(http.StatusCreated)
	_, errWrite := w.Write([]byte(rest.baseURL + "/" + shortened))
	if errWrite != nil {
		log.Error().Err(errWrite).Msg("Cannot write response")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (rest *Rest) ReturnURL(w http.ResponseWriter, r *http.Request) {
	short := chi.URLParam(r, "short")
	_, ok := rest.storage.Records[short]
	if ok {
		log.Info().Msgf("Получен запрос на возврат урла короткий урл: %v, длинный: %v", short, rest.storage.Records[short])
		//w.Header().Set("Location", rest.storage.Records[short])
		w.WriteHeader(http.StatusTemporaryRedirect)
		//http.Redirect(w, r, rest.storage.Records[short], http.StatusTemporaryRedirect)
		w.Write([]byte("Location: " + rest.storage.Records[short]))
	}
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Info().Msg("provided key does not appear in the map")
	}
}
