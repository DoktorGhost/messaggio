package broker

import (
	"context"
	"messaggio/internal/models"
)

type RepositoryBroker interface {
	SendMessage(message models.Message) error
	ConsumeResponses(ctx context.Context, output chan<- string) error
	Close()
}
