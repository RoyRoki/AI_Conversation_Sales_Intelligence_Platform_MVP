package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"ai-conversation-platform/internal/api/handlers"
	"ai-conversation-platform/internal/auth"
	"ai-conversation-platform/internal/ai"
	"ai-conversation-platform/internal/rules"
	"ai-conversation-platform/internal/services/agentassist"
	"ai-conversation-platform/internal/services/analytics"
	"ai-conversation-platform/internal/services/autoreply"
	"ai-conversation-platform/internal/services/conversation"
	"ai-conversation-platform/internal/storage/chroma"
	"ai-conversation-platform/internal/storage/postgres"
)

func main() {
	// Initialize database client
	dbClient, err := postgres.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbClient.Close()

	// Initialize storage layers
	conversationStorage := postgres.NewConversationStorage(dbClient)

	// Initialize Chroma DB client
	chromaClient, err := chroma.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize Chroma client: %v", err)
		log.Println("AI features will be disabled")
		chromaClient = nil
	} else {
		if err := chromaClient.HealthCheck(); err != nil {
			log.Printf("Warning: Chroma DB health check failed: %v", err)
		}
	}

	// Initialize AI components (if Chroma and Gemini are available)
	var analyzer *ai.Analyzer
	var embeddingService *ai.EmbeddingService
	if chromaClient != nil {
		geminiClient, err := ai.NewGeminiClient()
		if err != nil {
			log.Printf("Warning: Failed to initialize Gemini client: %v", err)
			log.Println("AI features will be disabled")
		} else {
			// Initialize AI services
			retriever := chroma.NewRetriever(chromaClient)
			embeddingService = ai.NewEmbeddingService(geminiClient, chromaClient)
			analyzer = ai.NewAnalyzer(geminiClient, retriever, embeddingService, conversationStorage)

			// Health check Gemini
			if err := geminiClient.HealthCheck(); err != nil {
				log.Printf("Warning: Gemini API health check failed: %v", err)
			} else {
				log.Println("AI components initialized successfully")
			}
		}
	}

	// Initialize services
	ingestionService := conversation.NewIngestionService(conversationStorage)
	if analyzer != nil {
		ingestionService.SetAnalyzer(analyzer)
	}

	// Initialize storage layers
	userStorage := postgres.NewUserStorage(dbClient)
	ruleStorage := postgres.NewRuleStorage(dbClient)
	memoryStorage := postgres.NewMemoryStorage(dbClient)
	brandToneStorage := postgres.NewBrandToneStorage(dbClient)
	productStorage := postgres.NewProductStorage(dbClient)
	autoReplyGlobalStorage := postgres.NewAutoReplyStorage(dbClient)
	autoReplyConversationStorage := postgres.NewAutoReplyStorage(dbClient)
	suggestionsStorage := postgres.NewSuggestionsStorage(dbClient)

	// Initialize AI components for agent assist (if available)
	var agentAssistService *agentassist.AgentAssistService
	if analyzer != nil && chromaClient != nil && embeddingService != nil {
		geminiClient, err := ai.NewGeminiClient()
		if err == nil {
			retriever := chroma.NewRetriever(chromaClient)
			ruleEngine := rules.NewRuleEngine()
			
			agentAssistService = agentassist.NewAgentAssistService(
				analyzer,
				geminiClient,
				retriever,
				embeddingService,
				ruleEngine,
				ruleStorage,
				conversationStorage,
				memoryStorage,
				brandToneStorage,
				suggestionsStorage,
			)
			log.Println("Agent assist service initialized successfully")
		}
	}

	// Initialize analytics service
	analyticsService := analytics.NewAnalyticsService(conversationStorage)

	// Initialize auto-reply service (if agent assist is available)
	var autoReplyService *autoreply.AutoReplyService
	if agentAssistService != nil {
		autoReplyService = autoreply.NewAutoReplyService(
			autoReplyGlobalStorage,
			autoReplyConversationStorage,
			conversationStorage,
			agentAssistService,
			ingestionService,
		)
		ingestionService.SetAutoReplyService(autoReplyService)
		log.Println("Auto-reply service initialized successfully")
	}

	// Initialize default admin user if it doesn't exist
	if err := userStorage.InitDefaultAdmin(); err != nil {
		log.Printf("Warning: Failed to initialize default admin user: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userStorage)
	conversationHandler := handlers.NewConversationHandler(ingestionService, userStorage)
	ruleHandler := handlers.NewRuleHandler(ruleStorage)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, ingestionService, userStorage)
	productHandler := handlers.NewProductHandler(productStorage, embeddingService)
	memoryHandler := handlers.NewMemoryHandler(memoryStorage)
	
	var agentAssistHandler *handlers.AgentAssistHandler
	if agentAssistService != nil {
		agentAssistHandler = handlers.NewAgentAssistHandler(agentAssistService)
	}

	var autoReplyHandler *handlers.AutoReplyHandler
	if autoReplyService != nil {
		autoReplyHandler = handlers.NewAutoReplyHandler(
			autoReplyService,
			autoReplyGlobalStorage,
			autoReplyConversationStorage,
		)
	}

	// Set rule loader for analyzer if available
	if analyzer != nil {
		analyzer.SetRuleLoader(ruleStorage)
	}

	// Set up router
	router := gin.Default()

	// Middleware
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public routes (no JWT required)
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/customer-login", authHandler.CustomerLogin)
		}
	}

	// Protected API routes (JWT required)
	api = router.Group("/api")
	api.Use(jwtAuthMiddleware())
	{
		api.POST("/conversations", conversationHandler.CreateConversation)
		api.POST("/conversations/:id/messages", conversationHandler.SendMessage)
		api.GET("/conversations/:id", conversationHandler.GetConversation)
		api.GET("/conversations", conversationHandler.ListConversations)

		// Agent assist routes (agent only)
		if agentAssistHandler != nil {
			api.POST("/conversations/:id/suggestions", agentAssistHandler.GetSuggestions)
			api.GET("/conversations/:id/insights", agentAssistHandler.GetInsights)
		}

		// Rule management routes (admin only)
		rules := api.Group("/rules")
		rules.Use(adminMiddleware())
		{
			rules.GET("", ruleHandler.ListRules)
			rules.GET("/:id", ruleHandler.GetRule)
			rules.POST("", ruleHandler.CreateRule)
			rules.PUT("/:id", ruleHandler.UpdateRule)
			rules.DELETE("/:id", ruleHandler.DeleteRule)
		}

		// Analytics routes
		analyticsGroup := api.Group("/analytics")
		{
			analyticsGroup.GET("/leads", analyticsHandler.GetLeads)
			analyticsGroup.GET("/conversations/:id/win-probability", analyticsHandler.GetWinProbability)
			analyticsGroup.GET("/conversations/:id/churn-risk", analyticsHandler.GetChurnRisk)
			analyticsGroup.GET("/conversations/:id/trends", analyticsHandler.GetTrends)
			analyticsGroup.GET("/conversations/:id/clv", analyticsHandler.GetCLV)
			analyticsGroup.GET("/conversations/:id/sales-cycle", analyticsHandler.GetSalesCycle)
			analyticsGroup.GET("/dashboard", analyticsHandler.GetDashboard)

			// Admin-only analytics routes
			analyticsAdmin := analyticsGroup.Group("")
			analyticsAdmin.Use(adminMiddleware())
			{
				analyticsAdmin.GET("/conversations/:id/quality", analyticsHandler.GetQuality)
			}
		}

		// Product routes
		products := api.Group("/products")
		{
			// Public GET routes (authenticated users can view products)
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)

			// Admin-only management routes
			productsAdmin := products.Group("")
			productsAdmin.Use(adminMiddleware())
			{
				productsAdmin.POST("", productHandler.CreateProduct)
				productsAdmin.PUT("/:id", productHandler.UpdateProduct)
				productsAdmin.DELETE("/:id", productHandler.DeleteProduct)
			}
		}

		// Auto-reply routes
		if autoReplyHandler != nil {
			autoreplyGroup := api.Group("/autoreply")
			{
				// Global config (admin only)
				autoreplyGlobal := autoreplyGroup.Group("/global")
				autoreplyGlobal.Use(adminMiddleware())
				{
					autoreplyGlobal.GET("", autoReplyHandler.GetGlobalAutoReply)
					autoreplyGlobal.PUT("", autoReplyHandler.UpdateGlobalAutoReply)
				}

				// Conversation config (agent/admin)
				api.GET("/conversations/:id/autoreply", autoReplyHandler.GetConversationAutoReply)
				api.PUT("/conversations/:id/autoreply", autoReplyHandler.UpdateConversationAutoReply)

				// Test auto-reply (admin only)
				api.POST("/conversations/:id/autoreply/test", autoReplyHandler.TestAutoReply)
			}
		}

		// Customer memory routes (admin only)
		memories := api.Group("/memories")
		memories.Use(adminMiddleware())
		{
			memories.GET("", memoryHandler.ListMemories)
			memories.GET("/:id", memoryHandler.GetMemory)
			memories.POST("", memoryHandler.CreateMemory)
			memories.PUT("/:id", memoryHandler.UpdateMemory)
			memories.DELETE("/:id", memoryHandler.DeleteMemory)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	fmt.Printf("Server running on port %s\n", port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("%s %s %d %v", method, path, status, latency)
	}
}

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Log extracted claims for debugging
		log.Printf("[JWT] Extracted claims - UserID: %s, TenantID: %s, Role: %s", claims.UserID, claims.TenantID, claims.Role)

		// Validate that tenant_id is present in claims
		if claims.TenantID == "" {
			log.Printf("[JWT] WARNING: TenantID is empty in token claims for UserID: %s", claims.UserID)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is missing tenant_id. Please log out and log in again.",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

