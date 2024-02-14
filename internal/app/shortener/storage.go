package shortener

import (
	"errors"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net/url"
)

// Storage содержит мапу всех записаных урлов и их хешей
type Storage struct {
	Records map[string]string
}

// Append добавляет новую ссылку в мапу
func (s *Storage) Append(LongURL string) (string, error) {
	_, errIsNotURL := url.ParseRequestURI(LongURL)
	if errIsNotURL != nil {
		log.Error().Err(errIsNotURL).Msg("Provided string is not an url")
		return "", errIsNotURL
	}
	shortKey := generateShortKey()
	// check if key exists
	_, ok := s.Records[LongURL]
	// If the key NOT exists
	if !ok {
		s.Records[shortKey] = LongURL
		return shortKey, nil
	}
	if ok {
		// regenerate key if got a collision, yes i know this is recursion and may end poorly
		return s.Append(LongURL)
	}
	return "", errors.New("something went wrong")
}

// честно скопипащеная функция генерации рандомных символов из интернетов
// 56800235584 комбинации, главно отслеживать повторы иначе будут коллизии
// и вообще добавить возможности расширения при заполнении на определённую длинну
// или проводить переодическую чистку чтобы сохранять не большую длинну сокращённой ссылки
// но в таком случае придётся определять какими пользуются а какими нет
func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}
