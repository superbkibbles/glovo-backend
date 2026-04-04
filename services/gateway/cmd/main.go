package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/pkg/logger"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
	"github.com/mendmzury/food-delivery/services/gateway/internal/handlers"
	"github.com/mendmzury/food-delivery/services/gateway/internal/middleware"
	"github.com/mendmzury/food-delivery/services/gateway/internal/websocket"

	_ "github.com/mendmzury/food-delivery/services/gateway/docs" // Swagger docs
)

// @title Food Delivery API
// @version 1.0
// @description Food Delivery Backend API - Restaurants, orders, delivery
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("gateway")

	log.Info().Msg("Starting API Gateway...")

	grpcClients, err := grpc.NewClients(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to gRPC services")
	}
	defer grpcClients.Close()

	wsHub := websocket.NewHub()
	go wsHub.Run()

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(100, time.Minute))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login(cfg, grpcClients))
			auth.POST("/refresh-token", handlers.RefreshToken(cfg, grpcClients))

			// Customer signup (no auth)
			auth.POST("/signup/phone", handlers.SignUpWithPhone(cfg, grpcClients))
			auth.POST("/signup/google", handlers.SignUpWithGoogle(cfg, grpcClients))
			auth.POST("/signup/email", handlers.SignUpWithEmail(cfg, grpcClients))

			// Customer signin - OTP only (no auth)
			auth.POST("/signin/phone", handlers.SignInWithPhone(cfg, grpcClients))
			auth.POST("/signin/email", handlers.SignInWithEmail(cfg, grpcClients))
			auth.POST("/signin/google", handlers.SignInWithGoogle(cfg, grpcClients))

			// OTP delivery (no auth)
			auth.POST("/send-otp", handlers.SendOTP(cfg, grpcClients))
		}

		internal := api.Group("/internal")
		{
			internal.POST("/notifications/broadcast", handlers.BroadcastNotification(cfg, wsHub))
			internal.GET("/notifications/check-connection", handlers.CheckWebSocketConnection(cfg, wsHub))
		}

		api.GET("/ws/notifications", websocket.HandleWebSocket(wsHub, cfg))

		protected := api.Group("")
		protected.Use(middleware.Auth(cfg))
		{
			protected.GET("/auth/me", handlers.GetCurrentUser(cfg, grpcClients))
			protected.PUT("/auth/profile", handlers.UpdateMyProfile(cfg, grpcClients))
			protected.POST("/auth/change-password", handlers.ChangePassword(cfg, grpcClients))
			protected.POST("/auth/logout", handlers.Logout(cfg, grpcClients))

			users := protected.Group("/users")
			{
				users.GET("", handlers.ListUsers(cfg, grpcClients))
				users.POST("", handlers.CreateUser(cfg, grpcClients))
				users.GET("/:id", handlers.GetUser(cfg, grpcClients))
				users.PUT("/:id", handlers.UpdateUser(cfg, grpcClients))
				users.DELETE("/:id", handlers.DeleteUser(cfg, grpcClients))
			}

			roles := protected.Group("/roles")
			{
				roles.GET("", handlers.ListRoles(cfg, grpcClients))
				roles.GET("/:id", handlers.GetRole(cfg, grpcClients))
			}

			restaurants := protected.Group("/restaurants")
			{
				restaurants.GET("", handlers.ListRestaurants(cfg, grpcClients))
				restaurants.POST("", handlers.CreateRestaurant(cfg, grpcClients))
				restaurants.GET("/:id", handlers.GetRestaurant(cfg, grpcClients))
				restaurants.PUT("/:id", handlers.UpdateRestaurant(cfg, grpcClients))
				restaurants.DELETE("/:id", handlers.DeleteRestaurant(cfg, grpcClients))
				restaurants.GET("/:id/categories", handlers.ListCategories(cfg, grpcClients))
				restaurants.POST("/categories", handlers.CreateCategory(cfg, grpcClients))
				restaurants.GET("/menu-items", handlers.ListMenuItems(cfg, grpcClients))
				restaurants.POST("/menu-items", handlers.CreateMenuItem(cfg, grpcClients))
			}

			orders := protected.Group("/orders")
			{
				orders.GET("", handlers.ListOrders(cfg, grpcClients))
				orders.POST("", handlers.CreateOrder(cfg, grpcClients))
				orders.GET("/:id", handlers.GetOrder(cfg, grpcClients))
				orders.POST("/:id/accept", handlers.AcceptOrder(cfg, grpcClients))
				orders.PUT("/:id/status", handlers.UpdateOrderStatus(cfg, grpcClients))
				orders.POST("/:id/cancel", handlers.CancelOrder(cfg, grpcClients))
			}

			delivery := protected.Group("/delivery")
			{
				delivery.GET("/assignments", handlers.ListAssignments(cfg, grpcClients))
				delivery.GET("/assignments/by", handlers.GetAssignment(cfg, grpcClients))
				delivery.POST("/assign", handlers.AssignDriver(cfg, grpcClients))
				delivery.POST("/location", handlers.UpdateDriverLocation(cfg, grpcClients))
				delivery.POST("/assignments/:id/complete", handlers.CompleteDelivery(cfg, grpcClients))
			delivery.GET("/assignments/:id/tracking-history", handlers.GetTrackingHistory(cfg, grpcClients))
			delivery.GET("/driver-locations", handlers.ListDriverLocations(cfg, grpcClients))
			}

			wsAdmin := protected.Group("/websocket")
			{
				wsAdmin.GET("/connections", handlers.ListConnections(cfg, wsHub, grpcClients))
			}

			settings := protected.Group("/settings")
			{
				settings.GET("/commission", handlers.GetCommission(cfg, grpcClients))
				settings.PUT("/commission", handlers.UpdateCommission(cfg, grpcClients))
				settings.GET("/operating-areas", handlers.ListOperatingAreas(cfg, grpcClients))
				settings.POST("/operating-areas", handlers.CreateOperatingArea(cfg, grpcClients))
				settings.PUT("/operating-areas/:id", handlers.UpdateOperatingArea(cfg, grpcClients))
				settings.DELETE("/operating-areas/:id", handlers.DeleteOperatingArea(cfg, grpcClients))
			}
		}
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.GatewayHTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("address", server.Addr).Msg("API Gateway started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down API Gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	log.Info().Msg("API Gateway stopped")
}
