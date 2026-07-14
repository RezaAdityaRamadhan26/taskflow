// Package main is the entry point for the TaskFlow backend server.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/config"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/database"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/handler"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/middleware"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/repository"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/service"
)

func main() {
	// ========================================
	// Load Configuration
	// ========================================
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Failed to load config: %v", err)
	}
	log.Printf("[INFO] Environment: %s", cfg.ServerEnv)

	// ========================================
	// Database Connection
	// ========================================
	pool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("[INFO] Database connected successfully")

	// ========================================
	// Initialize Repository (sqlc)
	// ========================================
	queries := repository.New(pool)

	// ========================================
	// Initialize Services
	// ========================================
	authService := service.NewAuthService(queries, cfg)
	workspaceService := service.NewWorkspaceService(queries)
	boardService := service.NewBoardService(queries)
	listService := service.NewListService(queries)
	cardService := service.NewCardService(queries)

	// ========================================
	// Initialize Handlers
	// ========================================
	authHandler := handler.NewAuthHandler(authService, cfg)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
	boardHandler := handler.NewBoardHandler(boardService)
	listHandler := handler.NewListHandler(listService)
	cardHandler := handler.NewCardHandler(cardService)

	// ========================================
	// Fiber App Setup
	// ========================================
	app := fiber.New(fiber.Config{
		AppName:      "TaskFlow API v1.0",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorHandler: globalErrorHandler,
	})

	// ========================================
	// Global Middleware
	// ========================================
	// Recovery middleware - prevents panics from crashing the server
	app.Use(recover.New())

	// Logger middleware - logs all HTTP requests
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// ========================================
	// Routes
	// ========================================
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "TaskFlow API is running",
			"version": "1.0.0",
		})
	})

	// Auth routes
	auth := api.Group("/auth")

	// Rate limiter for auth endpoints (5 requests per minute per IP)
	authLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Too many requests. Please try again later.",
			})
		},
	})

	auth.Post("/register", authLimiter, authHandler.Register)
	auth.Post("/login", authLimiter, authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	// Protected auth routes (require valid JWT)
	authProtected := auth.Group("", middleware.AuthMiddleware(cfg.JWTAccessSecret))
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/me", authHandler.Me)

	// Workspace routes (all protected)
	workspaces := api.Group("/workspaces", middleware.AuthMiddleware(cfg.JWTAccessSecret))
	workspaces.Post("/", workspaceHandler.Create)
	workspaces.Get("/", workspaceHandler.List)
	workspaces.Get("/:id", workspaceHandler.Get)
	workspaces.Put("/:id", workspaceHandler.Update)
	workspaces.Delete("/:id", workspaceHandler.Delete)

	// Workspace member routes
	workspaces.Post("/:id/members", workspaceHandler.InviteMember)
	workspaces.Get("/:id/members", workspaceHandler.ListMembers)
	workspaces.Patch("/:id/members/:userId", workspaceHandler.UpdateMemberRole)
	workspaces.Delete("/:id/members/:userId", workspaceHandler.RemoveMember)

	// Board routes (all protected)
	boards := api.Group("/boards", middleware.AuthMiddleware(cfg.JWTAccessSecret))
	boards.Post("/", boardHandler.Create)
	boards.Get("/:id", boardHandler.Get)
	boards.Put("/:id", boardHandler.Update)
	boards.Delete("/:id", boardHandler.Delete)
	
	// Nested boards under workspace
	workspaces.Get("/:workspaceId/boards", boardHandler.List)

	// List routes (all protected)
	lists := api.Group("/lists", middleware.AuthMiddleware(cfg.JWTAccessSecret))
	lists.Post("/", listHandler.Create)
	lists.Get("/:id", listHandler.Get)
	lists.Put("/:id", listHandler.Update)
	lists.Delete("/:id", listHandler.Delete)

	// Nested lists under board
	boards.Get("/:boardId/lists", listHandler.List)

	// Card routes (all protected)
	cards := api.Group("/cards", middleware.AuthMiddleware(cfg.JWTAccessSecret))
	cards.Post("/", cardHandler.Create)
	cards.Get("/:id", cardHandler.Get)
	cards.Put("/:id", cardHandler.Update)
	cards.Delete("/:id", cardHandler.Delete)

	// Nested cards under list
	lists.Get("/:listId/cards", cardHandler.List)

	// ========================================
	// 404 handler
	// ========================================
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Endpoint not found",
		})
	})

	// ========================================
	// Graceful Shutdown
	// ========================================
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("[INFO] Shutting down server gracefully...")
		if err := app.Shutdown(); err != nil {
			log.Printf("[ERROR] Server shutdown error: %v", err)
		}
	}()

	// ========================================
	// Start Server
	// ========================================
	addr := ":" + cfg.ServerPort
	log.Printf("[INFO] Starting TaskFlow API on http://localhost%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}

// globalErrorHandler is the default error handler for unhandled Fiber errors.
func globalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	log.Printf("[ERROR] %d - %s - %s %s", code, message, c.Method(), c.Path())

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}
