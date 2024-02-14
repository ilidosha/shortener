package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"shortener/internal/app/shortener"
	"testing"
)

func TestShortenURLHandler(t *testing.T) {
	// Создаем тестовый хендлер
	mockRest := &Rest{
		storage: shortener.Storage{
			Records: map[string]string{},
		},
		baseURL: "localhost:8080",
	}

	// Тестовый запрос
	longURL := "http://example.com/very/long/url"
	requestBody := bytes.NewBuffer([]byte(longURL))
	request, err := http.NewRequest("POST", "/shorten", requestBody)
	if err != nil {
		t.Fatal(err)
	}

	// Тестовый ResponseWriter
	recorder := httptest.NewRecorder()

	// Вызываем хендлер
	mockRest.ShortenURL(recorder, request)

	// Проверяем результат
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, recorder.Code)
	}

	expectedShortURL := "localhost:8080/YYRVyE" // замените на ожидаемое значение
	if body := recorder.Body.String(); len(body) != len(expectedShortURL) {
		t.Errorf("Expected response body %s, got %s", expectedShortURL, body)
	}
}

func TestShortenURLHandlerErrorCases(t *testing.T) {
	// Тестовые случаи с ошибками

	// Тестовый хендлер с имитацией ошибок
	mockRest := &Rest{
		storage: shortener.Storage{
			Records: nil, // имитация ошибки при добавлении в хранилище
		},
		baseURL: "localhost:8080",
	}

	testCases := []struct {
		name            string
		requestBody     io.Reader
		expectedCode    int
		expectedSubstr  string
		storageWithData bool
	}{
		{
			name:            "Invalid request body",
			requestBody:     bytes.NewBuffer([]byte("invalid")),
			expectedCode:    http.StatusBadRequest,
			expectedSubstr:  "Cannot read request body",
			storageWithData: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Обнуляем хранилище, если нужно
			if !tc.storageWithData {
				mockRest.storage.Records = nil
			}

			// Создаем тестовый запрос
			request, err := http.NewRequest("POST", "/shorten", tc.requestBody)
			if err != nil {
				t.Fatal(err)
			}

			// Тестовый ResponseWriter
			recorder := httptest.NewRecorder()

			// Вызываем хендлер
			mockRest.ShortenURL(recorder, request)

			// Проверяем результат
			if recorder.Code != tc.expectedCode {
				t.Errorf("Expected status code %d, got %d", tc.expectedCode, recorder.Code)
			}
		})
	}
}

func TestReturnURLHandler(t *testing.T) {
	// Создаем тестовый хендлер
	mockRest := &Rest{
		storage: shortener.Storage{
			Records: map[string]string{
				"abc123": "http://example.com/long/url",
			},
		},
		baseURL: "localhost:8080",
	}

	// Тестовый запрос
	request, err := http.NewRequest("GET", "/abc123", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Тестовый ResponseWriter
	recorder := httptest.NewRecorder()

	// Вызываем хендлер
	mockRest.ReturnURL(recorder, request)

	// Проверяем результат
	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
	}

	expectedLocation := "http://example.com/long/url"
	if location := recorder.Header().Get("Location"); location != expectedLocation {
		t.Errorf("Expected Location header %s, got %s", expectedLocation, location)
	}
}

func TestReturnURLHandlerKeyNotFound(t *testing.T) {
	// Тест, когда ключ не найден в хранилище

	// Создаем тестовый хендлер
	mockRest := &Rest{
		storage: shortener.Storage{
			Records: map[string]string{},
		},
		baseURL: "http://example.com",
	}

	// Тестовый запрос
	request, err := http.NewRequest("GET", "/nonexistentkey", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Тестовый ResponseWriter
	recorder := httptest.NewRecorder()

	// Вызываем хендлер
	mockRest.ReturnURL(recorder, request)

	// Проверяем результат
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
