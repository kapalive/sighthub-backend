// internal/models/general/chat_message.go
package general

import "time"

// ChatMessage ⇄ chat_messages
type ChatMessage struct {
	IDChatMessage int64     `gorm:"column:id_chat_message;primaryKey;autoIncrement" json:"id_chat_message"`
	RoomName      string    `gorm:"column:room_name;type:varchar(255);not null"      json:"room_name"`
	SenderID      int64     `gorm:"column:sender_id;not null"                        json:"sender_id"`
	Message       string    `gorm:"column:message;type:text;not null"                json:"message"`
	Timestamp     time.Time `gorm:"column:timestamp;type:timestamptz;not null;default:now()" json:"timestamp"`
}

func (ChatMessage) TableName() string { return "chat_messages" }

func (c *ChatMessage) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_chat_message": c.IDChatMessage,
		"room_name":       c.RoomName,
		"sender_id":       c.SenderID,
		"message":         c.Message,
		"timestamp":       c.Timestamp.Format(time.RFC3339),
	}
}
