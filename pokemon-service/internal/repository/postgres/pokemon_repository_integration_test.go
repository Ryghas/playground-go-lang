//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pokemon-service/internal/domain/entity"
	"pokemon-service/internal/domain/repository"
	"pokemon-service/internal/repository/postgres"
	apperrors "pokemon-service/pkg/errors"

	"github.com/google/uuid"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// setupDB membuat schema baru yang terisolasi per test, menjalankan migration,
// lalu menghapus schema tersebut setelah test selesai.
//
// Butuh env TEST_DATABASE_DSN, contoh:
//
//	host=localhost port=5432 user=postgres password=postgres dbname=pokemon sslmode=disable
func setupDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN tidak di-set, integration test dilewati")
	}

	gormCfg := &gorm.Config{TranslateError: true, Logger: gormlogger.Discard}

	admin, err := gorm.Open(gormpostgres.Open(dsn), gormCfg)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	schema := fmt.Sprintf("it_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(dsn+" search_path="+schema), gormCfg)
	if err != nil {
		t.Fatalf("connect schema: %v", err)
	}

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		_ = admin.Exec("DROP SCHEMA " + schema + " CASCADE").Error
		if sqlDB, err := admin.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	migrations, err := filepath.Glob("../../../migrations/*.up.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("migration tidak ditemukan: %v", err)
	}
	for _, path := range migrations {
		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("baca %s: %v", path, err)
		}
		if err := db.Exec(string(sql)).Error; err != nil {
			t.Fatalf("jalankan %s: %v", path, err)
		}
	}
	return db
}

func newPokemon(number int, name string, primary entity.PokemonType, secondary *entity.PokemonType) *entity.Pokemon {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &entity.Pokemon{
		ID: uuid.NewString(), PokedexNumber: number, Name: name,
		PrimaryType: primary, SecondaryType: secondary,
		HP: 50, Attack: 50, Defense: 50, Speed: 50,
		CreatedAt: now, UpdatedAt: now,
	}
}

func typePtr(t entity.PokemonType) *entity.PokemonType { return &t }

func mustCreate(t *testing.T, repo repository.PokemonRepository, p *entity.Pokemon) {
	t.Helper()
	if err := repo.Create(context.Background(), p); err != nil {
		t.Fatalf("create %s: %v", p.Name, err)
	}
}

func TestPokemonRepository_CreateAndFind(t *testing.T) {
	repo := postgres.NewPokemonRepository(setupDB(t))
	ctx := context.Background()

	charizard := newPokemon(6, "charizard", entity.PokemonTypeFire, typePtr(entity.PokemonTypeFlying))
	mustCreate(t, repo, charizard)

	finders := map[string]func() (*entity.Pokemon, error){
		"FindByID":            func() (*entity.Pokemon, error) { return repo.FindByID(ctx, charizard.ID) },
		"FindByName":          func() (*entity.Pokemon, error) { return repo.FindByName(ctx, "charizard") },
		"FindByPokedexNumber": func() (*entity.Pokemon, error) { return repo.FindByPokedexNumber(ctx, 6) },
	}
	for name, find := range finders {
		t.Run(name, func(t *testing.T) {
			got, err := find()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != charizard.ID || got.Name != "charizard" || got.PrimaryType != entity.PokemonTypeFire {
				t.Errorf("got %+v", got)
			}
			if got.SecondaryType == nil || *got.SecondaryType != entity.PokemonTypeFlying {
				t.Errorf("secondary type = %v", got.SecondaryType)
			}
			if !got.CreatedAt.Equal(charizard.CreatedAt) {
				t.Errorf("created_at = %v, want %v", got.CreatedAt, charizard.CreatedAt)
			}
		})
	}
}

func TestPokemonRepository_NotFound(t *testing.T) {
	repo := postgres.NewPokemonRepository(setupDB(t))

	_, err := repo.FindByID(context.Background(), uuid.NewString())
	if !apperrors.IsNotFound(err) {
		t.Fatalf("expected NOT_FOUND, got %v", err)
	}
}

func TestPokemonRepository_UniqueConstraint(t *testing.T) {
	repo := postgres.NewPokemonRepository(setupDB(t))
	mustCreate(t, repo, newPokemon(25, "pikachu", entity.PokemonTypeElectric, nil))

	tests := []struct {
		name string
		p    *entity.Pokemon
	}{
		{"nama duplikat", newPokemon(26, "pikachu", entity.PokemonTypeElectric, nil)},
		{"pokedex number duplikat", newPokemon(25, "raichu", entity.PokemonTypeElectric, nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.p)
			if !apperrors.IsConflict(err) {
				t.Fatalf("expected CONFLICT dari unique index, got %v", err)
			}
		})
	}
}

