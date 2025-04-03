package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Tbits007/url-shortener/internal/lib/api/response"
	"github.com/Tbits007/url-shortener/internal/lib/logger/sl"
	"github.com/Tbits007/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
    GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        const op = "handlers.url.redirect.New"

        log = log.With(
            slog.String("op", op),
            slog.String("request_id", middleware.GetReqID(r.Context())),
        )		

        alias := chi.URLParam(r, "alias")
        if alias == "" {
            log.Info("alias is empty")
			render.Status(r, http.StatusNotFound)
            render.JSON(w, r, resp.Error("not found"))

            return
        }
		
        // Находим URL по алиасу в БД
        resURL, err := urlGetter.GetURL(alias)
        if errors.Is(err, storage.ErrURLNotFound) {
            // Не нашли URL, сообщаем об этом клиенту
            log.Info("url not found", "alias", alias)
			render.Status(r, http.StatusNotFound)
            render.JSON(w, r, resp.Error("not found"))

            return
        }
        if err != nil {
            // Не удалось осуществить поиск
            log.Error("failed to get url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
            render.JSON(w, r, resp.Error("internal error"))

            return
        }

        log.Info("got url", slog.String("url", resURL))

        // Делаем редирект на найденный URL
        http.Redirect(w, r, resURL, http.StatusFound)
    }		
}