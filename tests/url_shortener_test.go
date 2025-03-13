package tests

import (
	"URL-Shortener/internal/http-server/handlers/url/save"
	"URL-Shortener/internal/lib/random"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8082"
	passDefaultLen = 10
)

const app_id = 1

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",	
		Host: host,
	}

	e := httpexpect.Default(t, u.String())

	email := gofakeit.Email()
	password := randomFakePassword()

	response := e.POST("/register").WithJSON(map[string]string{
		"email": email,
		"password": password,
	}).Expect().Status(200)

	respJson := response.JSON().Object()

	status := respJson.Value("status").String().Raw()
	assert.Equal(t, status, "OK")
	
	uid := respJson.Value("uid").Number().Raw()
	require.NotEmpty(t, uid)
	
	response = e.POST("/login").WithJSON(map[string]interface{}{
		"email": email,
		"password": password,
		"app_id": app_id,
	}).Expect().Status(200)

	respJson = response.JSON().Object()

	token := respJson.Value("token").String().Raw()
	assert.NotEmpty(t, token)

	e.POST("/url").WithJSON(save.Request{
		URL: gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}).
	WithHeader("Authorization", "Bearer " + token).
	Expect().
	Status(200).
	JSON().Object().
	ContainsKey("alias")
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}