func TestPokemonRepository_FindAll(t *testing.T) {
	repo := postgres.NewPokemonRepository(setupDB(t))
	ctx := context.Background()

	// Sengaja di-insert tidak urut untuk mengecek ORDER BY pokedex_number.
	mustCreate(t, repo, newPokemon(94, "gengar", entity.PokemonTypeGhost, typePtr(entity.PokemonTypePoison)))
	mustCreate(t, repo, newPokemon(4, "charmander", entity.PokemonTypeFire, nil))
	mustCreate(t, repo, newPokemon(1, "bulbasaur", entity.PokemonTypeGrass, typePtr(entity.PokemonTypePoison)))
	mustCreate(t, repo, newPokemon(6, "charizard", entity.PokemonTypeFire, typePtr(entity.PokemonTypeFlying)))

	tests := []struct {
		name      string
		filter    repository.PokemonFilter
		wantNames []string
		wantTotal int64
	}{
		{
			name:      "pagination halaman 1",
			filter:    repository.PokemonFilter{Page: 1, PageSize: 3},
			wantNames: []string{"bulbasaur", "charmander", "charizard"},
			wantTotal: 4,
		},
		{
			name:      "pagination halaman 2",
			filter:    repository.PokemonFilter{Page: 2, PageSize: 3},
			wantNames: []string{"gengar"},
			wantTotal: 4,
		},
		{
			name:      "filter type cocok ke secondary_type",
			filter:    repository.PokemonFilter{Page: 1, PageSize: 10, Type: typePtr(entity.PokemonTypePoison)},
			wantNames: []string{"bulbasaur", "gengar"},
			wantTotal: 2,
		},
		{
			name:      "search case-insensitive",
			filter:    repository.PokemonFilter{Page: 1, PageSize: 10, Search: "CHAR"},
			wantNames: []string{"charmander", "charizard"},
			wantTotal: 2,
		},
		{
			name:      "filter type + search digabung (AND)",
			filter:    repository.PokemonFilter{Page: 1, PageSize: 10, Type: typePtr(entity.PokemonTypeFlying), Search: "char"},
			wantNames: []string{"charizard"},
			wantTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := repo.FindAll(ctx, tt.filter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if len(got) != len(tt.wantNames) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.wantNames))
			}
			for i, p := range got {
				if p.Name != tt.wantNames[i] {
					t.Errorf("[%d] = %s, want %s", i, p.Name, tt.wantNames[i])
				}
			}
		})
	}
}

func TestPokemonRepository_Update(t *testing.T) {
	repo := postgres.NewPokemonRepository(setupDB(t))
	ctx := context.Background()

	p := newPokemon(6, "charizard", entity.PokemonTypeFire, typePtr(entity.PokemonTypeFlying))
	mustCreate(t, repo, p)

	p.Speed = 120
	p.SecondaryType = nil // harus tersimpan sebagai NULL
	p.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Speed != 120 {
		t.Errorf("speed = %d, want 120", got.Speed)
	}
	if got.SecondaryType != nil {
		t.Errorf("secondary type harus NULL, got %v", *got.SecondaryType)
	}

	missing := newPokemon(999, "missingno", entity.PokemonTypeNormal, nil)
	if err := repo.Update(ctx, missing); !apperrors.IsNotFound(err) {
		t.Fatalf("update data tidak ada harus NOT_FOUND, got %v", err)
	}
}

func TestPokemonRepository_SoftDelete(t *testing.T) {
	repo := postgres.NewPokemonRepository(setupDB(t))
	ctx := context.Background()

	p := newPokemon(25, "pikachu", entity.PokemonTypeElectric, nil)
	mustCreate(t, repo, p)

	if err := repo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := repo.FindByID(ctx, p.ID); !apperrors.IsNotFound(err) {
		t.Errorf("data terhapus tidak boleh ditemukan, got %v", err)
	}
	if _, total, _ := repo.FindAll(ctx, repository.PokemonFilter{Page: 1, PageSize: 10}); total != 0 {
		t.Errorf("data terhapus tidak boleh muncul di list, total = %d", total)
	}
	if err := repo.Delete(ctx, p.ID); !apperrors.IsNotFound(err) {
		t.Errorf("delete kedua harus NOT_FOUND, got %v", err)
	}
	if err := repo.Update(ctx, p); !apperrors.IsNotFound(err) {
		t.Errorf("update data terhapus harus NOT_FOUND, got %v", err)
	}

	// Partial unique index: nama & nomor boleh dipakai lagi setelah soft delete.
	mustCreate(t, repo, newPokemon(25, "pikachu", entity.PokemonTypeElectric, nil))
}
