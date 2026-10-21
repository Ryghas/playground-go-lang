package handler

import (
	"time"

	"pokemon-service/internal/domain/entity"
	"pokemon-service/internal/domain/repository"
	domainUC "pokemon-service/internal/domain/usecase"
)

// CreatePokemonRequest — body untuk POST /pokemons.
type CreatePokemonRequest struct {
	PokedexNumber int     `json:"pokedex_number" binding:"required,min=1" example:"25"`
	Name          string  `json:"name" binding:"required,min=1,max=50" example:"Pikachu"`
	PrimaryType   string  `json:"primary_type" binding:"required" example:"electric" enums:"normal,fire,water,grass,electric,ice,fighting,poison,ground,flying,psychic,bug,rock,ghost,dragon,dark,steel,fairy"`
	SecondaryType *string `json:"secondary_type" example:"fairy" enums:"normal,fire,water,grass,electric,ice,fighting,poison,ground,flying,psychic,bug,rock,ghost,dragon,dark,steel,fairy"`
	HP            int     `json:"hp" binding:"required,min=1,max=255" example:"35" minimum:"1" maximum:"255"`
	Attack        int     `json:"attack" binding:"required,min=1,max=255" example:"55" minimum:"1" maximum:"255"`
	Defense       int     `json:"defense" binding:"required,min=1,max=255" example:"40" minimum:"1" maximum:"255"`
	Speed         int     `json:"speed" binding:"required,min=1,max=255" example:"90" minimum:"1" maximum:"255"`
}

func (r CreatePokemonRequest) toUseCase() domainUC.CreatePokemonRequest {
	req := domainUC.CreatePokemonRequest{
		PokedexNumber: r.PokedexNumber,
		Name:          r.Name,
		PrimaryType:   entity.PokemonType(r.PrimaryType),
		HP:            r.HP,
		Attack:        r.Attack,
		Defense:       r.Defense,
		Speed:         r.Speed,
	}
	if r.SecondaryType != nil && *r.SecondaryType != "" {
		t := entity.PokemonType(*r.SecondaryType)
		req.SecondaryType = &t
	}
	return req
}

// UpdatePokemonRequest — semua field opsional (PATCH).
// Kirim "secondary_type": "" untuk menghapus secondary type.
type UpdatePokemonRequest struct {
	PokedexNumber *int    `json:"pokedex_number" binding:"omitempty,min=1" example:"25"`
	Name          *string `json:"name" binding:"omitempty,min=1,max=50" example:"pikachu"`
	PrimaryType   *string `json:"primary_type" example:"electric"`
	SecondaryType *string `json:"secondary_type" example:"" extensions:"x-nullable"`
	HP            *int    `json:"hp" binding:"omitempty,min=1,max=255" example:"40" minimum:"1" maximum:"255"`
	Attack        *int    `json:"attack" binding:"omitempty,min=1,max=255" example:"60" minimum:"1" maximum:"255"`
	Defense       *int    `json:"defense" binding:"omitempty,min=1,max=255" example:"45" minimum:"1" maximum:"255"`
	Speed         *int    `json:"speed" binding:"omitempty,min=1,max=255" example:"100" minimum:"1" maximum:"255"`
}

func (r UpdatePokemonRequest) toUseCase() domainUC.UpdatePokemonRequest {
	req := domainUC.UpdatePokemonRequest{
		PokedexNumber: r.PokedexNumber,
		Name:          r.Name,
		HP:            r.HP,
		Attack:        r.Attack,
		Defense:       r.Defense,
		Speed:         r.Speed,
	}
	if r.PrimaryType != nil {
		t := entity.PokemonType(*r.PrimaryType)
		req.PrimaryType = &t
	}
	if r.SecondaryType != nil {
		if *r.SecondaryType == "" {
			req.ClearSecondaryType = true
		} else {
			t := entity.PokemonType(*r.SecondaryType)
			req.SecondaryType = &t
		}
	}
	return req
}

// PokemonResponse — representasi pokemon di response HTTP.
// Dipisah dari entity agar tag dokumentasi tidak bocor ke domain layer.
type PokemonResponse struct {
	ID            string    `json:"id" example:"3f1c2b7e-8a4d-4c39-9f0e-2d5b6a7c8e91"`
	PokedexNumber int       `json:"pokedex_number" example:"25"`
	Name          string    `json:"name" example:"pikachu"`
	PrimaryType   string    `json:"primary_type" example:"electric"`
	SecondaryType *string   `json:"secondary_type" example:"fairy" extensions:"x-nullable"`
	HP            int       `json:"hp" example:"35"`
	Attack        int       `json:"attack" example:"55"`
	Defense       int       `json:"defense" example:"40"`
	Speed         int       `json:"speed" example:"90"`
	CreatedAt     time.Time `json:"created_at" example:"2026-10-21T03:47:54Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2026-10-21T03:47:54Z"`
}

func newPokemonResponse(p *entity.Pokemon) PokemonResponse {
	var secondary *string
	if p.SecondaryType != nil {
		s := string(*p.SecondaryType)
		secondary = &s
	}
	return PokemonResponse{
		ID:            p.ID,
		PokedexNumber: p.PokedexNumber,
		Name:          p.Name,
		PrimaryType:   string(p.PrimaryType),
		SecondaryType: secondary,
		HP:            p.HP,
		Attack:        p.Attack,
		Defense:       p.Defense,
		Speed:         p.Speed,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func newPokemonResponses(pokemons []*entity.Pokemon) []PokemonResponse {
	out := make([]PokemonResponse, 0, len(pokemons))
	for _, p := range pokemons {
		out = append(out, newPokemonResponse(p))
	}
	return out
}

type listPokemonQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Type     string `form:"type"`
	Search   string `form:"search"`
}

func (q listPokemonQuery) toFilter() repository.PokemonFilter {
	filter := repository.PokemonFilter{
		Page:     q.Page,
		PageSize: q.PageSize,
		Search:   q.Search,
	}
	if q.Type != "" {
		t := entity.PokemonType(q.Type)
		filter.Type = &t
	}
	return filter
}
