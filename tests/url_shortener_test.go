package tests

import (
	"net/url"
	"testing"

	"github.com/Tbits007/url-shortener/internal/http-server/handlers/url/save"
	"github.com/Tbits007/url-shortener/internal/lib/random"
	"github.com/brianvoe/gofakeit"
	"github.com/gavv/httpexpect/v2"
)

const (
	host = "localhost:8080"
)

func TestURLShortener_HappyPath(t *testing.T) {
	// Универсальный способ создания URL
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	// Создаем клиент httpexpect
	e := httpexpect.Default(t, u.String())

	e.POST("/saveURL"). // Отправляем POST-запрос, путь - '/saveURL'
			WithJSON(save.Request{ // Формируем тело запроса
			URL:   gofakeit.URL(),             // Генерируем случайный URL
			Alias: random.NewRandomString(10), // Генерируем случайную строку
		}).
		WithBasicAuth("admin", "12345"). // Добавляем к запросу креды авторизации
		Expect().                          // Далее перечисляем наши ожидания от ответа
		Status(200).                       // Код должен быть 200
		JSON().Object().                   // Получаем JSON-объект тела ответа
		ContainsKey("alias")               // Проверяем, что в нём есть ключ 'alias'
}