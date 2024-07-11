package usecase

import (
	"context"
    
	"rss/internal/entity"
)


type Repository interface {
    Available(ctx context.Context) ([]entity.Feed, error)
    AddFeed(ctx context.Context, feedUrl string) error
    Subscribe(ctx context.Context, personPk string, feedPk string) error
    Unsubscribe(ctx context.Context, personPk string, feedPk string) error
    Article(ctx context.Context, personPk string) ([]entity.Article, error)
    Viewed(ctx context.Context, personPk string) error
}

type UseCase struct {
    repo Repository
}

func New(r Repository) *UseCase{
    return &UseCase{
        repo: r,
    }
}

// Available возвращает список доступных RSS каналов.
func (uc *UseCase) Available(ctx context.Context) ([]entity.Feed, error) {
    return uc.repo.Available(ctx)
}

// AddFeed добавляет новый RSS источник.
func (uc *UseCase) AddFeed(ctx context.Context, feedUrl string) error {
    return uc.repo.AddFeed(ctx, feedUrl)
}

// Subscribe подписывает пользователя на RSS канал.
func (uc *UseCase) Subscribe(ctx context.Context, personPk string, feedPk string) error {
    return uc.repo.Subscribe(ctx, personPk, feedPk)
}

// Unsubscribe отписывает пользователя на RSS канал.
func (uc *UseCase) Unsubscribe(ctx context.Context, personPk string, feedPk string) error {
    return uc.repo.Unsubscribe(ctx, personPk, feedPk)
}

// Article возвращает список статей для пользователя
// и обновляет дату последнего просмотра у пользователя.
func (uc *UseCase) Article(ctx context.Context, personPk string) ([]entity.Article, error) {
    entities, err := uc.repo.Article(ctx, personPk)
    if err != nil {
		return nil, err
	}
    // обновляет дату последнего просмотра новостей пользователем
    // на практике это должно инициироваться с фронтенда 
    // после фактического просмотра пользователем.
    if err := uc.repo.Viewed(ctx, personPk); err != nil {
		return nil, err
	}
    return entities, nil
}