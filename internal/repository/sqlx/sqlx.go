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
	"go.uber.org/zap"
)

var ErrIDAlreadyExists = errors.New("id уже занят")

type SQLXRepository struct {
	db     *sqlx.DB
	delCh  chan model.URL
	logger *zap.Logger
}

func NewRepository(dsn string, logger *zap.Logger) (*SQLXRepository, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	repo := &SQLXRepository{db: db, delCh: make(chan model.URL, 1024), logger: logger}

	if err := repo.runMigrations(); err != nil {
		return nil, fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	go repo.startDeleteQueue()

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
		SELECT short_url AS short, original_url AS original, user_id, is_deleted
		FROM urls
		WHERE short_url = $1
	`, id)
	if err != nil {
		return nil, false
	}
	if url.DeletedFlag {
		r.logger.Debug("URL удалён", zap.String("id", id), zap.String("url", url.Original))
	}

	return &url, true
}

func (r *SQLXRepository) GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error) {
	var urls []model.URL
	err := r.db.SelectContext(ctx, &urls, `
		SELECT short_url AS short, original_url AS original, user_id
		FROM urls
		WHERE user_id = $1
		AND is_deleted = FALSE
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

func (r *SQLXRepository) DeleteURLs(ctx context.Context, urls []model.URL) error {

	for _, url := range urls {
		if err := r.pushToDeleteQueue(url); err != nil {
			return fmt.Errorf("ошибка добавления url в очередь на удаление: %w", err)
		}
	}

	return nil
}

func (r *SQLXRepository) pushToDeleteQueue(url model.URL) error {
	r.delCh <- url

	return nil
}

func (r *SQLXRepository) startDeleteQueue() {
	ticker := time.NewTicker(3 * time.Second)
	var urls []model.URL

	r.logger.Info("запуск очереди на удаление URL")
	for {
		select {
		case url := <-r.delCh:
			urls = append(urls, url)
			r.logger.Debug("URL добавлен в очередь на удаление", zap.String("short_url", url.Short), zap.String("original_url", url.Original))
		case <-ticker.C:
			if len(urls) == 0 {
				continue
			}
			r.logger.Debug("ticker.C: запуск удаления URL", zap.Int("count", len(urls)))
			err := r.runDeleteQuery(context.TODO(), urls)
			if err != nil {
				r.logger.Error("ошибка при удалении url", zap.Error(err))
			}
			urls = nil
		}
	}
}

func (r *SQLXRepository) runDeleteQuery(ctx context.Context, urls []model.URL) error {

	inSql := fmt.Sprintf("(%s)", placeholders(len(urls)))

	args := make([]interface{}, len(urls))
	for i, url := range urls {
		args[i] = url.Short
	}

	query := fmt.Sprintf("UPDATE urls SET is_deleted = TRUE WHERE short_url IN %s", inSql)
	_, err := r.db.ExecContext(ctx, query, args...)

	return err
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	result := "$1"
	for i := 2; i <= n; i++ {
		result += fmt.Sprintf(", $%d", i)
	}

	return result
}
