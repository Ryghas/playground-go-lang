package handler

import (
	"net/http"

	"pokemon-service/internal/delivery/http/response"
	domainUC "pokemon-service/internal/domain/usecase"
	"pokemon-service/pkg/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PokemonHandler struct {
	pokemonUC domainUC.PokemonUseCase
}

func NewPokemonHandler(pokemonUC domainUC.PokemonUseCase) *PokemonHandler {
	return &PokemonHandler{pokemonUC: pokemonUC}
}

func (h *PokemonHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/pokemons")
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.POST("", h.Create)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

// List godoc
//
//	@Summary		List pokemon
//	@Description	Daftar pokemon dengan pagination, filter tipe, dan pencarian nama (case-insensitive). Diurutkan berdasarkan pokedex_number.
//	@Tags			pokemons
//	@Produce		json
//	@Param			page		query		int		false	"Nomor halaman"							minimum(1)	default(1)
//	@Param			page_size	query		int		false	"Jumlah item per halaman"				minimum(1)	maximum(100)	default(10)
//	@Param			type		query		string	false	"Filter tipe (primary atau secondary)"	Enums(normal,fire,water,grass,electric,ice,fighting,poison,ground,flying,psychic,bug,rock,ghost,dragon,dark,steel,fairy)
//	@Param			search		query		string	false	"Cari berdasarkan nama"
//	@Success		200			{object}	response.Body{data=[]PokemonResponse,meta=pagination.Meta}
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Router			/pokemons [get]
func (h *PokemonHandler) List(c *gin.Context) {
	var q listPokemonQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := q.toFilter()
	pokemons, total, err := h.pokemonUC.GetAll(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	page, pageSize := pagination.Normalize(filter.Page, filter.PageSize)
	response.OKWithMeta(c, http.StatusOK, newPokemonResponses(pokemons), pagination.NewMeta(page, pageSize, total))
}

// GetByID godoc
//
//	@Summary	Detail pokemon
//	@Tags		pokemons
//	@Produce	json
//	@Param		id	path		string	true	"Pokemon ID (UUID)"	format(uuid)
//	@Success	200	{object}	response.Body{data=PokemonResponse}
//	@Failure	400	{object}	response.ErrorResponse
//	@Failure	404	{object}	response.ErrorResponse
//	@Failure	500	{object}	response.ErrorResponse
//	@Router		/pokemons/{id} [get]
func (h *PokemonHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	pokemon, err := h.pokemonUC.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, http.StatusOK, newPokemonResponse(pokemon))
}

// Create godoc
//
//	@Summary		Tambah pokemon
//	@Description	Nama disimpan lowercase. Nama dan pokedex_number harus unik. secondary_type tidak boleh sama dengan primary_type.
//	@Tags			pokemons
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreatePokemonRequest	true	"Data pokemon"
//	@Success		201		{object}	response.Body{data=PokemonResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		409		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/pokemons [post]
func (h *PokemonHandler) Create(c *gin.Context) {
	var req CreatePokemonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	pokemon, err := h.pokemonUC.Create(c.Request.Context(), req.toUseCase())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, http.StatusCreated, newPokemonResponse(pokemon))
}

// Update godoc
//
//	@Summary		Update sebagian data pokemon
//	@Description	Partial update: hanya field yang dikirim yang diubah. Kirim "secondary_type": "" untuk menghapus secondary type.
//	@Tags			pokemons
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Pokemon ID (UUID)"	format(uuid)
//	@Param			body	body		UpdatePokemonRequest	true	"Field yang diubah"
//	@Success		200		{object}	response.Body{data=PokemonResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Failure		409		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/pokemons/{id} [patch]
func (h *PokemonHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req UpdatePokemonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	pokemon, err := h.pokemonUC.Update(c.Request.Context(), id, req.toUseCase())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, http.StatusOK, newPokemonResponse(pokemon))
}

// Delete godoc
//
//	@Summary		Hapus pokemon
//	@Description	Soft delete. Nama dan pokedex_number bisa dipakai lagi setelah dihapus.
//	@Tags			pokemons
//	@Produce		json
//	@Param			id	path	string	true	"Pokemon ID (UUID)"	format(uuid)
//	@Success		204
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/pokemons/{id} [delete]
func (h *PokemonHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.pokemonUC.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// parseID memvalidasi path param :id berformat UUID.
func parseID(c *gin.Context) (string, bool) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		response.BadRequest(c, "id harus berformat UUID")
		return "", false
	}
	return id, true
}
