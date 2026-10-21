package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pokemon-service/internal/delivery/http/handler"
	"pokemon-service/internal/domain/entity"
	"pokemon-service/internal/domain/repository"
	domainUC "pokemon-service/internal/domain/usecase"
	apperrors "pokemon-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

const validID = "11111111-1111-1111-1111-111111111111"

// fakePokemonUseCase — stub domainUC.PokemonUseCase; tiap method bisa di-override per test.
type fakePokemonUseCase struct {
	getByID func(ctx context.Context, id string) (*entity.Pokemon, error)
	getAll  func(ctx context.Context, f repository.PokemonFilter) ([]*entity.Pokemon, int64, error)
	create  func(ctx context.Context, req domainUC.CreatePokemonRequest) (*entity.Pokemon, error)
	update  func(ctx context.Context, id string, req domainUC.UpdatePokemonRequest) (*entity.Pokemon, error)
	delete  func(ctx context.Context, id string) error
}

func (f *fakePokemonUseCase) GetByID(ctx context.Context, id string) (*entity.Pokemon, error) {
	return f.getByID(ctx, id)
}

func (f *fakePokemonUseCase) GetAll(ctx context.Context, filter repository.PokemonFilter) ([]*entity.Pokemon, int64, error) {
	return f.getAll(ctx, filter)
}

func (f *fakePokemonUseCase) Create(ctx context.Context, req domainUC.CreatePokemonRequest) (*entity.Pokemon, error) {
	return f.create(ctx, req)
}

func (f *fakePokemonUseCase) Update(ctx context.Context, id string, req domainUC.UpdatePokemonRequest) (*entity.Pokemon, error) {
	return f.update(ctx, id, req)
}

func (f *fakePokemonUseCase) Delete(ctx context.Context, id string) error {
	return f.delete(ctx, id)
}

