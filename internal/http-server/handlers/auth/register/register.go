package register

import (
	resp "URL-Shortener/internal/lib/api/response"
	"URL-Shortener/internal/lib/logger/sl"
	"context"
	"errors"
	"io"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type AuthClient interface {
	Register(
		ctx context.Context, 
		email string, 
		password string,
	) (int64, error)
}

type Response struct {
	resp.Response
	UserId int64 `json:"uid"`
}

type Request struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func responseOK(w http.ResponseWriter, r *http.Request, uid int64) {
	render.JSON(w, r, Response {
		Response: resp.OK(),
		UserId: uid,
	})
}

func New(log *slog.Logger, authClient AuthClient) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		const op = "handlers.Register.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			if errors.Is(err,io.EOF) {
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
			log.Error("invalid request", sl.Err(err))

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

		uid, err := authClient.Register(r.Context(), email, password)
		if err != nil {
			log.Error("failed to register user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to register user"))

			return
		}

		responseOK(w, r, uid)
	}
}



