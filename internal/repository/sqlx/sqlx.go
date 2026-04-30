package sqlx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var ErrIDAlreadyExists = errors.New("id уже занят")

type SQLXRepository struct {
	db *sqlx.DB
	delCh chan model.URL
}

func NewRepository(dsn string) (*SQLXRepository, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	repo := &SQLXRepository{db: db, delCh: make(chan model.URL, 3)}

	if err := repo.runMigrations(); err != nil {
		return nil, fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	return repo, nil
}

func (r *SQLXRepository) runMigrations() error {
	driver, err := migratepgx.WithInstance(r.db.DB, &migratepgx.Config{})
	if err != nil {
		return fmt.Errorf("не удалось создать драйвер миграций: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("не удалось инициализировать migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("не удалось применить миграции: %w", err)
	}

	return nil
}

func (r *SQLXRepository) GetByID(ctx context.Context, id string) (*model.URL, bool) {
	var url model.URL
	err := r.db.GetContext(ctx, &url, `
		SELECT short_url AS short, original_url AS original, user_id
		FROM urls
		WHERE short_url = $1
	`, id)
	if err != nil {
		return nil, false
	}

	return &url, true
}

func (r *SQLXRepository) GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error) {
	var urls []model.URL
	err := r.db.SelectContext(ctx, &urls, `
		SELECT short_url AS short, original_url AS original, user_id
		FROM urls
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

func (r *SQLXRepository) Save(ctx context.Context, url *model.URL) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO urls (short_url, original_url, user_id)
		VALUES ($1, $2, $3)
	`, url.Short, url.Original, url.UserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "u_idx_urls_original_url" {
				var existing model.URL
				if err := r.db.GetContext(ctx, &existing, `
					SELECT short_url AS short, original_url AS original, user_id
					FROM urls WHERE original_url = $1
				`, url.Original); err != nil {
					return fmt.Errorf("ошибка поиска существующего url: %w", err)
				}
				return &model.ErrURLAlreadyExists{URL: existing}
			}
			return ErrIDAlreadyExists
		}
		return fmt.Errorf("ошибка сохранения url: %w", err)
	}

	return nil
}

func (r *SQLXRepository) SaveMany(ctx context.Context, urls []*model.URL) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		WITH ins AS (
			INSERT INTO urls (short_url, original_url, user_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (original_url) DO NOTHING
			RETURNING short_url, original_url, user_id
		)
		SELECT short_url AS short, original_url AS original, user_id, FALSE AS conflicted
		FROM ins
		UNION ALL
		SELECT short_url AS short, original_url AS original, user_id, TRUE AS conflicted
		FROM urls
		WHERE original_url = $2
		AND NOT EXISTS (
			SELECT 1 FROM ins
		)
	`)
	if err != nil {
		return fmt.Errorf("не удалось создать prepared statement: %w", err)
	}
	defer stmt.Close()

	for _, url := range urls {
		var insResult struct {
			model.URL
			Conflicted bool `db:"conflicted"`
		}
		if err := stmt.GetContext(ctx, &insResult, url.Short, url.Original, url.UserID); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation { // конфликт по short_url на всякий случай
				return ErrIDAlreadyExists
			}
			return fmt.Errorf("ошибка сохранения url: %w", err)
		}
		if insResult.Conflicted {
			return &model.ErrURLAlreadyExists{URL: insResult.URL}
		}
	}

	return tx.Commit()
}

func (r *SQLXRepository) Close() error {
	return r.db.Close()
}

func (r *SQLXRepository) Ping() error {
	return r.db.Ping()
}

func (r *SQLXRepository) DeleteURLs(ctx context.Context, ids []string) error {

	return nil
}

func (r *SQLXRepository) pushToDeleteQueue(url model.URL) {
	select {
	case r.delCh <- url:
	default:
		// если канал заполнен, просто пропускаем удаление, чтобы не блокировать основной поток
	}
}

func (r *SQLXRepository) startDeleteQueue() {
	ticker := time.NewTicker(3 * time.Second)
	var urls []model.URL

	for {
		select {
		case url := <-r.delCh:
			urls = append(urls, url)
		case <-ticker.C:
			if len(urls) == 0 {
				continue
			}
	}
}
