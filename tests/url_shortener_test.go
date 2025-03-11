package tests

import (
	"URL-Shortener/internal/http-server/handlers/url/save"
	"URL-Shortener/internal/lib/random"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",	
		Host: host,
	}

	e := httpexpect.Default(t, u.String())

	username := gofakeit.Email()
	password := randomFakePassword()

	response := e.POST("/register").WithJSON(map[string]string{
		"username": username,
		"password": password,
	})

	e.POST("/url").WithJSON(save.Request{
		URL: gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}).
	
	
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}