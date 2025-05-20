package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Chatroom represents a chat room in the system
type Chatroom struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	CreatedBy uint               `bson:"created_by" json:"created_by"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	Members   []ChatroomMember   `bson:"members" json:"members"`
}

// ChatroomMember represents a user in a chatroom
type ChatroomMember struct {
	UserID   uint      `bson:"user_id" json:"user_id"`
	Username string    `bson:"username" json:"username"`
	JoinedAt time.Time `bson:"joined_at" json:"joined_at"`
}

// Message represents a message in a chatroom
type Message struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChatroomID  primitive.ObjectID `bson:"chatroom_id" json:"chatroom_id"`
	SenderID    uint               `bson:"sender_id" json:"sender_id"`
	SenderName  string             `bson:"sender_name" json:"sender_name"`
	MessageType string             `bson:"message_type" json:"message_type"` // text, picture, audio, video, etc.
	TextContent string             `bson:"text_content,omitempty" json:"text_content,omitempty"`
	MediaURL    string             `bson:"media_url,omitempty" json:"media_url,omitempty"`
	SentAt      time.Time          `bson:"sent_at" json:"sent_at"`
}

// ChatroomResponse is a struct for returning chatroom data
type ChatroomResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	CreatedBy uint             `json:"created_by"`
	CreatedAt time.Time        `json:"created_at"`
	Members   []ChatroomMember `json:"members"`
}

// ToResponse converts a Chatroom to a ChatroomResponse
func (c *Chatroom) ToResponse() ChatroomResponse {
	return ChatroomResponse{
		ID:        c.ID.Hex(),
		Name:      c.Name,
		CreatedBy: c.CreatedBy,
		CreatedAt: c.CreatedAt,
		Members:   c.Members,
	}
}

// MessageResponse is a struct for returning message data
type MessageResponse struct {
	ID          string    `json:"id"`
	ChatroomID  string    `json:"chatroom_id"`
	SenderID    uint      `json:"sender_id"`
	SenderName  string    `json:"sender_name"`
	MessageType string    `json:"message_type"`
	TextContent string    `json:"text_content,omitempty"`
	MediaURL    string    `json:"media_url,omitempty"`
	SentAt      time.Time `json:"sent_at"`
}

// ToResponse converts a Message to a MessageResponse
func (m *Message) ToResponse() MessageResponse {
	return MessageResponse{
		ID:          m.ID.Hex(),
		ChatroomID:  m.ChatroomID.Hex(),
		SenderID:    m.SenderID,
		SenderName:  m.SenderName,
		MessageType: m.MessageType,
		TextContent: m.TextContent,
		MediaURL:    m.MediaURL,
		SentAt:      m.SentAt,
	}
}
