package preliminary
type Accommodation struct {
	IDAccommodation      int     `gorm:"column:id_accommodation;primaryKey;autoIncrement" json:"id_accommodation"`
	Pra1                 *string `gorm:"column:pra1;type:varchar(50)"                     json:"pra1,omitempty"`
	Nra1                 *string `gorm:"column:nra1;type:varchar(50)"                     json:"nra1,omitempty"`
	Pra2                 *string `gorm:"column:pra2;type:varchar(50)"                     json:"pra2,omitempty"`
	Nra2                 *string `gorm:"column:nra2;type:varchar(50)"                     json:"nra2,omitempty"`
	MemOd                *string `gorm:"column:mem_od;type:varchar(50)"                   json:"mem_od,omitempty"`
	MemOs                *string `gorm:"column:mem_os;type:varchar(50)"                   json:"mem_os,omitempty"`
	Baf                  *string `gorm:"column:baf;type:varchar(50)"                      json:"baf,omitempty"`
	VergenceFacilityCpm  *string `gorm:"column:vergence_facility_cpm;type:varchar(50)"    json:"vergence_facility_cpm,omitempty"`
	VergenceFacilityWith *string `gorm:"column:vergence_facility_with;type:varchar(50)"   json:"vergence_facility_with,omitempty"`
	PushUpOd             *string `gorm:"column:push_up_od;type:varchar(50)"               json:"push_up_od,omitempty"`
	PushUpOs             *string `gorm:"column:push_up_os;type:varchar(50)"               json:"push_up_os,omitempty"`
	PushUpOu             *string `gorm:"column:push_up_ou;type:varchar(50)"               json:"push_up_ou,omitempty"`
	SlowWith             *bool   `gorm:"column:slow_with"                                 json:"slow_with,omitempty"`
}
func (Accommodation) TableName() string { return "accommodation" }
func (a *Accommodation) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_accommodation": a.IDAccommodation,
		"pra1": a.Pra1, "nra1": a.Nra1, "pra2": a.Pra2, "nra2": a.Nra2,
		"mem_od": a.MemOd, "mem_os": a.MemOs, "baf": a.Baf,
		"vergence_facility_cpm": a.VergenceFacilityCpm, "vergence_facility_with": a.VergenceFacilityWith,
		"push_up_od": a.PushUpOd, "push_up_os": a.PushUpOs, "push_up_ou": a.PushUpOu,
		"slow_with": a.SlowWith,
	}
}
