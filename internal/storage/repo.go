package storage

import (
	"context"
	"messaggio/internal/models"
)

type RepositoryDB interface {
	Create(ctx context.Context, message *models.Message) (int, error)
	Read(ctx context.Context, id int) (*models.Message, error)
	Update(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
	GetStats(ctx context.Context) (*models.Stats, error)
	Close() error
}
