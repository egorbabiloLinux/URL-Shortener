package delete

import (
	resp "URL-Shortener/internal/lib/api/response"
	"URL-Shortener/internal/lib/logger/sl"
	"URL-Shortener/internal/storage"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLDelete interface {
	DeleteURL(alias string) (int64, error)
}

type Response struct {
	resp.Response
    Id  int64 `json:"id"`
}

func responseOK(w http.ResponseWriter, r *http.Request, id int64) {
	render.JSON(w, r, Response {
		Response: resp.OK(),
		Id: 	  id,
	})
}

func New(log *slog.Logger, urlDelete URLDelete) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("alias is empty"))

			return
		}

		id, err := urlDelete.DeleteURL(alias)
		if errors.Is(err, storage.ErrNotFound) {
			log.Error("url not found", sl.Err(err))

			render.JSON(w, r, resp.Error("url not found"))

			return
		}
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to delete url"))

			return
		}

		log.Info("url deleted")

		responseOK(w, r, id)
	}
}