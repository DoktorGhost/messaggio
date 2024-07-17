package server

import (
	"context"
	"github.com/joho/godotenv"
	"messaggio/internal/broker/kafka"
	"messaggio/internal/config"
	"messaggio/internal/handlers"
	"messaggio/internal/logger"
	"messaggio/internal/storage/psg"
	"messaggio/internal/usecase"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func StartServer() error {
	messageIDChannel := make(chan string)
	// Инициализация логгера
	if err := logger.InitLogger("logs/log.txt"); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	//считываем файл .env
	err := godotenv.Load(".env")
	if err != nil {
		logger.SugaredLogger().Errorw("Ошибка загрузки файла .env", "error", err)
	}
	//парсим переменные окружения
	conf, err := config.ParseConfigServer()
	if err != nil {
		logger.SugaredLogger().Errorw("Ошибка считывания переменных окружения", "error", err)
		return err
	}

	logger.SugaredLogger().Infow("Старт сервера", "addr", conf.SERVER_HOST+":"+conf.SERVER_PORT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//подключение к БД
	db, err := psg.NewPostgresStorage(conf)
	if err != nil {
		logger.SugaredLogger().Fatalw("Ошибка при подключении к БД", "error", err)
		return err
	}
	defer db.Close()

	br, err := kafka.NewKafkaBroker(conf)
	if err != nil {
		logger.SugaredLogger().Fatalw("Ошибка при инициализации брокера Kafka", "error", err)
		return err
	}
	defer br.Close()

	useCase := usecase.NewUseCase(db, br)
	logger.SugaredLogger().Infow("Успешное подключение к БД")
	logger.SugaredLogger().Infow("Успешная инициализация брокера Kafka")

	r := handlers.InitRoutes(useCase, ctx, conf)

	// Запуск HTTP-сервера
	go func() {
		err := http.ListenAndServe(conf.SERVER_HOST+":"+conf.SERVER_PORT, r)
		if err != nil {
			logger.SugaredLogger().Errorw("Ошибка при запуске HTTP-сервера", "error", err)
		}
	}()

	// Запуск получения сообщений из Kafka
	go func() {
		err := useCase.UseCaseConsumeResponses(ctx, messageIDChannel)
		if err != nil {
			logger.SugaredLogger().Errorw("Ошибка при получении сообщений из Kafka", "error", err)
			return
		}
	}()

	go func() {
		defer close(messageIDChannel)
		for {
			select {
			case messageIDstr, ok := <-messageIDChannel:
				if !ok {
					// Канал закрыт, завершаем горутину
					return
				}
				// Обновляем статус сообщения в БД
				messageID, err := strconv.Atoi(messageIDstr)
				if err != nil {
					return
				}
				err = useCase.UseCaseUpdate(ctx, messageID)
				if err != nil {
					logger.SugaredLogger().Errorw("Ошибка при обновлении статуса сообщения в БД", "error", err)
				}
			case <-ctx.Done():
				logger.SugaredLogger().Info("Остановка обработки сообщений из канала")
				return
			}
		}
	}()

	// Ожидание сигналов завершения программы
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// Завершение работы сервера
	logger.SugaredLogger().Info("Принят сигнал завершения, закрытие сервера...")

	return nil
}
