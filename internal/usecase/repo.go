package usecase

import (
	"context"
	"messaggio/internal/models"
)

type UseCaseStorage interface {
	UseCaseCreate(ctx context.Context, message *models.Message) (int, error)
	UseCaseRead(ctx context.Context, id int) (*models.Message, error)
	UseCaseUpdate(ctx context.Context, id int) error
	UseCaseDelete(ctx context.Context, id int) error
	UseCaseGetStats(ctx context.Context) (*models.Stats, error)
	UseCaseDBClose() error
	UseCaseSendMessage(message models.Message) error
	UseCaseConsumeResponses(ctx context.Context, output chan<- string) error
	UseCaseBrokerClose()
}
