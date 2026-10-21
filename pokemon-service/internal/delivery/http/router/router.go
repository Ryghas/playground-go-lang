package router

import (
	"log/slog"
	"net/http"

	"pokemon-service/config"
	"pokemon-service/docs"
	"pokemon-service/internal/delivery/http/handler"
	"pokemon-service/internal/delivery/http/middleware"
	domainUC "pokemon-service/internal/domain/usecase"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(cfg *config.Config, log *slog.Logger, pokemonUC domainUC.PokemonUseCase) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if cfg.SwaggerEnabled {
		// Host dikosongkan agar "Try it out" memakai host tempat UI dibuka.
		docs.SwaggerInfo.Host = ""
		r.GET("/swagger", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
		})
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1 := r.Group("/api/v1")
	handler.NewPokemonHandler(pokemonUC).Register(v1)

	return r
}
