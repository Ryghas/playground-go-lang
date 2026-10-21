package usecase

import (
	"context"

	"pokemon-service/internal/domain/entity"
	"pokemon-service/internal/domain/repository"
)

type PokemonUseCase interface {
	GetByID(ctx context.Context, id string) (*entity.Pokemon, error)
	GetAll(ctx context.Context, filter repository.PokemonFilter) ([]*entity.Pokemon, int64, error)
	Create(ctx context.Context, req CreatePokemonRequest) (*entity.Pokemon, error)
	Update(ctx context.Context, id string, req UpdatePokemonRequest) (*entity.Pokemon, error)
	Delete(ctx context.Context, id string) error
}

type CreatePokemonRequest struct {
	PokedexNumber int
	Name          string
	PrimaryType   entity.PokemonType
	SecondaryType *entity.PokemonType
	HP            int
	Attack        int
	Defense       int
	Speed         int
}

// UpdatePokemonRequest - field nil artinya tidak diubah (partial update).
type UpdatePokemonRequest struct {
	PokedexNumber *int
	Name          *string
	PrimaryType   *entity.PokemonType
	SecondaryType *entity.PokemonType
	// ClearSecondaryType true untuk menghapus secondary type (set NULL).
	ClearSecondaryType bool
	HP                 *int
	Attack             *int
	Defense            *int
	Speed              *int
}
