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

const UniqueConstrintViolation = "23505"

type Repo struct {
	db    *pgxpool.Pool
	log   zerolog.Logger
}

// New Инициализирует репозиторий.
func New(ctx context.Context, cfg config.Config, log zerolog.Logger) (*Repo, error) {
	db, err := pgxpool.New(ctx, cfg.PgString)
	if err != nil {
		return nil, err
	}

	repo := &Repo{
		db:    db,
		log:   log,
	}
	return repo, nil
}

// Feed возвращает список доступных RSS каналов.
func (r *Repo) Feed() ([]entity.Feed, error) {
	const sql = `SELECT pk, feed_url FROM feed ORDER BY pk DESC;`

	rows, err := r.db.Query(context.Background(), sql)
	if err != nil {
		r.log.Err(err).Msg("db error")
		return nil, err
	}
	defer rows.Close()

	var entities []entity.Feed
	for rows.Next() {
		var item entity.Feed
		rows.Scan(&item.Pk, &item.FeedUrl)
		entities = append(entities, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entities, nil
}

// AddFeed добавляет новый RSS источник.
func (r *Repo) AddFeed(feedUrl string) error {
	const sql = `INSERT INTO feed(feed_url) VALUES ($1);`

	_, err := r.db.Exec(context.Background(), sql, feedUrl)
	if err != nil {
		return err
	}
	return nil
}

// Subscribe подписывает пользователя на RSS канал.
func (r *Repo) Subscribe(personPk string, feedPk string) error {
	const sql = `INSERT INTO subscribe(person_pk, feed_pk) VALUES ($1, $2);`

	_, err := r.db.Exec(context.Background(), sql, personPk, feedPk)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UniqueConstrintViolation {
			return nil
		}
		return err
	}
	return nil
}

// Unsubscribe отписывает пользователя на RSS канал.
func (r *Repo) Unsubscribe(personPk string, feedPk string) error {
	const sql = `DELETE FROM subscribe WHERE person_pk = $1 AND feed_pk = $2;`

	_, err := r.db.Exec(context.Background(), sql, personPk, feedPk)
	if err != nil {
		return err
	}
	return nil
}

// Article возвращает список статей для пользователя.
func (r *Repo) Article(personPk string) ([]entity.Article, error){
	const sql = `SELECT pk, title, content, source_url, updated, article.feed_pk FROM article JOIN subscribe as sub ON sub.feed_pk = article.feed_pk AND sub.person_pk = $1 WHERE updated > (SELECT viewed FROM person WHERE person.pk = $1);`

	rows, err := r.db.Query(context.Background(), sql, personPk)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []entity.Article
	for rows.Next() {
		var a entity.Article
		rows.Scan(&a.Pk, &a.Title, &a.Content, &a.SourceUrl, &a.Updated, &a.FeedPk)
		entities = append(entities, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.viewed(personPk); err != nil {
		return nil, err
	}
	return entities, nil
}

// AddArticle добавляет пакет статей.
func (r *Repo) AddArticle(batch []entity.Article)  {
	pgBatch := &pgx.Batch{}
	const sql = `INSERT INTO article (title, content, source_url, updated, feed_pk) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (source_url) DO UPDATE SET (title, content, updated) = (EXCLUDED.title, EXCLUDED.content, EXCLUDED.updated);`

	for _, a := range batch {
		pgBatch.Queue(sql, a.Title, a.Content, a.SourceUrl, a.Updated, a.FeedPk)
	}

	results := r.db.SendBatch(context.Background(), pgBatch)
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

func (r *Repo) viewed(personPk string) error {
	const sql = `UPDATE person SET viewed = now() WHERE person.pk = $1;`

	_, err := r.db.Exec(context.Background(), sql, personPk)
	if err != nil {
		return err
	}
	return nil
}

