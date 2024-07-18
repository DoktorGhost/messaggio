package usecase

import (
	"context"
	"messaggio/internal/broker"
	"messaggio/internal/models"
	"messaggio/internal/storage"
)

type useCase struct {
	storage storage.RepositoryDB
	broker  broker.RepositoryBroker
}

func NewUseCase(storage storage.RepositoryDB, broker broker.RepositoryBroker) UseCaseStorage {
	return &useCase{storage: storage, broker: broker}
}
func (uc *useCase) UseCaseCreate(ctx context.Context, message *models.Message) (int, error) {
	return uc.storage.Create(ctx, message)
}

func (uc *useCase) UseCaseRead(ctx context.Context, id int) (*models.Message, error) {
	return uc.storage.Read(ctx, id)
}

func (uc *useCase) UseCaseUpdate(ctx context.Context, id int) error {
	return uc.storage.Update(ctx, id)
}

func (uc *useCase) UseCaseDelete(ctx context.Context, id int) error {
	return uc.storage.Delete(ctx, id)
}

func (uc *useCase) UseCaseGetStats(ctx context.Context) (*models.Stats, error) {
	return uc.storage.GetStats(ctx)
}

func (uc *useCase) UseCaseDBClose() error {
	return uc.storage.Close()
}

func (uc *useCase) UseCaseSendMessage(message models.Message) error {
	return uc.broker.SendMessage(message)
}

func (uc *useCase) UseCaseConsumeResponses(ctx context.Context, output chan<- string) error {
	return uc.broker.ConsumeResponses(ctx, output)
}

func (uc *useCase) UseCaseBrokerClose() {
	uc.broker.Close()
}
