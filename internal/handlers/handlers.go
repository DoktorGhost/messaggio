package handlers

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"messaggio/internal/config"
	"messaggio/internal/logger"
	"messaggio/internal/models"
	"messaggio/internal/usecase"
	"net/http"
)

func InitRoutes(useCase usecase.UseCaseStorage, ctx context.Context, conf *config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(logger.WithLogging)
	r.Post("/message", func(w http.ResponseWriter, r *http.Request) {
		HandlerCreateMessage(w, r, useCase, ctx)
	})
	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		HandlerGetStats(w, r, useCase, ctx)
	})

	/*
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://"+conf.SERVER_HOST+":"+conf.SERVER_PORT+"/swagger/doc.json"), //The url pointing to API definition
		))

	*/
	return r
}

// @Summary Добавление нового пользователя
// @Description Добавляет нового пользователя на основе серии и номера паспорта, обогащает информацию через внешний API (если в .env не указан URL API - получим ответ 500)
// @Tags Users
// @Accept json
// @Produce json
// @Param body body models.PassportRequest true "Серия и номер пасспорта в формате `1234 123456` (4 цифры, пробел, 6 цифр)"
// @Success 200 {string} string "UserID"
// @Failure 400 {string} string "Ошибка декодирования тела запроса"
// @Failure 409 {string} string "Ошибка записи: Пользователь с таким номером паспорта уже существует"
// @Failure 422 {string} string "Ошибка валидации серии паспорта или номера паспорта"
// @Failure 500 {string} string "Ошибка сервера"
// @Failure 503 {string} string "Ошибка запроса к стороннему API"
// @Router /user [post]
func HandlerCreateMessage(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage, ctx context.Context) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if request.Content == "" {
		logger.SugaredLogger().Info("Content is required")
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

// @Summary Получение информации о пользователе
// @Description Получает информацию о пользователе по его уникальному идентификатору.
// @Tags Users
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {object} models.UserData "Успешный ответ с данными пользователя"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 422 {string} string "Ошибка конвертирования ID"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /user/{userID} [get]
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
