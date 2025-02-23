package delete

import (
	"URL-Shortener/internal/http-server/middleware/auth"
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

		if authErr, ok := auth.ErrorFromContext(r.Context()); ok {
			log.Error("authorization error", sl.Err(authErr))

			render.JSON(w, r, resp.Error("authorization error"))

			return
		}

		isAdmin, adminOk := auth.IsAdminFromContext(r.Context())
		// в В proto3 булево поле (bool) по умолчанию имеет значение false, 
		//и оно не сериализуется в JSON, если явно не установлено.
		uid, uidOk := auth.UIDFromContext(r.Context()) //TODO
		if !adminOk || !uidOk {
			log.Error("missing auth context")

			render.JSON(w, r, resp.Error("missing auth context"))

			return
		}
		if !isAdmin {
			log.Error("the user does not have enough rights to perform the action")

			render.JSON(w, r, resp.Error("the user does not have enough rights to perform the action"))
			
			return
		}

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

		log.Info("url deleted", slog.Int64("uid", uid))

		responseOK(w, r, id)
	}
}