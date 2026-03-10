package preliminary
type NearPointTesting struct {
	IDNearPointTesting    int64  `gorm:"column:id_near_point_testing;primaryKey;autoIncrement" json:"id_near_point_testing"`
	DistPhoriaTestingID   *int64 `gorm:"column:dist_phoria_testing_id"                         json:"dist_phoria_testing_id,omitempty"`
	NearPhoriaTestingID   *int64 `gorm:"column:near_phoria_testing_id"                         json:"near_phoria_testing_id,omitempty"`
	DistVergenceTestingID *int64 `gorm:"column:dist_vergence_testing_id"                       json:"dist_vergence_testing_id,omitempty"`
	NearVergenceTestingID *int64 `gorm:"column:near_vergence_testing_id"                       json:"near_vergence_testing_id,omitempty"`
	AccommodationID       *int64 `gorm:"column:accommodation_id"                               json:"accommodation_id,omitempty"`

	DistPhoria    *DistPhoriaTest    `gorm:"foreignKey:DistPhoriaTestingID;references:IDDistPhoriaTest"     json:"-"`
	NearPhoria    *NearPhoriaTest    `gorm:"foreignKey:NearPhoriaTestingID;references:IDNearPhoriaTest"     json:"-"`
	DistVergence  *DistVergenceTest  `gorm:"foreignKey:DistVergenceTestingID;references:IDDistVergenceTest" json:"-"`
	NearVergence  *NearVergenceTest  `gorm:"foreignKey:NearVergenceTestingID;references:IDNearVergenceTest" json:"-"`
	Accommodation *Accommodation     `gorm:"foreignKey:AccommodationID;references:IDAccommodation"          json:"-"`
}
func (NearPointTesting) TableName() string { return "near_point_testing" }
func (n *NearPointTesting) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_near_point_testing":    n.IDNearPointTesting,
		"dist_phoria_testing_id":   n.DistPhoriaTestingID,
		"near_phoria_testing_id":   n.NearPhoriaTestingID,
		"dist_vergence_testing_id": n.DistVergenceTestingID,
		"near_vergence_testing_id": n.NearVergenceTestingID,
		"accommodation_id":         n.AccommodationID,
	}
}
