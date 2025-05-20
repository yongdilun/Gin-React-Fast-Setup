package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ginchat/controllers"
	"github.com/ginchat/middleware"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(r *gin.Engine, db *gorm.DB, mongodb *mongo.Database, logger *logrus.Logger) {
	// Create controllers
	userController := controllers.NewUserController(db)
	chatroomController := controllers.NewChatroomController(db, mongodb)
	messageController := controllers.NewMessageController(db, mongodb)
	websocketController := controllers.NewWebSocketController(logger)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Swagger documentation is set up in main.go

	// API routes
	api := r.Group("/api")
	{
		// Auth routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userController.Register)
			auth.POST("/login", userController.Login)
		}

		// Protected routes (auth required)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			protected.POST("/auth/logout", userController.Logout)

			// Chatroom routes
			protected.GET("/chatrooms", chatroomController.GetChatrooms)
			protected.POST("/chatrooms", chatroomController.CreateChatroom)
			protected.POST("/chatrooms/:id/join", chatroomController.JoinChatroom)

			// Message routes
			protected.GET("/chatrooms/:id/messages", messageController.GetMessages)
			protected.POST("/chatrooms/:id/messages", messageController.SendMessage)

			// WebSocket route
			protected.GET("/ws", websocketController.HandleConnection)
		}
	}
}
