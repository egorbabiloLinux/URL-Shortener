package save

import (
	resp "URL-Shortener/internal/lib/api/response"
	"URL-Shortener/internal/lib/logger/sl"
	"URL-Shortener/internal/lib/random"
	"URL-Shortener/internal/storage"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const aliasLength = 6

type Request struct {
	URL   string `json:"url" validate:"required,url"` 
	// required: Указывает, что поле обязательно для заполнения; url: Проверяет, что значение поля соответствует формату URL
	Alias string `json:"alias,omitempty"` 
	// omitempty - если поле имеет пустое значение оно не должно быть включено в десериализацию.
}

type Response struct {
	resp.Response
    Alias  string `json:"alias"`
}

type URLSaver interface {
	SaveURL(URL, alias string) (int64, error)
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response {
		Response: resp.OK(),
		Alias: 	  alias,
	})
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())), 
			// Middleware RequestID из пакета chi/middleware автоматически создает уникальный идентификатор запроса 
			// для каждого HTTP-запроса и сохраняет его в контекст запроса r.Context()
			// Функция middleware.GetReqID извлекает значение уникального идентификатора из контекста
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("request body is empty"))

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request body"))

			return
		}

		log.Info("request body decoded", slog.Any("req", req)) // slog.Any принимает значение,
		// определяет его тип и автоматически преобразует его в строку или другой подходящий формат для включения в лог.

		if err := validator.New().Struct(&req); err != nil { // «required,url» в объекте Request — он как раз будет использован валидатором.
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidateError(validateErr))

			return
		}

		alias := req.Alias

		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("URL already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("URL already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add URL", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add URL"))

			return
		}

		log.Info("URL added", slog.Int64("id", id))

		responseOK(w, r, alias)
	}
}