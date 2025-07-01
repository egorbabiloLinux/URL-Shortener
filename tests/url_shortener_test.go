package tests

import (
	"URL-Shortener/internal/http-server/handlers/url/save"
	"URL-Shortener/internal/lib/random"
	"net/http"
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
	adminEmail = "admin@gmail.com"
	adminPassword = "1"
)

const app_id = 1
func TestAdminActionsURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host: host,
	}

	e := httpexpect.Default(t, u.String())

	response := e.POST("/login").WithJSON(map[string]interface{}{
		"email": adminEmail,
		"password": adminPassword,
		"app_id": app_id,
	}).Expect().Status(http.StatusOK)

	jsonResp := response.JSON().Object()

	token := jsonResp.Value("token").String().Raw()
	assert.NotEmpty(t, token)

	url := gofakeit.URL()
	alias := random.NewRandomString(10)

	e.POST("/url").WithJSON(save.Request{
		URL: url,
		Alias: alias,
	}).
	WithHeader("Authorization", "Bearer " + token).
	Expect().
	Status(http.StatusOK).
	JSON().Object().
	ContainsKey("alias")

	e.DELETE("/url/" + alias).
	WithHeader("Authorization", "Bearer " + token).
	Expect().
	Status(http.StatusOK).
	JSON().Object().
	ContainsKey("id")
}

func TestUserURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",	
		Host: host,
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL: u.String(),
		Client: client,
		Reporter: httpexpect.NewRequireReporter(t),
	})

	response := e.POST("/login").WithJSON(map[string]interface{}{
		"email": adminEmail,
		"password": adminPassword,
		"app_id": app_id,
	}).Expect().Status(http.StatusOK)

	jsonResp := response.JSON().Object()

	token := jsonResp.Value("token").String().Raw()
	assert.NotEmpty(t, token)

	url := gofakeit.URL()
	alias := random.NewRandomString(10)

	e.POST("/url").WithJSON(save.Request{
		URL: url,
		Alias: alias,
	}).
	WithHeader("Authorization", "Bearer " + token).
	Expect().
	Status(http.StatusOK).
	JSON().Object().
	ContainsKey("alias")

	email := gofakeit.Email()
	password := randomFakePassword()

	response = e.POST("/register").WithJSON(map[string]string{
		"email": email,
		"password": password,
	}).Expect().Status(http.StatusOK)

	respJson := response.JSON().Object()

	status := respJson.Value("status").String().Raw()
	assert.Equal(t, status, "OK")
	
	uid := respJson.Value("uid").Number().Raw()
	require.NotEmpty(t, uid)
	
	response = e.POST("/login").WithJSON(map[string]interface{}{
		"email": email,
		"password": password,
		"app_id": app_id,
	}).Expect().Status(http.StatusOK)
	respJson = response.JSON().Object()

	token = respJson.Value("token").String().Raw()
	assert.NotEmpty(t, token)

	e.GET("/" + alias).
	WithHeader("Authorization", "Bearer " + token).
	Expect().
	Status(http.StatusFound).
	Header("Location").IsEqual(url)
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}