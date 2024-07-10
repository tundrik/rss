package crawly

import (
	"context"
	"time"

	"rss/configs"
	"rss/internal/entity"
	"rss/internal/repository"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
)

const startKeeperDelay = 5 * time.Second

type Crawly struct {
	parser *gofeed.Parser
	repo   *repository.Repo
	cfg    config.Config
	log    zerolog.Logger
}

func New(repo *repository.Repo, cfg config.Config, log zerolog.Logger) *Crawly {
	return &Crawly{
		parser: gofeed.NewParser(),
		repo: repo,
		cfg: cfg,
		log: log,
	}
}

func (c *Crawly) Run() {
	itemsCh := make(chan entity.Article, 1)

	go c.keeper(itemsCh)
	go c.cumulative(itemsCh)
}

// keeper переодически получает список rss источников
// на каждый источник запускает горутину
func (c *Crawly) keeper(itemsCh chan<- entity.Article) {
	sem := newSemaphore(c.cfg.ConnLimit)

	ticker := time.NewTicker(startKeeperDelay)
    defer ticker.Stop()
	for {
		<-ticker.C
		c.log.Debug().Msg("ticker get feeds db")
		ticker.Reset(c.cfg.KeeperDelay)

		feeds, err := c.repo.Feed()
		if err != nil {
			c.log.Err(err).Msg("repo feed")
		}

		for _, source := range feeds {
			sem.Acquire()

			go func() {
				c.requester(itemsCh, source)
				sem.Release()
			}()
		}
	}
}

// requester делает get запрос, каждый item из ответа пишет в канал
func (c *Crawly) requester(itemsCh chan<- entity.Article, source entity.Feed) {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.ReqTimeout)
	defer cancel()

	feed, err := c.parser.ParseURLWithContext(source.FeedUrl, ctx)
	if err != nil {
		c.log.Err(err).Str("url", source.FeedUrl).Msg("gofeed get")
		return
	}

	for _, item := range feed.Items {
		article := entity.Article{
			Title: item.Title,
			SourceUrl: item.Link,
			FeedPk: source.Pk,
		}
		// нам нужна последняя дата
		if item.UpdatedParsed != nil {
			article.Published = *item.UpdatedParsed
		} else {
			article.Published = *item.PublishedParsed
		}
		// контента может не быть 
		if item.Content != "" {
			article.Content = item.Content
		} else {
			article.Content = item.Description
		}

		itemsCh <- article
	}
}

// cumulative накаплевает Article к себе,
// при накоплении до лимита или по дедлайну сливает в базу данных.
func (c *Crawly) cumulative(itemsCh <-chan entity.Article) {
	batch := make([]entity.Article, 0, c.cfg.CumLimit)

	flush := func() {
		c.log.Debug().Int("len batch", len(batch)).Msg("flush")
		c.repo.AddArticle(batch)
		// batch на переиспользование
		batch = batch[:0]
	}

	ticker := time.NewTicker(c.cfg.CumDeadline)

	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				// flush по дедлайну
				flush()
			}

		case article, ok := <-itemsCh:
			if !ok { 
				// flush если канал закрыли
				flush()
				return
			}

			batch = append(batch, article)

			if len(batch) == c.cfg.CumLimit {
				// flush по лимиту размера
				flush()
				ticker.Reset(c.cfg.CumDeadline)
			}
		}
	}
}