func pikachu() *entity.Pokemon {
	ts := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	return &entity.Pokemon{
		ID: validID, PokedexNumber: 25, Name: "pikachu", PrimaryType: entity.PokemonTypeElectric,
		HP: 35, Attack: 55, Defense: 40, Speed: 90, CreatedAt: ts, UpdatedAt: ts,
	}
}

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    json.RawMessage `json:"meta"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newRouter(uc domainUC.PokemonUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler.NewPokemonHandler(uc).Register(r.Group("/api/v1"))
	return r
}

func doRequest(t *testing.T, r http.Handler, method, path, body string) (*httptest.ResponseRecorder, apiResponse) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("response bukan JSON valid: %v (%s)", err, w.Body.String())
		}
	}
	return w, resp
}

func assertError(t *testing.T, w *httptest.ResponseRecorder, resp apiResponse, status int, code string) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status = %d, want %d (body: %s)", w.Code, status, w.Body.String())
	}
	if resp.Success || resp.Error == nil || resp.Error.Code != code {
		t.Fatalf("error code tidak sesuai, want %s, body: %s", code, w.Body.String())
	}
}

func TestList(t *testing.T) {
	var gotFilter repository.PokemonFilter
	uc := &fakePokemonUseCase{getAll: func(_ context.Context, f repository.PokemonFilter) ([]*entity.Pokemon, int64, error) {
		gotFilter = f
		return []*entity.Pokemon{pikachu()}, 21, nil
	}}

	w, resp := doRequest(t, newRouter(uc), http.MethodGet, "/api/v1/pokemons?page=2&page_size=10&type=electric&search=pika", "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	if gotFilter.Page != 2 || gotFilter.PageSize != 10 || gotFilter.Search != "pika" ||
		gotFilter.Type == nil || *gotFilter.Type != entity.PokemonTypeElectric {
		t.Errorf("query tidak diteruskan ke use case dengan benar: %+v", gotFilter)
	}

	var meta struct {
		Page       int   `json:"page"`
		TotalItems int64 `json:"total_items"`
		TotalPages int   `json:"total_pages"`
	}
	if err := json.Unmarshal(resp.Meta, &meta); err != nil {
		t.Fatalf("meta tidak valid: %v", err)
	}
	if meta.Page != 2 || meta.TotalItems != 21 || meta.TotalPages != 3 {
		t.Errorf("meta = %+v", meta)
	}
}

func TestList_InvalidQuery(t *testing.T) {
	uc := &fakePokemonUseCase{}
	w, resp := doRequest(t, newRouter(uc), http.MethodGet, "/api/v1/pokemons?page_size=1000", "")
	assertError(t, w, resp, http.StatusBadRequest, apperrors.CodeBadRequest)
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		ucErr  error
		status int
		code   string
	}{
		{"sukses", "/api/v1/pokemons/" + validID, nil, http.StatusOK, ""},
		{"id bukan uuid", "/api/v1/pokemons/abc", nil, http.StatusBadRequest, apperrors.CodeBadRequest},
		{"tidak ditemukan", "/api/v1/pokemons/" + validID, apperrors.NewNotFound("pokemon tidak ditemukan"), http.StatusNotFound, apperrors.CodeNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &fakePokemonUseCase{getByID: func(_ context.Context, id string) (*entity.Pokemon, error) {
				if tt.ucErr != nil {
					return nil, tt.ucErr
				}
				return pikachu(), nil
			}}

			w, resp := doRequest(t, newRouter(uc), http.MethodGet, tt.path, "")
			if tt.code != "" {
				assertError(t, w, resp, tt.status, tt.code)
				return
			}
			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}
			var got handler.PokemonResponse
			if err := json.Unmarshal(resp.Data, &got); err != nil {
				t.Fatalf("data tidak valid: %v", err)
			}
			if got.ID != validID || got.Name != "pikachu" || got.SecondaryType != nil {
				t.Errorf("data = %+v", got)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	validBody := `{"pokedex_number":25,"name":"Pikachu","primary_type":"electric","secondary_type":"","hp":35,"attack":55,"defense":40,"speed":90}`

	t.Run("sukses", func(t *testing.T) {
		var gotReq domainUC.CreatePokemonRequest
		uc := &fakePokemonUseCase{create: func(_ context.Context, req domainUC.CreatePokemonRequest) (*entity.Pokemon, error) {
			gotReq = req
			return pikachu(), nil
		}}

		w, _ := doRequest(t, newRouter(uc), http.MethodPost, "/api/v1/pokemons", validBody)
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
		}
		if gotReq.Name != "Pikachu" || gotReq.PrimaryType != entity.PokemonTypeElectric || gotReq.HP != 35 {
			t.Errorf("request ke use case = %+v", gotReq)
		}
		if gotReq.SecondaryType != nil {
			t.Error("secondary_type kosong harus dianggap tidak ada")
		}
	})

	t.Run("field wajib kosong", func(t *testing.T) {
		uc := &fakePokemonUseCase{}
		w, resp := doRequest(t, newRouter(uc), http.MethodPost, "/api/v1/pokemons", `{"name":"missingno"}`)
		assertError(t, w, resp, http.StatusBadRequest, apperrors.CodeBadRequest)
	})

	t.Run("stat di luar rentang", func(t *testing.T) {
		uc := &fakePokemonUseCase{}
		body := strings.Replace(validBody, `"hp":35`, `"hp":999`, 1)
		w, resp := doRequest(t, newRouter(uc), http.MethodPost, "/api/v1/pokemons", body)
		assertError(t, w, resp, http.StatusBadRequest, apperrors.CodeBadRequest)
	})

	t.Run("json rusak", func(t *testing.T) {
		uc := &fakePokemonUseCase{}
		w, resp := doRequest(t, newRouter(uc), http.MethodPost, "/api/v1/pokemons", `{"name":`)
		assertError(t, w, resp, http.StatusBadRequest, apperrors.CodeBadRequest)
	})

	t.Run("conflict", func(t *testing.T) {
		uc := &fakePokemonUseCase{create: func(context.Context, domainUC.CreatePokemonRequest) (*entity.Pokemon, error) {
			return nil, apperrors.NewConflict("pokedex number 25 sudah terdaftar")
		}}
		w, resp := doRequest(t, newRouter(uc), http.MethodPost, "/api/v1/pokemons", validBody)
		assertError(t, w, resp, http.StatusConflict, apperrors.CodeConflict)
	})

	t.Run("error tak terduga tidak bocor ke client", func(t *testing.T) {
		uc := &fakePokemonUseCase{create: func(context.Context, domainUC.CreatePokemonRequest) (*entity.Pokemon, error) {
			return nil, errors.New("dial tcp 10.0.0.5:5432: connection refused")
		}}
		w, resp := doRequest(t, newRouter(uc), http.MethodPost, "/api/v1/pokemons", validBody)
		assertError(t, w, resp, http.StatusInternalServerError, apperrors.CodeInternal)
		if strings.Contains(w.Body.String(), "10.0.0.5") {
			t.Errorf("detail error internal bocor: %s", w.Body.String())
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("partial update dan hapus secondary type", func(t *testing.T) {
		var gotID string
		var gotReq domainUC.UpdatePokemonRequest
		uc := &fakePokemonUseCase{update: func(_ context.Context, id string, req domainUC.UpdatePokemonRequest) (*entity.Pokemon, error) {
			gotID, gotReq = id, req
			return pikachu(), nil
		}}

		w, _ := doRequest(t, newRouter(uc), http.MethodPatch, "/api/v1/pokemons/"+validID, `{"speed":100,"secondary_type":""}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
		}
		if gotID != validID {
			t.Errorf("id = %s", gotID)
		}
		if gotReq.Speed == nil || *gotReq.Speed != 100 {
			t.Errorf("speed tidak diteruskan: %+v", gotReq.Speed)
		}
		if gotReq.HP != nil || gotReq.Name != nil {
			t.Error("field yang tidak dikirim harus nil")
		}
		if !gotReq.ClearSecondaryType || gotReq.SecondaryType != nil {
			t.Errorf("secondary_type \"\" harus jadi ClearSecondaryType: %+v", gotReq)
		}
	})

	t.Run("stat tidak valid", func(t *testing.T) {
		uc := &fakePokemonUseCase{}
		w, resp := doRequest(t, newRouter(uc), http.MethodPatch, "/api/v1/pokemons/"+validID, `{"speed":0}`)
		assertError(t, w, resp, http.StatusBadRequest, apperrors.CodeBadRequest)
	})

	t.Run("tidak ditemukan", func(t *testing.T) {
		uc := &fakePokemonUseCase{update: func(context.Context, string, domainUC.UpdatePokemonRequest) (*entity.Pokemon, error) {
			return nil, apperrors.NewNotFound("pokemon tidak ditemukan")
		}}
		w, resp := doRequest(t, newRouter(uc), http.MethodPatch, "/api/v1/pokemons/"+validID, `{"speed":100}`)
		assertError(t, w, resp, http.StatusNotFound, apperrors.CodeNotFound)
	})
}

func TestDelete(t *testing.T) {
	t.Run("sukses", func(t *testing.T) {
		uc := &fakePokemonUseCase{delete: func(context.Context, string) error { return nil }}
		w, _ := doRequest(t, newRouter(uc), http.MethodDelete, "/api/v1/pokemons/"+validID, "")
		if w.Code != http.StatusNoContent || w.Body.Len() != 0 {
			t.Fatalf("status = %d, body = %q", w.Code, w.Body.String())
		}
	})

	t.Run("tidak ditemukan", func(t *testing.T) {
		uc := &fakePokemonUseCase{delete: func(context.Context, string) error {
			return apperrors.NewNotFound("pokemon tidak ditemukan")
		}}
		w, resp := doRequest(t, newRouter(uc), http.MethodDelete, "/api/v1/pokemons/"+validID, "")
		assertError(t, w, resp, http.StatusNotFound, apperrors.CodeNotFound)
	})
}
