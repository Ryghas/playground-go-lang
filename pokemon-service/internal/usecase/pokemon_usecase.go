package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pokemon-service/internal/domain/entity"
	"pokemon-service/internal/domain/repository"
	domainUC "pokemon-service/internal/domain/usecase"
	apperrors "pokemon-service/pkg/errors"
	"pokemon-service/pkg/pagination"

	"github.com/google/uuid"
)

type pokemonUseCase struct {
	pokemonRepo repository.PokemonRepository
}

func NewPokemonUseCase(pokemonRepo repository.PokemonRepository) domainUC.PokemonUseCase {
	return &pokemonUseCase{pokemonRepo: pokemonRepo}
}

func (uc *pokemonUseCase) GetByID(ctx context.Context, id string) (*entity.Pokemon, error) {
	pokemon, err := uc.pokemonRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("pokemonUseCase.GetByID: %w", err)
	}
	return pokemon, nil
}

func (uc *pokemonUseCase) GetAll(ctx context.Context, filter repository.PokemonFilter) ([]*entity.Pokemon, int64, error) {
	filter.Page, filter.PageSize = pagination.Normalize(filter.Page, filter.PageSize)
	filter.Search = strings.TrimSpace(filter.Search)

	if filter.Type != nil && !filter.Type.IsValid() {
		return nil, 0, apperrors.NewBadRequest(fmt.Sprintf("type %q tidak valid", *filter.Type))
	}

	pokemons, total, err := uc.pokemonRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("pokemonUseCase.GetAll: %w", err)
	}
	return pokemons, total, nil
}

func (uc *pokemonUseCase) Create(ctx context.Context, req domainUC.CreatePokemonRequest) (*entity.Pokemon, error) {
	now := time.Now().UTC()
	pokemon := &entity.Pokemon{
		ID:            uuid.NewString(),
		PokedexNumber: req.PokedexNumber,
		Name:          normalizeName(req.Name),
		PrimaryType:   req.PrimaryType,
		SecondaryType: req.SecondaryType,
		HP:            req.HP,
		Attack:        req.Attack,
		Defense:       req.Defense,
		Speed:         req.Speed,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := validatePokemon(pokemon); err != nil {
		return nil, err
	}
	if err := uc.ensureUnique(ctx, pokemon); err != nil {
		return nil, fmt.Errorf("pokemonUseCase.Create: %w", err)
	}

	if err := uc.pokemonRepo.Create(ctx, pokemon); err != nil {
		return nil, fmt.Errorf("pokemonUseCase.Create: %w", err)
	}
	return pokemon, nil
}

func (uc *pokemonUseCase) Update(ctx context.Context, id string, req domainUC.UpdatePokemonRequest) (*entity.Pokemon, error) {
	pokemon, err := uc.pokemonRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("pokemonUseCase.Update: %w", err)
	}

	if req.PokedexNumber != nil {
		pokemon.PokedexNumber = *req.PokedexNumber
	}
	if req.Name != nil {
		pokemon.Name = normalizeName(*req.Name)
	}
	if req.PrimaryType != nil {
		pokemon.PrimaryType = *req.PrimaryType
	}
	if req.ClearSecondaryType {
		pokemon.SecondaryType = nil
	} else if req.SecondaryType != nil {
		pokemon.SecondaryType = req.SecondaryType
	}
	if req.HP != nil {
		pokemon.HP = *req.HP
	}
	if req.Attack != nil {
		pokemon.Attack = *req.Attack
	}
	if req.Defense != nil {
		pokemon.Defense = *req.Defense
	}
	if req.Speed != nil {
		pokemon.Speed = *req.Speed
	}
	pokemon.UpdatedAt = time.Now().UTC()

	if err := validatePokemon(pokemon); err != nil {
		return nil, err
	}
	if err := uc.ensureUnique(ctx, pokemon); err != nil {
		return nil, fmt.Errorf("pokemonUseCase.Update: %w", err)
	}

	if err := uc.pokemonRepo.Update(ctx, pokemon); err != nil {
		return nil, fmt.Errorf("pokemonUseCase.Update: %w", err)
	}
	return pokemon, nil
}

func (uc *pokemonUseCase) Delete(ctx context.Context, id string) error {
	if err := uc.pokemonRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("pokemonUseCase.Delete: %w", err)
	}
	return nil
}

// ensureUnique memastikan nama dan nomor pokedex belum dipakai pokemon lain.
// Unique index di DB tetap jadi pengaman terakhir untuk race condition.
func (uc *pokemonUseCase) ensureUnique(ctx context.Context, pokemon *entity.Pokemon) error {
	existing, err := uc.pokemonRepo.FindByPokedexNumber(ctx, pokemon.PokedexNumber)
	if err != nil && !apperrors.IsNotFound(err) {
		return fmt.Errorf("check pokedex number: %w", err)
	}
	if existing != nil && existing.ID != pokemon.ID {
		return apperrors.NewConflict(fmt.Sprintf("pokedex number %d sudah terdaftar", pokemon.PokedexNumber))
	}

	existing, err = uc.pokemonRepo.FindByName(ctx, pokemon.Name)
	if err != nil && !apperrors.IsNotFound(err) {
		return fmt.Errorf("check name: %w", err)
	}
	if existing != nil && existing.ID != pokemon.ID {
		return apperrors.NewConflict(fmt.Sprintf("pokemon %q sudah terdaftar", pokemon.Name))
	}
	return nil
}

func validatePokemon(p *entity.Pokemon) error {
	if p.Name == "" {
		return apperrors.NewBadRequest("name wajib diisi")
	}
	if p.PokedexNumber < 1 {
		return apperrors.NewBadRequest("pokedex_number minimal 1")
	}
	if !p.PrimaryType.IsValid() {
		return apperrors.NewBadRequest(fmt.Sprintf("primary_type %q tidak valid", p.PrimaryType))
	}
	if p.SecondaryType != nil {
		if !p.SecondaryType.IsValid() {
			return apperrors.NewBadRequest(fmt.Sprintf("secondary_type %q tidak valid", *p.SecondaryType))
		}
		if *p.SecondaryType == p.PrimaryType {
			return apperrors.NewBadRequest("secondary_type tidak boleh sama dengan primary_type")
		}
	}
	stats := []struct {
		name  string
		value int
	}{{"hp", p.HP}, {"attack", p.Attack}, {"defense", p.Defense}, {"speed", p.Speed}}
	for _, s := range stats {
		if s.value < 1 || s.value > 255 {
			return apperrors.NewBadRequest(fmt.Sprintf("%s harus di antara 1-255", s.name))
		}
	}
	return nil
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
