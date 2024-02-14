package server

import (
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

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
	//short := chi.URLParam(r, "short")э

	s := r.URL.Path[1:]
	log.Info().Msgf("Получен запрос на возврат урла короткий урл: %v", s)
	_, ok := rest.storage.Records[s]
	if ok {

		w.Header().Set("Location", rest.storage.Records[s])
		w.WriteHeader(http.StatusTemporaryRedirect)

		//http.Redirect(w, r, rest.storage.Records[short], http.StatusTemporaryRedirect)
		//w.Write([]byte("Location: " + rest.storage.Records[short]))
	}
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Info().Msg("provided key does not appear in the map")
	}
}
