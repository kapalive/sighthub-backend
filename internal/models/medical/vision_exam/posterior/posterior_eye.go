package posterior

type PosteriorEye struct {
	IDPosteriorEye          int64   `gorm:"column:id_posterior_eye;primaryKey;autoIncrement" json:"id_posterior_eye"`
	InfoDirect              *bool   `gorm:"column:info_direct;default:false" json:"info_direct"`
	InfoBio                 *bool   `gorm:"column:info_bio;default:false" json:"info_bio"`
	Info90d                 *bool   `gorm:"column:info_90d;default:false" json:"info_90d"`
	InfoOptomap             *bool   `gorm:"column:info_optomap;default:false" json:"info_optomap"`
	InfoRha                 *bool   `gorm:"column:info_rha;default:false" json:"info_rha"`
	InfoOther               *string `gorm:"column:info_other;type:text" json:"info_other"`
	MedicationPatientEducated *bool `gorm:"column:medication_patient_educated;default:false" json:"medication_patient_educated"`
	MedicationIlationDeclined *bool `gorm:"column:medication_ilation_declined;default:false" json:"medication_ilation_declined"`
	MedicationParemyd         *bool `gorm:"column:medication_paremyd;default:false" json:"medication_paremyd"`
	MedicationAtropine        *bool `gorm:"column:medication_atropine;default:false" json:"medication_atropine"`
	MedicationTropicamide     *bool `gorm:"column:medication_tropicamide;default:false" json:"medication_tropicamide"`
	MedicationCyclopentolate  *bool `gorm:"column:medication_cyclopentolate;default:false" json:"medication_cyclopentolate"`
	MedicationHomatropine     *bool `gorm:"column:medication_homatropine;default:false" json:"medication_homatropine"`
	MedicationPhenylephrine   *bool `gorm:"column:medication_phenylephrine;default:false" json:"medication_phenylephrine"`
	MedicationRha             *bool `gorm:"column:medication_rha;default:false" json:"medication_rha"`
	TimeDilated               *string `gorm:"column:time_dilated;type:time" json:"time_dilated"`
	Other                     *string `gorm:"column:other;size:100" json:"other"`
	FindingsPosteriorID       int64   `gorm:"column:findings_posterior_id;not null;uniqueIndex" json:"findings_posterior_id"`
	CupDiscRatioPosteriorID   int64   `gorm:"column:cup_disc_ratio_posterior_id;not null;uniqueIndex" json:"cup_disc_ratio_posterior_id"`
	Note                      *string `gorm:"column:note;type:text" json:"note"`
	AddDrawing                *string `gorm:"column:add_drawing;type:text" json:"add_drawing"`
	EyeExamID                 int64   `gorm:"column:eye_exam_id;not null" json:"eye_exam_id"`

	Findings     FindingsPosterior     `gorm:"foreignKey:IDFindingsPosterior;references:FindingsPosteriorID" json:"findings"`
	CupDiscRatio CupDiscRatioPosterior `gorm:"foreignKey:IDCupDiscRatioPosterior;references:CupDiscRatioPosteriorID" json:"cup_disc_ratio"`
}

func (PosteriorEye) TableName() string { return "posterior_eye" }
