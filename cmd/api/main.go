package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mhakimsaputra17/xyz-football-api/internal/config"
	"github.com/mhakimsaputra17/xyz-football-api/internal/handler"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"github.com/mhakimsaputra17/xyz-football-api/internal/repository"
	"github.com/mhakimsaputra17/xyz-football-api/internal/router"
	"github.com/mhakimsaputra17/xyz-football-api/internal/service"
	jwtpkg "github.com/mhakimsaputra17/xyz-football-api/pkg/jwt"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//	@title						XYZ Football API
//	@version					1.0
//	@description				REST API for managing football teams, players, match schedules, results, and reports for Perusahaan XYZ.
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				admin@xyz-football.com
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter your bearer token in the format: Bearer {token}

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	slog.Info("configuration loaded",
		"app", cfg.App.Name,
		"env", cfg.App.Env,
		"port", cfg.Server.Port,
	)

	// 2. Set GIN mode based on environment
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. Connect to PostgreSQL
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	slog.Info("database connected successfully")

	// 4. Run AutoMigrate
	if err := autoMigrate(db); err != nil {
		log.Fatalf("failed to run auto migration: %v", err)
	}
	slog.Info("database migration completed")

	// 5. Seed default admin
	if err := seedAdmin(db, cfg.App.Env); err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	// 6. Initialize JWT service
	jwtService := jwtpkg.NewService(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiration,
		cfg.JWT.RefreshExpiration,
	)

	// 7. Initialize repositories (all take *gorm.DB)
	adminRepo := repository.NewAdminRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	playerRepo := repository.NewPlayerRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	goalRepo := repository.NewGoalRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// 8. Initialize services
	authService := service.NewAuthService(adminRepo, refreshTokenRepo, jwtService)
	teamService := service.NewTeamService(teamRepo)
	playerService := service.NewPlayerService(playerRepo, teamRepo)
	matchService := service.NewMatchService(matchRepo, teamRepo, playerRepo, goalRepo)
	reportService := service.NewReportService(matchRepo, goalRepo)

	// 9. Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	teamHandler := handler.NewTeamHandler(teamService)
	playerHandler := handler.NewPlayerHandler(playerService)
	matchHandler := handler.NewMatchHandler(matchService)
	reportHandler := handler.NewReportHandler(reportService)

	// 10. Setup router
	r := router.Setup(
		cfg.App.Env,
		jwtService,
		authHandler,
		teamHandler,
		playerHandler,
		matchHandler,
		reportHandler,
	)

	// 11. Start HTTP server with graceful configuration
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	slog.Info("starting server", "port", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}

// connectDB establishes a connection to the PostgreSQL database using GORM.
func connectDB(cfg *config.Config) (*gorm.DB, error) {
	// Configure GORM logger based on environment
	var gormLogLevel logger.LogLevel
	switch cfg.App.Env {
	case "production":
		gormLogLevel = logger.Silent
	case "development":
		gormLogLevel = logger.Info
	default:
		gormLogLevel = logger.Warn
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// autoMigrate runs GORM AutoMigrate for all models.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Admin{},
		&model.RefreshToken{},
		&model.Team{},
		&model.Player{},
		&model.Match{},
		&model.Goal{},
	)
}

// seedAdmin creates a default admin user if none exists.
// Credentials are read from ADMIN_USERNAME and ADMIN_PASSWORD environment
// variables. In development, defaults are used when those vars are unset.
// In production the application refuses to start with default credentials.
func seedAdmin(db *gorm.DB, appEnv string) error {
	var count int64
	if err := db.Model(&model.Admin{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count admins: %w", err)
	}

	if count > 0 {
		slog.Info("admin already exists, skipping seeder")
		return nil
	}

	username := viper.GetString("ADMIN_USERNAME")
	password := viper.GetString("ADMIN_PASSWORD")

	// Fall back to defaults only in non-production environments
	if username == "" {
		if appEnv == "production" {
			return fmt.Errorf("ADMIN_USERNAME is required in production")
		}
		username = "admin"
	}
	if password == "" {
		if appEnv == "production" {
			return fmt.Errorf("ADMIN_PASSWORD is required in production")
		}
		password = "password123"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	admin := model.Admin{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	slog.Info("default admin seeded", "username", username)

	return nil
}
