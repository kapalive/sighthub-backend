// internal/models/general/race.go
package general

import "fmt"

type Race struct {
	IDRace   int    `gorm:"column:id_race;primaryKey"                  json:"id_race"`
	RaceName string `gorm:"column:race_name;type:varchar(50);not null" json:"race_name"`
}

func (Race) TableName() string { return "race" }

func (r *Race) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_race":   r.IDRace,
		"race_name": r.RaceName,
	}
}

func (r *Race) String() string {
	return fmt.Sprintf("<Race %s>", r.RaceName)
}
