package usecase_test

import (
	"context"
	"errors"
	"testing"

	"pokemon-service/internal/domain/entity"
	"pokemon-service/internal/domain/repository"
	domainUC "pokemon-service/internal/domain/usecase"
	"pokemon-service/internal/usecase"
	apperrors "pokemon-service/pkg/errors"
)

// fakePokemonRepo - implementasi in-memory dari repository.PokemonRepository.
type fakePokemonRepo struct {
	data       map[string]*entity.Pokemon
	lastFilter repository.PokemonFilter
	findAllErr error
}

func newFakeRepo(seed ...*entity.Pokemon) *fakePokemonRepo {
	r := &fakePokemonRepo{data: map[string]*entity.Pokemon{}}
	for _, p := range seed {
		r.data[p.ID] = p
	}
	return r
}

func (r *fakePokemonRepo) FindByID(_ context.Context, id string) (*entity.Pokemon, error) {
	if p, ok := r.data[id]; ok {
		cp := *p
		return &cp, nil
	}
	return nil, apperrors.NewNotFound("pokemon tidak ditemukan")
}

func (r *fakePokemonRepo) FindByName(_ context.Context, name string) (*entity.Pokemon, error) {
	for _, p := range r.data {
		if p.Name == name {
			cp := *p
			return &cp, nil
		}
	}
	return nil, apperrors.NewNotFound("pokemon tidak ditemukan")
}

func (r *fakePokemonRepo) FindByPokedexNumber(_ context.Context, number int) (*entity.Pokemon, error) {
	for _, p := range r.data {
		if p.PokedexNumber == number {
			cp := *p
			return &cp, nil
		}
	}
	return nil, apperrors.NewNotFound("pokemon tidak ditemukan")
}

func (r *fakePokemonRepo) FindAll(_ context.Context, filter repository.PokemonFilter) ([]*entity.Pokemon, int64, error) {
	r.lastFilter = filter
	if r.findAllErr != nil {
		return nil, 0, r.findAllErr
	}
	out := make([]*entity.Pokemon, 0, len(r.data))
	for _, p := range r.data {
		out = append(out, p)
	}
	return out, int64(len(out)), nil
}

func (r *fakePokemonRepo) Create(_ context.Context, p *entity.Pokemon) error {
	r.data[p.ID] = p
	return nil
}

func (r *fakePokemonRepo) Update(_ context.Context, p *entity.Pokemon) error {
	if _, ok := r.data[p.ID]; !ok {
		return apperrors.NewNotFound("pokemon tidak ditemukan")
	}
	r.data[p.ID] = p
	return nil
}

func (r *fakePokemonRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.data[id]; !ok {
		return apperrors.NewNotFound("pokemon tidak ditemukan")
	}
	delete(r.data, id)
	return nil
}

func typePtr(t entity.PokemonType) *entity.PokemonType { return &t }
func intPtr(v int) *int                                { return &v }
func strPtr(v string) *string                          { return &v }

func pikachu() *entity.Pokemon {
	return &entity.Pokemon{
		ID: "11111111-1111-1111-1111-111111111111", PokedexNumber: 25, Name: "pikachu",
		PrimaryType: entity.PokemonTypeElectric, HP: 35, Attack: 55, Defense: 40, Speed: 90,
	}
}

func validCreateReq() domainUC.CreatePokemonRequest {
	return domainUC.CreatePokemonRequest{
		PokedexNumber: 6, Name: "  Charizard ", PrimaryType: entity.PokemonTypeFire,
		SecondaryType: typePtr(entity.PokemonTypeFlying), HP: 78, Attack: 84, Defense: 78, Speed: 100,
	}
}

func TestCreate_Success(t *testing.T) {
	repo := newFakeRepo()
	uc := usecase.NewPokemonUseCase(repo)

	got, err := uc.Create(context.Background(), validCreateReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID == "" {
		t.Error("ID harus di-generate")
	}
	if got.Name != "charizard" {
		t.Errorf("name harus dinormalisasi, got %q", got.Name)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Error("timestamp harus terisi")
	}
	if _, ok := repo.data[got.ID]; !ok {
		t.Error("pokemon harus tersimpan di repository")
	}
}

func TestCreate_ValidationAndConflict(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(r *domainUC.CreatePokemonRequest)
		checkFn func(error) bool
	}{
		{"primary type tidak valid", func(r *domainUC.CreatePokemonRequest) { r.PrimaryType = "plasma" }, apperrors.IsBadRequest},
		{"secondary type tidak valid", func(r *domainUC.CreatePokemonRequest) { r.SecondaryType = typePtr("cosmic") }, apperrors.IsBadRequest},
		{"secondary sama dengan primary", func(r *domainUC.CreatePokemonRequest) { r.SecondaryType = typePtr(entity.PokemonTypeFire) }, apperrors.IsBadRequest},
		{"stat di luar rentang", func(r *domainUC.CreatePokemonRequest) { r.Speed = 300 }, apperrors.IsBadRequest},
		{"nama kosong", func(r *domainUC.CreatePokemonRequest) { r.Name = "   " }, apperrors.IsBadRequest},
		{"pokedex number duplikat", func(r *domainUC.CreatePokemonRequest) { r.PokedexNumber = 25 }, apperrors.IsConflict},
		{"nama duplikat (case-insensitive)", func(r *domainUC.CreatePokemonRequest) { r.Name = "PIKACHU" }, apperrors.IsConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := usecase.NewPokemonUseCase(newFakeRepo(pikachu()))
			req := validCreateReq()
			tt.mutate(&req)

			_, err := uc.Create(context.Background(), req)
			if !tt.checkFn(err) {
				t.Fatalf("error tidak sesuai, got %v", err)
			}
		})
	}
}

