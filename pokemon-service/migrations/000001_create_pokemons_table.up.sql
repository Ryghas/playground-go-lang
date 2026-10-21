CREATE TABLE IF NOT EXISTS pokemons (
    id              UUID PRIMARY KEY,
    pokedex_number  INTEGER      NOT NULL CHECK (pokedex_number > 0),
    name            VARCHAR(50)  NOT NULL,
    primary_type    VARCHAR(20)  NOT NULL,
    secondary_type  VARCHAR(20),
    hp              INTEGER      NOT NULL CHECK (hp BETWEEN 1 AND 255),
    attack          INTEGER      NOT NULL CHECK (attack BETWEEN 1 AND 255),
    defense         INTEGER      NOT NULL CHECK (defense BETWEEN 1 AND 255),
    speed           INTEGER      NOT NULL CHECK (speed BETWEEN 1 AND 255),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Partial unique index: data yang sudah di-soft-delete tidak menghalangi insert ulang.
CREATE UNIQUE INDEX IF NOT EXISTS uq_pokemons_pokedex_number ON pokemons (pokedex_number) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_pokemons_name ON pokemons (name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pokemons_deleted_at ON pokemons (deleted_at);
