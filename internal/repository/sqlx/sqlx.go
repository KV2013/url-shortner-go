package sqlx

import (
	"errors"
	"fmt"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var ErrIDAlreadyExists = errors.New("id уже занят")

type SQLXRepository struct {
	db *sqlx.DB
}

func NewRepository(cfg *config.Config) (*SQLXRepository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbSSLMode,
	)
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	return &SQLXRepository{db: db}, nil
}

func (r *SQLXRepository) GetByID(id string) (*model.URL, bool) {
	var url model.URL
	err := r.db.Get(&url, `
		SELECT short_url AS short, original_url AS original
		FROM urls
		WHERE short_url = $1
	`, id)
	if err != nil {
		return nil, false
	}

	return &url, true
}

func (r *SQLXRepository) Save(url *model.URL) error {
	_, err := r.db.Exec(`
		INSERT INTO urls (short_url, original_url)
		VALUES ($1, $2)
	`, url.Short, url.Original)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrIDAlreadyExists
		}
		return fmt.Errorf("ошибка сохранения url: %w", err)
	}

	return nil
}

func (r *SQLXRepository) Close() error {
	return r.db.Close()
}

func (r *SQLXRepository) Ping() error {
	return r.db.Ping()
}
