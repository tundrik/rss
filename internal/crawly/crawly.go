package crawly
/*
	Концепция
	1) keeper переодически получает список rss источников из базы данных. 
	2) множество горутин ограниченное семафором, по горутине на каждый url.
	3) cumulative накаплевает Article к себе, при накоплении до лимита или по дедлайну сливает в базу данных.
*/
import (
	"context"
	"time"

	"rss/configs"
	"rss/internal/entity"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
)


const startKeeperDelay = 5 * time.Second


type Repository interface {
    Available(ctx context.Context) ([]entity.Feed, error)
    AddArticle(ctx context.Context, batch []entity.Article)
}

type Crawly struct {
	parser *gofeed.Parser
	repo   Repository
	cfg    config.CrawlyConfig
	log    zerolog.Logger
}

func New(repo Repository, cfg config.CrawlyConfig, log zerolog.Logger) *Crawly {
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
	ctx := context.TODO()
	sem := newSemaphore(c.cfg.ConnLimit)

	ticker := time.NewTicker(startKeeperDelay)
	defer ticker.Stop()
	for {
		<-ticker.C
		c.log.Debug().Msg("ticker get feeds db")
		ticker.Reset(c.cfg.KeeperDelay)

		feeds, err := c.repo.Available(ctx)
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
		c.repo.AddArticle(context.TODO(), batch)
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