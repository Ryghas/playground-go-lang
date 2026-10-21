package postgres

import (
	"time"

	"pokemon-service/internal/domain/entity"

	"gorm.io/gorm"
)

// pokemonModel — representasi tabel `pokemons`. Sengaja dipisah dari entity
// agar domain layer tidak bergantung pada tag/tipe GORM.
type pokemonModel struct {
	ID            string `gorm:"type:uuid;primaryKey"`
	PokedexNumber int
	Name          string
	PrimaryType   string
	SecondaryType *string
	HP            int `gorm:"column:hp"`
	Attack        int
	Defense       int
	Speed         int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (pokemonModel) TableName() string { return "pokemons" }

func (m *pokemonModel) toEntity() *entity.Pokemon {
	var secondary *entity.PokemonType
	if m.SecondaryType != nil {
		t := entity.PokemonType(*m.SecondaryType)
		secondary = &t
	}
	return &entity.Pokemon{
		ID:            m.ID,
		PokedexNumber: m.PokedexNumber,
		Name:          m.Name,
		PrimaryType:   entity.PokemonType(m.PrimaryType),
		SecondaryType: secondary,
		HP:            m.HP,
		Attack:        m.Attack,
		Defense:       m.Defense,
		Speed:         m.Speed,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func newPokemonModel(p *entity.Pokemon) *pokemonModel {
	var secondary *string
	if p.SecondaryType != nil {
		s := string(*p.SecondaryType)
		secondary = &s
	}
	return &pokemonModel{
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
