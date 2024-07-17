package psg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"messaggio/internal/config"
	"messaggio/internal/models"
	"sync"
)

type PostgresStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewPostgresStorage(conf *config.Config) (*PostgresStorage, error) {

	login := conf.DB_LOGIN
	password := conf.DB_PASS
	host := conf.DB_HOST
	port := conf.DB_PORT
	dbname := conf.DB_NAME

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", login, password, host, port, dbname)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		dbname, driver)
	if err != nil {
		return nil, err
	}

	//применение миграций
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Create(ctx context.Context, message *models.Message) (int, error) {
	var id int
	query := `INSERT INTO messages (content, status) VALUES ($1, $2) RETURNING id`
	err := s.db.QueryRowContext(ctx, query, message.Content, message.Status).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *PostgresStorage) Read(ctx context.Context, id int) (*models.Message, error) {
	var message models.Message
	query := `SELECT id, content, status FROM messages WHERE id=$1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&message.ID, &message.Content, &message.Status)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *PostgresStorage) Update(ctx context.Context, id int) error {
	query := `UPDATE messages SET status=$1 WHERE id=$2`
	_, err := s.db.ExecContext(ctx, query, "processed", id)
	return err
}

func (s *PostgresStorage) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM messages WHERE id=$1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *PostgresStorage) GetStats(ctx context.Context) (*models.Stats, error) {
	var stats models.Stats
	query := `
        SELECT 
            (SELECT COUNT(*) FROM messages WHERE status='pending') AS pending,
            (SELECT COUNT(*) FROM messages WHERE status='processed') AS processed,
            (SELECT COUNT(*) FROM messages) AS total`
	err := s.db.QueryRowContext(ctx, query).Scan(&stats.Pending, &stats.Processed, &stats.Total)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (s *PostgresStorage) Close() error {
	if s.db != nil {
		s.db.Close()
	}
	return nil
}