func TestGetByID_NotFound(t *testing.T) {
	uc := usecase.NewPokemonUseCase(newFakeRepo())

	_, err := uc.GetByID(context.Background(), "does-not-exist")
	if !apperrors.IsNotFound(err) {
		t.Fatalf("expected NOT_FOUND, got %v", err)
	}
}

func TestGetAll_NormalizesFilter(t *testing.T) {
	repo := newFakeRepo(pikachu())
	uc := usecase.NewPokemonUseCase(repo)

	_, total, err := uc.GetAll(context.Background(), repository.PokemonFilter{Page: 0, PageSize: 1000, Search: "  pika "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if repo.lastFilter.Page != 1 || repo.lastFilter.PageSize != 100 || repo.lastFilter.Search != "pika" {
		t.Errorf("filter tidak dinormalisasi: %+v", repo.lastFilter)
	}
}

func TestGetAll_InvalidType(t *testing.T) {
	uc := usecase.NewPokemonUseCase(newFakeRepo())

	_, _, err := uc.GetAll(context.Background(), repository.PokemonFilter{Type: typePtr("plasma")})
	if !apperrors.IsBadRequest(err) {
		t.Fatalf("expected BAD_REQUEST, got %v", err)
	}
}

func TestGetAll_RepositoryErrorIsWrapped(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := newFakeRepo()
	repo.findAllErr = dbErr
	uc := usecase.NewPokemonUseCase(repo)

	_, _, err := uc.GetAll(context.Background(), repository.PokemonFilter{})
	if !errors.Is(err, dbErr) {
		t.Fatalf("error harus di-wrap dengan %%w, got %v", err)
	}
}

func TestUpdate_PartialFields(t *testing.T) {
	p := pikachu()
	repo := newFakeRepo(p)
	uc := usecase.NewPokemonUseCase(repo)

	got, err := uc.Update(context.Background(), p.ID, domainUC.UpdatePokemonRequest{
		Speed:         intPtr(120),
		SecondaryType: typePtr(entity.PokemonTypeFairy),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Speed != 120 || got.HP != 35 {
		t.Errorf("hanya speed yang berubah, got speed=%d hp=%d", got.Speed, got.HP)
	}
	if got.SecondaryType == nil || *got.SecondaryType != entity.PokemonTypeFairy {
		t.Errorf("secondary type tidak ter-update: %v", got.SecondaryType)
	}

	got, err = uc.Update(context.Background(), p.ID, domainUC.UpdatePokemonRequest{ClearSecondaryType: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SecondaryType != nil {
		t.Error("secondary type harus terhapus")
	}
}

func TestUpdate_SameNameIsNotConflict(t *testing.T) {
	p := pikachu()
	uc := usecase.NewPokemonUseCase(newFakeRepo(p))

	if _, err := uc.Update(context.Background(), p.ID, domainUC.UpdatePokemonRequest{Name: strPtr("Pikachu")}); err != nil {
		t.Fatalf("update ke nama sendiri tidak boleh conflict: %v", err)
	}
}

func TestUpdate_ConflictWithOtherPokemon(t *testing.T) {
	other := &entity.Pokemon{
		ID: "22222222-2222-2222-2222-222222222222", PokedexNumber: 26, Name: "raichu",
		PrimaryType: entity.PokemonTypeElectric, HP: 60, Attack: 90, Defense: 55, Speed: 110,
	}
	p := pikachu()
	uc := usecase.NewPokemonUseCase(newFakeRepo(p, other))

	_, err := uc.Update(context.Background(), p.ID, domainUC.UpdatePokemonRequest{Name: strPtr("raichu")})
	if !apperrors.IsConflict(err) {
		t.Fatalf("expected CONFLICT, got %v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	uc := usecase.NewPokemonUseCase(newFakeRepo())

	_, err := uc.Update(context.Background(), "missing", domainUC.UpdatePokemonRequest{})
	if !apperrors.IsNotFound(err) {
		t.Fatalf("expected NOT_FOUND, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	p := pikachu()
	repo := newFakeRepo(p)
	uc := usecase.NewPokemonUseCase(repo)

	if err := uc.Delete(context.Background(), p.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.data[p.ID]; ok {
		t.Error("pokemon harus terhapus")
	}
	if err := uc.Delete(context.Background(), p.ID); !apperrors.IsNotFound(err) {
		t.Fatalf("delete kedua harus NOT_FOUND, got %v", err)
	}
}
