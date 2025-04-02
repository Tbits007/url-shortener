package save

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "github.com/Tbits007/url-shortener/internal/lib/api/response"
	"github.com/Tbits007/url-shortener/internal/lib/logger/sl"
	"github.com/Tbits007/url-shortener/internal/lib/random"
	"github.com/Tbits007/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	validator "github.com/go-playground/validator/v10"
)

const aliasLength = 6

type Request struct {
    URL   string `json:"url" validate:"required,url"`
    Alias string `json:"alias,omitempty"`
}

type Response struct {
    resp.Response
    Alias string `json:"alias"`
}

type URLSaver interface {
    SaveURL(urlToSave, alias string) error
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
    render.JSON(w, r, Response{
        Response: resp.OK(),
        Alias:    alias,
    })
}


func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        const op = "handlers.url.save.New"

        // Добавляем к текущему объекту логгера поля op и request_id
        // Они могут очень упростить нам жизнь в будущем
        log = log.With(
            slog.String("op", op),
            slog.String("request_id", middleware.GetReqID(r.Context())),
        )

        // Создаем объект запроса и анмаршаллим в него запрос
        var req Request

        err := render.DecodeJSON(r.Body, &req)
        if errors.Is(err, io.EOF) {
            // Такую ошибку встретим, если получили запрос с пустым телом
            // Обработаем её отдельно
            log.Error("request body is empty")

            render.JSON(w, r, resp.Error("empty request"))
            return
        }
        if err != nil {
            log.Error("failed to decode request body", sl.Err(err))
            render.JSON(w, r, resp.Error("failed to decode request"))
            return
        }

        // Лучше больше логов, чем меньше - лишнее мы легко сможем почистить,
        // при необходимости. А вот недостающую информацию мы уже не получим.
        log.Info("request body decoded", slog.Any("req", req))

		// Создаем объект валидатора
		// и передаем в него структуру, которую нужно провалидировать
		if err := validator.New().Struct(req); err != nil {
			// Приводим ошибку к типу ошибки валидации
			validateErr := err.(validator.ValidationErrors)
		
			log.Error("invalid request", sl.Err(err))
            
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}
		
		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}	
		
        err = urlSaver.SaveURL(req.URL, alias)
        if errors.Is(err, storage.ErrURLExists) {
            log.Info("url already exists", slog.String("url", req.URL))
            w.WriteHeader(http.StatusBadRequest)
            render.JSON(w, r, resp.Error("url already exists"))
            return
        }
        if err != nil {
            log.Error("failed to add url", sl.Err(err))
            w.WriteHeader(http.StatusInternalServerError)
            render.JSON(w, r, resp.Error("failed to add url"))
            return
        }

        responseOK(w, r, alias)
    }	
		 
}