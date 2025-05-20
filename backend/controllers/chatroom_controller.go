package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ginchat/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// ChatroomController handles chatroom-related requests
type ChatroomController struct {
	DB       *gorm.DB
	MongoDB  *mongo.Database
	UserColl *mongo.Collection
	ChatColl *mongo.Collection
	MsgColl  *mongo.Collection
}

// NewChatroomController creates a new ChatroomController
func NewChatroomController(db *gorm.DB, mongodb *mongo.Database) *ChatroomController {
	return &ChatroomController{
		DB:       db,
		MongoDB:  mongodb,
		ChatColl: mongodb.Collection("chatrooms"),
		MsgColl:  mongodb.Collection("messages"),
	}
}

// CreateChatroomRequest represents the request body for creating a chatroom
type CreateChatroomRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100"`
}

// CreateChatroom handles chatroom creation
func (cc *ChatroomController) CreateChatroom(c *gin.Context) {
	var req CreateChatroomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	username, _ := c.Get("username")

	// Check if chatroom with the same name already exists
	var count int64
	count, err := cc.ChatColl.CountDocuments(context.Background(), bson.M{"name": req.Name}, options.Count())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check chatroom existence"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Chatroom with this name already exists"})
		return
	}

	// Create new chatroom
	chatroom := models.Chatroom{
		ID:        primitive.NewObjectID(),
		Name:      req.Name,
		CreatedBy: userID.(uint),
		CreatedAt: time.Now(),
		Members: []models.ChatroomMember{
			{
				UserID:   userID.(uint),
				Username: username.(string),
				JoinedAt: time.Now(),
			},
		},
	}

	// Save chatroom to MongoDB
	_, err = cc.ChatColl.InsertOne(context.Background(), chatroom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chatroom"})
		return
	}

	// Return chatroom data
	c.JSON(http.StatusCreated, gin.H{
		"chatroom": chatroom.ToResponse(),
	})
}

// GetChatrooms handles getting all chatrooms
func (cc *ChatroomController) GetChatrooms(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Find all chatrooms
	cursor, err := cc.ChatColl.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chatrooms"})
		return
	}
	defer cursor.Close(context.Background())

	// Decode chatrooms
	var chatrooms []models.Chatroom
	if err := cursor.All(context.Background(), &chatrooms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode chatrooms"})
		return
	}

	// Convert to response format
	var response []models.ChatroomResponse
	for _, chatroom := range chatrooms {
		response = append(response, chatroom.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"chatrooms": response,
	})
}

// JoinChatroom handles joining a chatroom
func (cc *ChatroomController) JoinChatroom(c *gin.Context) {
	// Get chatroom ID from URL
	chatroomID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chatroom ID"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	username, _ := c.Get("username")

	// Check if chatroom exists
	var chatroom models.Chatroom
	err = cc.ChatColl.FindOne(context.Background(), bson.M{"_id": chatroomID}).Decode(&chatroom)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chatroom not found"})
		return
	}

	// Check if user is already a member
	for _, member := range chatroom.Members {
		if member.UserID == userID.(uint) {
			c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this chatroom"})
			return
		}
	}

	// Add user to chatroom members
	_, err = cc.ChatColl.UpdateOne(
		context.Background(),
		bson.M{"_id": chatroomID},
		bson.M{
			"$push": bson.M{
				"members": models.ChatroomMember{
					UserID:   userID.(uint),
					Username: username.(string),
					JoinedAt: time.Now(),
				},
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join chatroom"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined chatroom successfully"})
}
