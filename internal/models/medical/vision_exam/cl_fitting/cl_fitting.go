package cl_fitting

// ClFitting ↔ table: cl_fitting
type ClFitting struct {
	IDClFitting    int64   `gorm:"column:id_cl_fitting;primaryKey;autoIncrement"  json:"id_cl_fitting"`
	Fitting1ID     int64   `gorm:"column:fitting_1_id;not null;uniqueIndex"       json:"fitting_1_id"`
	Fitting2ID     *int64  `gorm:"column:fitting_2_id;uniqueIndex"                json:"fitting_2_id,omitempty"`
	Fitting3ID     *int64  `gorm:"column:fitting_3_id;uniqueIndex"                json:"fitting_3_id,omitempty"`
	FirstTrialID   int64   `gorm:"column:first_trial_id;not null;uniqueIndex"     json:"first_trial_id"`
	SecondTrialID  *int64  `gorm:"column:second_trial_id;uniqueIndex"             json:"second_trial_id,omitempty"`
	ThirdTrialID   *int64  `gorm:"column:third_trial_id;uniqueIndex"              json:"third_trial_id,omitempty"`
	GasPermeableID *int64  `gorm:"column:gas_permeable_id;uniqueIndex"            json:"gas_permeable_id,omitempty"`
	DrNote         *string `gorm:"column:dr_note;type:text"                       json:"dr_note,omitempty"`
	EyeExamID      int64   `gorm:"column:eye_exam_id;not null"                    json:"eye_exam_id"`

	Fitting1     *Fitting1     `gorm:"foreignKey:Fitting1ID;references:IDFitting1"         json:"-"`
	Fitting2     *Fitting2     `gorm:"foreignKey:Fitting2ID;references:IDFitting2"         json:"-"`
	Fitting3     *Fitting3     `gorm:"foreignKey:Fitting3ID;references:IDFitting3"         json:"-"`
	FirstTrial   *FirstTrial   `gorm:"foreignKey:FirstTrialID;references:IDFirstTrial"     json:"-"`
	SecondTrial  *SecondTrial  `gorm:"foreignKey:SecondTrialID;references:IDSecondTrial"   json:"-"`
	ThirdTrial   *ThirdTrial   `gorm:"foreignKey:ThirdTrialID;references:IDThirdTrial"     json:"-"`
	GasPermeable *GasPermeable `gorm:"foreignKey:GasPermeableID;references:IDGasPermeable" json:"-"`
}
func (ClFitting) TableName() string { return "cl_fitting" }
func (c *ClFitting) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_cl_fitting": c.IDClFitting,
		"fitting_1_id": c.Fitting1ID, "fitting_2_id": c.Fitting2ID, "fitting_3_id": c.Fitting3ID,
		"first_trial_id": c.FirstTrialID, "second_trial_id": c.SecondTrialID,
		"third_trial_id": c.ThirdTrialID, "gas_permeable_id": c.GasPermeableID,
		"dr_note": c.DrNote, "eye_exam_id": c.EyeExamID,
	}
	if c.Fitting1 != nil     { m["fitting_1"]     = c.Fitting1.ToMap() }
	if c.Fitting2 != nil     { m["fitting_2"]     = c.Fitting2.ToMap() }
	if c.Fitting3 != nil     { m["fitting_3"]     = c.Fitting3.ToMap() }
	if c.FirstTrial != nil   { m["first_trial"]   = c.FirstTrial.ToMap() }
	if c.SecondTrial != nil  { m["second_trial"]  = c.SecondTrial.ToMap() }
	if c.ThirdTrial != nil   { m["third_trial"]   = c.ThirdTrial.ToMap() }
	if c.GasPermeable != nil { m["gas_permeable"] = c.GasPermeable.ToMap() }
	return m
}
