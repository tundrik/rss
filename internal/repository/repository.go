package repository

import (
	"context"
	"errors"

	"rss/configs"
	"rss/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const (
	UniqueConstrintViolation     = "23505"
	ViolatesForeignKeyConstraint = "23503"
)

var (
	ErrFeedExists      = errors.New("feed already exists")
	ErrArticleNotFound = errors.New("article not found")
	ErrReSubscription  = errors.New("re subscription")
	ErrNoSubscription  = errors.New("no found feed_pk")
)

type Repo struct {
	db  *pgxpool.Pool
	log zerolog.Logger
}

// New Инициализирует репозиторий.
func New(ctx context.Context, cfg config.Config, log zerolog.Logger) (*Repo, error) {
	db, err := pgxpool.New(ctx, cfg.PgString)
	if err != nil {
		return nil, err
	}

	repo := &Repo{
		db:  db,
		log: log,
	}
	return repo, nil
}

// Available возвращает список доступных RSS каналов.
func (r *Repo) Available(ctx context.Context) ([]entity.Feed, error) {
	const sql = `SELECT pk, feed_url FROM feed ORDER BY pk DESC;`

	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []entity.Feed
	for rows.Next() {
		var item entity.Feed
		if err := rows.Scan(&item.Pk, &item.FeedUrl); err != nil {
			return nil, err
		}
		entities = append(entities, item)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return entities, nil
}

// AddFeed добавляет новый RSS источник.
func (r *Repo) AddFeed(ctx context.Context, feedUrl string) error {
	const sql = `INSERT INTO feed(feed_url) VALUES ($1);`

	_, err := r.db.Exec(ctx, sql, feedUrl)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UniqueConstrintViolation {
			return ErrFeedExists
		}
		return err
	}
	return nil
}

// Subscribe подписывает пользователя на RSS канал.
func (r *Repo) Subscribe(ctx context.Context, personPk string, feedPk string) error {
	const sql = `INSERT INTO subscribe(person_pk, feed_pk) VALUES ($1, $2);`

	_, err := r.db.Exec(ctx, sql, personPk, feedPk)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case UniqueConstrintViolation:
				return ErrReSubscription
			case ViolatesForeignKeyConstraint:
				return ErrNoSubscription
			}
		}
		return err
	}

	return nil
}

// Unsubscribe отписывает пользователя от RSS канала.
func (r *Repo) Unsubscribe(ctx context.Context, personPk string, feedPk string) error {
	const sql = `DELETE FROM subscribe WHERE person_pk = $1 AND feed_pk = $2;`

	_, err := r.db.Exec(ctx, sql, personPk, feedPk)
	if err != nil {
		return err
	}
	return nil
}

// Article возвращает список статей для пользователя.
func (r *Repo) Article(ctx context.Context, personPk string) ([]entity.Article, error) {
	const sql = `SELECT pk, title, content, source_url, published, article.feed_pk FROM article 
	JOIN subscribe as sub ON sub.feed_pk = article.feed_pk AND sub.person_pk = $1 WHERE recorded > (SELECT viewed FROM person WHERE person.pk = $1);`

	rows, err := r.db.Query(ctx, sql, personPk)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []entity.Article
	for rows.Next() {
		var a entity.Article
		rows.Scan(&a.Pk, &a.Title, &a.Content, &a.SourceUrl, &a.Published, &a.FeedPk)
		entities = append(entities, a)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(entities) == 0 {
		return nil, ErrArticleNotFound
	}
	return entities, nil
}

// AddArticle добавляет пакет статей.
func (r *Repo) AddArticle(ctx context.Context, batch []entity.Article) {
	pgBatch := &pgx.Batch{}
	const sql = `INSERT INTO article (title, content, source_url, published, feed_pk) VALUES ($1, $2, $3, $4, $5) 
	ON CONFLICT (source_url) DO UPDATE SET (title, content, published) = (EXCLUDED.title, EXCLUDED.content, EXCLUDED.published) 
	WHERE article.published < EXCLUDED.published;`

	for _, a := range batch {
		pgBatch.Queue(sql, a.Title, a.Content, a.SourceUrl, a.Published, a.FeedPk)
	}

	results := r.db.SendBatch(ctx, pgBatch)
	defer results.Close()

	for _, item := range batch {
		_, err := results.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				r.log.Err(err).Msg("pg error")
			}
			r.log.Err(err).Str("url", item.SourceUrl).Msg("db error")
		}
	}
}

// Viewed обновляет дату последнего просмотра у пользователя.
func (r *Repo) Viewed(ctx context.Context, personPk string) error {
	const sql = `UPDATE person SET viewed = now() WHERE person.pk = $1;`

	_, err := r.db.Exec(ctx, sql, personPk)
	if err != nil {
		return err
	}
	return nil
}
