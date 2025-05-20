package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ginchat/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// MessageController handles message-related requests
type MessageController struct {
	DB       *gorm.DB
	MongoDB  *mongo.Database
	ChatColl *mongo.Collection
	MsgColl  *mongo.Collection
}

// NewMessageController creates a new MessageController
func NewMessageController(db *gorm.DB, mongodb *mongo.Database) *MessageController {
	return &MessageController{
		DB:       db,
		MongoDB:  mongodb,
		ChatColl: mongodb.Collection("chatrooms"),
		MsgColl:  mongodb.Collection("messages"),
	}
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	MessageType string `json:"message_type" binding:"required,oneof=text picture audio video text_and_picture text_and_audio text_and_video"`
	TextContent string `json:"text_content"`
	MediaURL    string `json:"media_url"`
}

// SendMessage handles sending a message to a chatroom
func (mc *MessageController) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	err = mc.ChatColl.FindOne(context.Background(), bson.M{"_id": chatroomID}).Decode(&chatroom)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chatroom not found"})
		return
	}

	// Check if user is a member of the chatroom
	isMember := false
	for _, member := range chatroom.Members {
		if member.UserID == userID.(uint) {
			isMember = true
			break
		}
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this chatroom"})
		return
	}

	// Create new message
	message := models.Message{
		ID:          primitive.NewObjectID(),
		ChatroomID:  chatroomID,
		SenderID:    userID.(uint),
		SenderName:  username.(string),
		MessageType: req.MessageType,
		TextContent: req.TextContent,
		MediaURL:    req.MediaURL,
		SentAt:      time.Now(),
	}

	// Save message to MongoDB
	_, err = mc.MsgColl.InsertOne(context.Background(), message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Return message data
	c.JSON(http.StatusCreated, gin.H{
		"message": message.ToResponse(),
	})
}

// GetMessages handles getting messages from a chatroom
func (mc *MessageController) GetMessages(c *gin.Context) {
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

	// Check if chatroom exists
	var chatroom models.Chatroom
	err = mc.ChatColl.FindOne(context.Background(), bson.M{"_id": chatroomID}).Decode(&chatroom)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chatroom not found"})
		return
	}

	// Check if user is a member of the chatroom
	isMember := false
	for _, member := range chatroom.Members {
		if member.UserID == userID.(uint) {
			isMember = true
			break
		}
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this chatroom"})
		return
	}

	// Get limit from query parameters
	limit := 50 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		// Try to parse the limit parameter
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Find messages for the chatroom
	findOptions := options.Find().SetSort(bson.M{"sent_at": -1}).SetLimit(int64(limit))
	cursor, err := mc.MsgColl.Find(context.Background(), bson.M{"chatroom_id": chatroomID}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}
	defer cursor.Close(context.Background())

	// Decode messages
	var messages []models.Message
	if err := cursor.All(context.Background(), &messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode messages"})
		return
	}

	// Convert to response format
	var response []models.MessageResponse
	for _, message := range messages {
		response = append(response, message.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": response,
	})
}
