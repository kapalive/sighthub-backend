// internal/models/general/communication_type.go
package general

import "fmt"

type CommunicationType struct {
	CommunicationTypeID int    `gorm:"column:communication_type_id;primaryKey"                                       json:"communication_type_id"`
	CommunicationType   string `gorm:"column:communication_type;type:varchar(50);not null;uniqueIndex:uniq_comm_type" json:"communication_type"`
}

func (CommunicationType) TableName() string { return "communication_type" }

func (c *CommunicationType) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"communication_type_id": c.CommunicationTypeID,
		"communication_type":    c.CommunicationType,
	}
}

func (c *CommunicationType) String() string {
	return fmt.Sprintf("<CommunicationType %s>", c.CommunicationType)
}
