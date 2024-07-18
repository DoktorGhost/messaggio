package handlers

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "messaggio/docs"
	"messaggio/internal/config"
	"messaggio/internal/logger"
	"messaggio/internal/models"
	"messaggio/internal/usecase"
	"net/http"
)

func InitRoutes(useCase usecase.UseCaseStorage, ctx context.Context, conf *config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(logger.WithLogging)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost"+":"+conf.SERVER_PORT+"/swagger/doc.json"), //The url pointing to API definition
	))

	r.Post("/message", func(w http.ResponseWriter, r *http.Request) {
		HandlerCreateMessage(w, r, useCase, ctx)
	})
	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		HandlerGetStats(w, r, useCase, ctx)
	})

	return r
}

// @Summary Добавление сообщения
// @Description Читает собщение из тела запроса, записывает в БД и отправляетв Кафку
// @Tags Message
// @Accept json
// @Produce json
// @Param body body models.Request true "Текст сообщения"
// @Success 200
// @Failure 400
// @Failure 405
// @Failure 500
// @Router /message [post]
func HandlerCreateMessage(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage, ctx context.Context) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request models.Request

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if request.Content == "" {
		logger.SugaredLogger().Info("Пустое содержимое запроса")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message := models.Message{
		Content: request.Content,
		Status:  "pending",
	}

	id, err := useCase.UseCaseCreate(ctx, &message)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	message.ID = id

	err = useCase.UseCaseSendMessage(message)

	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Статистика обработанных сообщений
// @Description Выводит количество всех сообщений, количество обработанных и не обработанных сообщений.
// @Tags Stats
// @Produce json
// @Success 200 {object} models.Stats
// @Failure 405
// @Failure 500
// @Router /stats [get]
func HandlerGetStats(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage, ctx context.Context) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	stats, err := useCase.UseCaseGetStats(ctx)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(stats)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
