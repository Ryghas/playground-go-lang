package entity

import "time"

type Pokemon struct {
	ID            string       `json:"id"`
	PokedexNumber int          `json:"pokedex_number"`
	Name          string       `json:"name"`
	PrimaryType   PokemonType  `json:"primary_type"`
	SecondaryType *PokemonType `json:"secondary_type"`
	HP            int          `json:"hp"`
	Attack        int          `json:"attack"`
	Defense       int          `json:"defense"`
	Speed         int          `json:"speed"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

type PokemonType string

const (
	PokemonTypeNormal   PokemonType = "normal"
	PokemonTypeFire     PokemonType = "fire"
	PokemonTypeWater    PokemonType = "water"
	PokemonTypeGrass    PokemonType = "grass"
	PokemonTypeElectric PokemonType = "electric"
	PokemonTypeIce      PokemonType = "ice"
	PokemonTypeFighting PokemonType = "fighting"
	PokemonTypePoison   PokemonType = "poison"
	PokemonTypeGround   PokemonType = "ground"
	PokemonTypeFlying   PokemonType = "flying"
	PokemonTypePsychic  PokemonType = "psychic"
	PokemonTypeBug      PokemonType = "bug"
	PokemonTypeRock     PokemonType = "rock"
	PokemonTypeGhost    PokemonType = "ghost"
	PokemonTypeDragon   PokemonType = "dragon"
	PokemonTypeDark     PokemonType = "dark"
	PokemonTypeSteel    PokemonType = "steel"
	PokemonTypeFairy    PokemonType = "fairy"
)

var validPokemonTypes = map[PokemonType]struct{}{
	PokemonTypeNormal: {}, PokemonTypeFire: {}, PokemonTypeWater: {}, PokemonTypeGrass: {},
	PokemonTypeElectric: {}, PokemonTypeIce: {}, PokemonTypeFighting: {}, PokemonTypePoison: {},
	PokemonTypeGround: {}, PokemonTypeFlying: {}, PokemonTypePsychic: {}, PokemonTypeBug: {},
	PokemonTypeRock: {}, PokemonTypeGhost: {}, PokemonTypeDragon: {}, PokemonTypeDark: {},
	PokemonTypeSteel: {}, PokemonTypeFairy: {},
}

func (t PokemonType) IsValid() bool {
	_, ok := validPokemonTypes[t]
	return ok
}
