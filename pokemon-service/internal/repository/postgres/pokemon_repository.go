package postgres

import (
	"context"
	"errors"
	"fmt"

	"pokemon-service/internal/domain/entity"
	"pokemon-service/internal/domain/repository"
	apperrors "pokemon-service/pkg/errors"
	"pokemon-service/pkg/pagination"

	"gorm.io/gorm"
)

type pokemonRepository struct {
	db *gorm.DB
}

func NewPokemonRepository(db *gorm.DB) repository.PokemonRepository {
	return &pokemonRepository{db: db}
}

func (r *pokemonRepository) FindByID(ctx context.Context, id string) (*entity.Pokemon, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *pokemonRepository) FindByName(ctx context.Context, name string) (*entity.Pokemon, error) {
	return r.findOne(ctx, "name = ?", name)
}

func (r *pokemonRepository) FindByPokedexNumber(ctx context.Context, number int) (*entity.Pokemon, error) {
	return r.findOne(ctx, "pokedex_number = ?", number)
}

func (r *pokemonRepository) FindAll(ctx context.Context, filter repository.PokemonFilter) ([]*entity.Pokemon, int64, error) {
	query := r.db.WithContext(ctx).Model(&pokemonModel{})

	if filter.Type != nil {
		query = query.Where("(primary_type = ? OR secondary_type = ?)", *filter.Type, *filter.Type)
	}
	if filter.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("pokemonRepository.FindAll count: %w", err)
	}

	var models []pokemonModel
	err := query.
		Order("pokedex_number ASC").
		Offset(pagination.Offset(filter.Page, filter.PageSize)).
		Limit(filter.PageSize).
		Find(&models).Error
	if err != nil {
		return nil, 0, fmt.Errorf("pokemonRepository.FindAll: %w", err)
	}

	pokemons := make([]*entity.Pokemon, 0, len(models))
	for i := range models {
		pokemons = append(pokemons, models[i].toEntity())
	}
	return pokemons, total, nil
}

func (r *pokemonRepository) Create(ctx context.Context, pokemon *entity.Pokemon) error {
	if err := r.db.WithContext(ctx).Create(newPokemonModel(pokemon)).Error; err != nil {
		return translateError("pokemonRepository.Create", err)
	}
	return nil
}

func (r *pokemonRepository) Update(ctx context.Context, pokemon *entity.Pokemon) error {
	m := newPokemonModel(pokemon)
	// Pakai map agar nilai NULL (secondary_type) ikut ter-update.
	res := r.db.WithContext(ctx).
		Model(&pokemonModel{}).
		Where("id = ?", pokemon.ID).
		Updates(map[string]any{
			"pokedex_number": m.PokedexNumber,
			"name":           m.Name,
			"primary_type":   m.PrimaryType,
			"secondary_type": m.SecondaryType,
			"hp":             m.HP,
			"attack":         m.Attack,
			"defense":        m.Defense,
			"speed":          m.Speed,
			"updated_at":     m.UpdatedAt,
		})
	if res.Error != nil {
		return translateError("pokemonRepository.Update", res.Error)
	}
	if res.RowsAffected == 0 {
		return apperrors.NewNotFound("pokemon tidak ditemukan")
	}
	return nil
}

// Delete melakukan soft delete (mengisi kolom deleted_at).
func (r *pokemonRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&pokemonModel{}, "id = ?", id)
	if res.Error != nil {
		return translateError("pokemonRepository.Delete", res.Error)
	}
	if res.RowsAffected == 0 {
		return apperrors.NewNotFound("pokemon tidak ditemukan")
	}
	return nil
}

func (r *pokemonRepository) findOne(ctx context.Context, query string, args ...any) (*entity.Pokemon, error) {
	var m pokemonModel
	if err := r.db.WithContext(ctx).Where(query, args...).First(&m).Error; err != nil {
		return nil, translateError("pokemonRepository.findOne", err)
	}
	return m.toEntity(), nil
}

func translateError(op string, err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return apperrors.NewNotFound("pokemon tidak ditemukan")
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return apperrors.NewConflict("pokemon dengan nama atau pokedex number tersebut sudah ada")
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
