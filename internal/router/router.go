package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/mhakimsaputra17/xyz-football-api/docs"
	"github.com/mhakimsaputra17/xyz-football-api/internal/handler"
	"github.com/mhakimsaputra17/xyz-football-api/internal/middleware"
	jwtpkg "github.com/mhakimsaputra17/xyz-football-api/pkg/jwt"
)

// Setup configures all API routes and returns the GIN engine.
// Swagger UI is only available in non-production environments.
func Setup(
	appEnv string,
	jwtService *jwtpkg.Service,
	authHandler *handler.AuthHandler,
	teamHandler *handler.TeamHandler,
	playerHandler *handler.PlayerHandler,
	matchHandler *handler.MatchHandler,
	reportHandler *handler.ReportHandler,
) *gin.Engine {
	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint — public, no auth required.
	// Used by Docker HEALTHCHECK and load balancers.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger UI endpoint — disabled in production to prevent API spec leakage.
	if appEnv != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API v1 group
	v1 := r.Group("/api/v1")

	// --- Public routes (no auth required) ---
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// --- Protected routes (JWT auth required) ---
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	{
		// Auth — logout requires authentication
		protected.POST("/auth/logout", authHandler.Logout)

		// Teams CRUD
		teams := protected.Group("/teams")
		{
			teams.GET("", teamHandler.GetAll)
			teams.GET("/:id", teamHandler.GetByID)
			teams.POST("", teamHandler.Create)
			teams.PUT("/:id", teamHandler.Update)
			teams.DELETE("/:id", teamHandler.Delete)

			// Players nested under teams (create + list)
			teams.GET("/:id/players", playerHandler.GetAllByTeamID)
			teams.POST("/:id/players", playerHandler.Create)
		}

		// Players (get, update, delete — not nested under teams)
		players := protected.Group("/players")
		{
			players.GET("/:id", playerHandler.GetByID)
			players.PUT("/:id", playerHandler.Update)
			players.DELETE("/:id", playerHandler.Delete)
		}

		// Matches CRUD + Results
		matches := protected.Group("/matches")
		{
			matches.GET("", matchHandler.GetAll)
			matches.GET("/:id", matchHandler.GetByID)
			matches.POST("", matchHandler.Create)
			matches.PUT("/:id", matchHandler.Update)
			matches.DELETE("/:id", matchHandler.Delete)

			// Match results (submit + update)
			matches.POST("/:id/result", matchHandler.SubmitResult)
			matches.PUT("/:id/result", matchHandler.UpdateResult)
		}

		// Reports (read-only)
		reports := protected.Group("/reports")
		{
			reports.GET("/matches", reportHandler.GetMatchReports)
			reports.GET("/matches/:id", reportHandler.GetMatchReportByID)
		}
	}

	return r
}
