package crawly

import (
	"time"
	"context"

	"rss/internal/entity"
	"rss/internal/repository"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
)


type Crawly struct {
	parser *gofeed.Parser
	repo   *repository.Repo
	log    zerolog.Logger
}

func New(repo *repository.Repo, log zerolog.Logger) *Crawly {
	return &Crawly{
		parser: gofeed.NewParser(),
		repo: repo,
		log: log,
	}
}

func (c *Crawly) Run() {
	ch := make(chan entity.Article, 1)

	go c.keeper(ch, 10000 * time.Millisecond)
	go c.cumulative(ch, 300, time.Duration(200))
}

func (c *Crawly) keeper(ch chan<- entity.Article, delay time.Duration) {
	ticker := time.NewTicker(10 * time.Millisecond)

	for {
		<-ticker.C
		c.log.Info().Msg("ticker")
		feeds, err := c.repo.Feed()
		if err != nil {
			c.log.Err(err).Msg("err")
		}
		for _, source := range feeds {
			c.log.Info().Str("source", source.FeedUrl).Msg("go requester")
			go c.requester(ch, source)
		}

		ticker.Reset(delay)
	}
}

func (c *Crawly) requester(ch chan<- entity.Article, source entity.Feed) {
	ctx, cancel := context.WithTimeout(context.Background(), 5000 * time.Millisecond)
    defer cancel()

	feed, err := c.parser.ParseURLWithContext(source.FeedUrl, ctx)

	if err != nil {
		c.log.Err(err).Msg("err")
		return
	}

	c.log.Info().Str("type", feed.FeedType).Msg("get")

	for _, item := range feed.Items {
		article := entity.Article{
			Title: item.Title,
			SourceUrl: item.Link,
			FeedPk: source.Pk,
		}
		if item.UpdatedParsed != nil {
			article.Updated = *item.UpdatedParsed
		} else {
			article.Updated = *item.PublishedParsed
		}
		if item.Content != "" {
			article.Content = item.Content
		} else {
			article.Content = item.Description
		}
		ch <- article
	}
}

func (c *Crawly) cumulative(ch <-chan entity.Article, size int, d time.Duration) {
	batch := make([]entity.Article, 0, size)

	flush := func() {
		c.log.Info().Int("len batch", len(batch)).Msg("flush")
		c.repo.AddArticle(batch)
		batch = batch[:0]
	}

	ticker := time.NewTicker(d * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				flush()
			}

		case article := <-ch:
			batch = append(batch, article)

			if len(batch) == size {
				flush()
				ticker.Reset(d * time.Millisecond)
			}
		}
	}
}