package repository

import (
	"context"

	"pokemon-service/internal/domain/entity"
)

type PokemonRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Pokemon, error)
	FindByName(ctx context.Context, name string) (*entity.Pokemon, error)
	FindByPokedexNumber(ctx context.Context, number int) (*entity.Pokemon, error)
	FindAll(ctx context.Context, filter PokemonFilter) ([]*entity.Pokemon, int64, error)
	Create(ctx context.Context, pokemon *entity.Pokemon) error
	Update(ctx context.Context, pokemon *entity.Pokemon) error
	Delete(ctx context.Context, id string) error
}

type PokemonFilter struct {
	Page     int
	PageSize int
	Type     *entity.PokemonType
	Search   string
}
