package login

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"log/slog"

	"URL-Shortener/internal/event"
	resp "URL-Shortener/internal/lib/api/response"
	"URL-Shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type AuthClient interface {
	Login (
		ctx context.Context,
		email string, 
		password string,
		appID int32,
	) (string, error)
}

type Response struct {
	resp.Response
	Token string `json:"token"`
}

type Request struct {
	Email string 	`json:"email"`
	Password string `json:"password"`
	AppID int32 	`json:"app_id"`
}

type KafkaProducer interface {
	ProduceEvent(
		ev event.Event, 
		topic string,
	) error
}

func responseOK(w http.ResponseWriter, r *http.Request, token string) {
	render.JSON(w, r, Response {
		Response: resp.OK(),
		Token: token,
	})
}

func New(log *slog.Logger, authClient AuthClient, kafka KafkaProducer) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		const op = "handlers.login.New"

		log = log.With(
			slog.String("op", op),
			slog.Any("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")

				render.JSON(w, r, resp.Error("request body is empty"))

				return
			}
			log.Error("failed to decode request body")

			render.JSON(w, r, resp.Error("failed to decode request body"))

			return 
		}

		log.Info("request body decoded", slog.Any("req", req))

		if err := validator.New().Struct(&req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request")

			render.JSON(w, r, resp.ValidateError(validateErr))

			return
		}

		email := req.Email
		if email == "" {
			log.Error("email is empty")

			render.JSON(w, r, resp.Error("email is empty"))

			return
		}

		password := req.Password
		if password == "" {
			log.Error("password is empty")

			render.JSON(w, r, resp.Error("password is empty"))

			return
		}

		appID := req.AppID
		if appID == 0 {
			log.Error("invalid appID")

			render.JSON(w, r, resp.Error("invalid appID"))

			return
		}

		token, err := authClient.Login(r.Context(), email, password, appID)
		if err != nil {
			log.Error("failed to login user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to login user"))

			return
		}

		loginEvent := event.AuthEvent{
			Type: "login",
			Email: email,
			TimeStamp: time.Now(),
		}

		err = kafka.ProduceEvent(loginEvent, "auth-topic")
		if err != nil {
			log.Error("failed to produce login event")
		}

		responseOK(w, r, token)
	}
}