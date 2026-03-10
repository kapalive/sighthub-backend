// internal/models/general/chat_room_participant.go
package general

// ChatRoomParticipant ⇄ chat_room_participants
type ChatRoomParticipant struct {
	IDChatRoomParticipant int64  `gorm:"column:id_chat_room_participant;primaryKey;autoIncrement" json:"id_chat_room_participant"`
	RoomName              string `gorm:"column:room_name;type:varchar(255);not null"              json:"room_name"`
	ParticipantID         int64  `gorm:"column:participant_id;not null"                           json:"participant_id"`
}

func (ChatRoomParticipant) TableName() string { return "chat_room_participants" }

func (c *ChatRoomParticipant) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_chat_room_participant": c.IDChatRoomParticipant,
		"room_name":                c.RoomName,
		"participant_id":           c.ParticipantID,
	}
}
