package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ahmedelhadi17776/Compass/Backend_go/docs" // swagger docs
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/routes"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/calendar"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/habits"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/organization"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/project"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/roles"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/task"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/todos"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/workflow"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/cache"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/migrations"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/scheduler"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/config"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title           Compass API
// @version         1.0
// @description     A task management API with user authentication and authorization.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @host      localhost:8000
// @BasePath

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// RequestLoggerMiddleware logs all incoming HTTP requests
func RequestLoggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		log.Info("Request started",
			zap.String("path", path),
			zap.String("method", method),
			zap.String("client_ip", c.ClientIP()),
		)

		c.Next()

		log.Info("Request completed",
			zap.String("path", path),
			zap.String("method", method),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("") // Empty string will make it search in default locations
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	log := logger.NewLogger()
	defer log.Sync()

	log.Info("Configuration loaded successfully")
	log.Info("Server mode: " + cfg.Server.Mode)

	// Log OAuth2 configuration
	log.Info("OAuth2 configuration",
		zap.Bool("enabled", cfg.Auth.OAuth2.Enabled),
		zap.String("callback_url", cfg.Auth.OAuth2.CallbackURL),
		zap.Int("state_timeout", cfg.Auth.OAuth2.StateTimeout),
		zap.Int("providers_count", len(cfg.Auth.OAuth2Providers)))

	for provider, _ := range cfg.Auth.OAuth2Providers {
		log.Info("OAuth2 provider configured", zap.String("provider", provider))
	}

	// Set up Gin
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Set default content type for JSON responses
	gin.DisableBindValidation()
	gin.SetMode(gin.ReleaseMode)
	// Use the standard logger instead of zap for gin
	gin.DefaultWriter = os.Stdout

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(RequestLoggerMiddleware(log))
	// Configure gin to use proper content type for JSON
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods: cfg.CORS.AllowedMethods,
		AllowHeaders: append(cfg.CORS.AllowedHeaders,
			"Accept-Encoding",
			"Content-Encoding",
			"Content-Type",
			"Authorization",
			"X-Organization-ID",
			"x-organization-id",
			"X-Forwarded-For",
			"X-Real-IP",
		),
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Encoding",
			"Content-Type",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Vary",
			"X-Organization-ID",
		},
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           12 * time.Hour,
	}))

	// Add Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Connect to database
	db, err := connection.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run database migrations
	if err := migrations.AutoMigrate(db, log.Logger); err != nil {
		log.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize logrus logger for workflow service
	workflowLogger := logrus.New()
	workflowLogger.SetFormatter(&logrus.JSONFormatter{})
	if cfg.Server.Mode == "production" {
		workflowLogger.SetLevel(logrus.InfoLevel)
	} else {
		workflowLogger.SetLevel(logrus.DebugLevel)
	}

	// Initialize repositories
	taskRepo := task.NewRepository(db)
	userRepo := user.NewRepository(db)
	projectRepo := project.NewRepository(db)
	organizationRepo := organization.NewRepository(db)
	rolesRepo := roles.NewRepository(db.DB)
	habitsRepo := habits.NewRepository(db)
	calendarRepo := calendar.NewRepository(db.DB)
	workflowRepo := workflow.NewRepository(db.DB, workflowLogger)
	todosRepo := todos.NewTodoRepository(db)

	// Initialize Redis
	redisConfig := cache.NewConfigFromEnv(cfg)
	redisClient, err := cache.NewRedisClient(redisConfig)
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize rate limiter with Redis client
	rateLimiter := auth.NewRedisRateLimiter(redisClient.GetClient(), 1*time.Minute, 1000)

	// Create cache middleware instances
	cacheMiddleware := middleware.NewCacheMiddleware(redisClient, "compass", 5*time.Minute)
	cacheHandler := cacheMiddleware.CacheResponse()

	// Initialize routes
	authMiddleware := middleware.NewAuthMiddleware(cfg.Auth.JWTSecret)

	// Initialize notification system
	notificationSystem, err := SetupNotificationSystem(
		db,
		log,
		cfg.Server.Mode != "production",
	)
	if err != nil {
		log.Fatal("Failed to initialize notification system", zap.Error(err))
	}
	defer notificationSystem.Shutdown()

	// Initialize habit notification service using the notification service from our system
	habitNotifySvc := habits.NewHabitNotificationService(notificationSystem.Service)
	// Add domain notifier for enhanced capabilities
	habitNotifySvc.WithDomainNotifier(notificationSystem.DomainNotifier)

	// Initialize services
	rolesService := roles.NewService(rolesRepo)
	userService := user.NewService(userRepo, rolesService, redisClient)
	taskService := task.NewService(taskRepo, redisClient, log.Logger)
	projectService := project.NewService(projectRepo)
	organizationService := organization.NewService(organizationRepo)
	habitsService := habits.NewService(habitsRepo, habitNotifySvc, redisClient, log.Logger)
	calendarService := calendar.NewService(calendarRepo, notificationSystem.DomainNotifier, redisClient, log.Logger)
	workflowExecutor := workflow.NewDefaultExecutor(workflowRepo, workflowLogger, notificationSystem.DomainNotifier, rolesService)
	workflowService := workflow.NewService(workflow.ServiceConfig{
		Repository:   workflowRepo,
		Logger:       workflowLogger,
		Executor:     workflowExecutor,
		RolesService: rolesService,
		Notifier:     notificationSystem.DomainNotifier,
	})
	todosService := todos.NewService(todosRepo, redisClient, log.Logger)

	// Initialize OAuth2 service
	oauthService := auth.NewOAuthService(cfg)

	// Initialize MFA service and handler
	// Create a logrus logger specifically for MFA handler
	mfaLogger := logrus.New()
	mfaLogger.SetFormatter(&logrus.JSONFormatter{})
	if cfg.Server.Mode == "production" {
		mfaLogger.SetLevel(logrus.InfoLevel)
	} else {
		mfaLogger.SetLevel(logrus.DebugLevel)
	}

	mfaHandler := handlers.NewMFAHandler(userService, cfg.Auth.JWTSecret, mfaLogger)

	// Initialize and start the scheduler
	habitScheduler := scheduler.NewScheduler(habitsService, log)
	habitScheduler.Start()
	log.Info("Habit scheduler started successfully")

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, cfg.Auth.JWTSecret)
	taskHandler := handlers.NewTaskHandler(taskService)
	authHandler := handlers.NewAuthHandler(rolesService)
	projectHandler := handlers.NewProjectHandler(projectService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	habitsHandler := handlers.NewHabitsHandler(habitsService)
	calendarHandler := handlers.NewCalendarHandler(calendarService)
	workflowHandler := handlers.NewWorkflowHandler(workflowService)
	todosHandler := handlers.NewTodoHandler(todosService)

	oauthHandler := handlers.NewOAuthHandler(oauthService, userService, cfg.Auth.JWTSecret, log.Logger)

	// Initialize dashboard handler
	dashboardHandler := handlers.NewDashboardHandler(
		habitsService,
		taskService,
		todosService,
		calendarService,
		userService,
		redisClient,
		log.Logger,
	)

	// Initialize dashboard routes
	dashboardRoutes := routes.NewDashboardRoutes(
		dashboardHandler,
		authMiddleware,
		cacheHandler,
		log.Logger,
	)
	dashboardRoutes.Register(router.Group("/api"))

	// Initialize notification handler
	notificationHandler := handlers.NewNotificationHandler(notificationSystem.Service, log)

	// Initialize habit notification handler
	habitNotificationHandler := handlers.NewHabitNotificationHandler(habitsService, notificationSystem.Service, habitNotifySvc)

	// Initialize PubSub manager
	pubsubManager := cache.NewPubSubManager(redisClient)

	// Start listening for dashboard events
	go func() {
		ctx := context.Background()
		if err := pubsubManager.StartListening(ctx, "dashboard_updates:*"); err != nil {
			log.Error("Failed to start listening for dashboard events", zap.Error(err))
		}
	}()

	// Debug: Print all registered routes
	log.Info("Registering routes...")

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Info("Registered swagger route at /swagger/*")

	// Set up user routes
	userRoutes := routes.NewUserRoutes(userHandler, cfg.Auth.JWTSecret, rateLimiter)
	userRoutes.RegisterRoutes(router)
	log.Info("Registered user routes at /api/users")

	// Set up MFA routes
	mfaRoutes := routes.NewMFARoutes(mfaHandler, cfg.Auth.JWTSecret)
	mfaRoutes.RegisterRoutes(router)
	log.Info("Registered MFA routes at /api/users/mfa and /api/auth/mfa")

	// Set up auth routes
	authRoutes := routes.NewAuthRoutes(authHandler, cfg.Auth.JWTSecret)
	authRoutes.RegisterRoutes(router)
	log.Info("Registered auth routes at /api/roles")

	// Health check routes (no /api prefix as these are system endpoints)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC(),
		})
	})
	log.Info("Registered health check routes at /health and /health/ready")

	// Add cache health check
	router.GET("/health/cache", func(c *gin.Context) {
		if err := redisClient.HealthCheck(c); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "unhealthy",
				"component": "cache",
				"error":     err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"component": "cache",
			"metrics":   redisClient.GetMetrics(),
		})
	})

	// Apply rate limiting middleware globally
	router.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Task routes (protected)
	taskRoutes := routes.NewTaskRoutes(taskHandler, cfg.Auth.JWTSecret)
	taskRoutes.RegisterRoutes(router, cacheMiddleware)
	log.Info("Registered task routes at /api/tasks")

	// Project routes (protected)
	projectRoutes := routes.NewProjectRoutes(projectHandler, cfg.Auth.JWTSecret)
	projectRoutes.RegisterRoutes(router, cacheMiddleware)
	log.Info("Registered project routes at /api/projects")

	// Organization routes (protected)
	organizationRoutes := routes.NewOrganizationRoutes(organizationHandler, cfg.Auth.JWTSecret)
	organizationRoutes.RegisterRoutes(router)
	log.Info("Registered organization routes at /api/organizations")

	// Habits routes (protected)
	habitsRoutes := routes.NewHabitsRoutes(habitsHandler, cfg.Auth.JWTSecret)
	habitsRoutes.RegisterRoutes(router, cacheMiddleware)
	log.Info("Registered habits routes at /habits")

	// Calendar routes (protected)
	calendarRoutes := routes.NewCalendarRoutes(calendarHandler, cfg.Auth.JWTSecret)
	calendarRoutes.RegisterRoutes(router)
	log.Info("Registered calendar routes at /api/calendar")

	// Workflow routes (protected)
	workflowRoutes := routes.NewWorkflowRoutes(workflowHandler, cfg.Auth.JWTSecret)
	workflowRoutes.RegisterRoutes(router)
	log.Info("Registered workflow routes at /api/workflows")

	// Todos routes (protected)
	todosRoutes := routes.NewTodosRoutes(todosHandler, cfg.Auth.JWTSecret)
	todosRoutes.RegisterRoutes(router, cacheMiddleware)
	log.Info("Registered todos routes at /api/todos")

	// Notification routes (protected)
	notificationRoutes := routes.NewNotificationRoutes(notificationHandler, cfg.Auth.JWTSecret, rateLimiter)
	notificationRoutes.RegisterRoutes(router, cacheMiddleware)

	// Initialize and register habit notification routes
	habitNotificationRoutes := routes.NewHabitNotificationRoutes(habitNotificationHandler, cfg.Auth.JWTSecret)
	habitNotificationRoutes.RegisterRoutes(router, cacheMiddleware)

	// OAuth2 routes
	if cfg.Auth.OAuth2.Enabled {
		log.Info("Registering OAuth2 routes",
			zap.Bool("enabled", cfg.Auth.OAuth2.Enabled))

		oauthRoutes := routes.NewOAuthRoutes(oauthHandler, rateLimiter)
		oauthRoutes.RegisterRoutes(router)

		log.Info("OAuth2 routes registered successfully",
			zap.String("path", "/api/auth/oauth"))
	} else {
		log.Warn("OAuth2 routes not registered because OAuth2 is disabled")
	}

	// Print all registered routes for debugging
	for _, route := range router.Routes() {
		log.Info("Route registered",
			zap.String("method", route.Method),
			zap.String("path", route.Path),
		)
	}

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Info(fmt.Sprintf("Server starting on port %d", cfg.Server.Port))
		log.Info("Swagger documentation available at http://localhost:8000/swagger/index.html")

		// Always use HTTP, never HTTPS
		err := server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited properly")
}
