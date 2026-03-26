package preliminary_service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	empLoginModel "sighthub-backend/internal/models/auth"
	empModel      "sighthub-backend/internal/models/employees"
	preliminary   "sighthub-backend/internal/models/medical/vision_exam/preliminary"
	visionModel   "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
	"sighthub-backend/pkg/defaults"
)

// ─── Service ─────────────────────────────────────────────────────────────────

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func irisColorOrDefault(s *string) string {
	if s == nil || *s == "" {
		return "n/a"
	}
	return *s
}

// ─── Local prescription stubs ─────────────────────────────────────────────────

type rxPrescription struct {
	IDPatientPrescription int64      `gorm:"column:id_patient_prescription;primaryKey"`
	PrescriptionDate      *time.Time `gorm:"column:prescription_date;type:date"`
	PatientID             int64      `gorm:"column:patient_id"`
}

func (rxPrescription) TableName() string { return "patient_prescription" }

type rxGlasses struct {
	IDGlassesPrescription int64   `gorm:"column:id_glasses_prescription;primaryKey"`
	PrescriptionID        int64   `gorm:"column:prescription_id"`
	OdSph                 *string `gorm:"column:od_sph"`
	OsSph                 *string `gorm:"column:os_sph"`
	OdCyl                 *string `gorm:"column:od_cyl"`
	OsCyl                 *string `gorm:"column:os_cyl"`
	OdAxis                *string `gorm:"column:od_axis"`
	OsAxis                *string `gorm:"column:os_axis"`
	OdAdd                 *string `gorm:"column:od_add"`
	OsAdd                 *string `gorm:"column:os_add"`
	OdHPrism              *string `gorm:"column:od_h_prism"`
	OsHPrism              *string `gorm:"column:os_h_prism"`
}

func (rxGlasses) TableName() string { return "glasses_prescription" }

type rxContactLens struct {
	IDContactLensPrescription int64   `gorm:"column:id_contact_lens_prescription;primaryKey"`
	PrescriptionID            int64   `gorm:"column:prescription_id"`
	OdContLens                *string `gorm:"column:od_cont_lens"`
	OsContLens                *string `gorm:"column:os_cont_lens"`
	OdBc                      *string `gorm:"column:od_bc"`
	OsBc                      *string `gorm:"column:os_bc"`
	OdDia                     *string `gorm:"column:od_dia"`
	OsDia                     *string `gorm:"column:os_dia"`
	OdPwr                     *string `gorm:"column:od_pwr"`
	OsPwr                     *string `gorm:"column:os_pwr"`
	OdCyl                     *string `gorm:"column:od_cyl"`
	OsCyl                     *string `gorm:"column:os_cyl"`
	OdAxis                    *string `gorm:"column:od_axis"`
	OsAxis                    *string `gorm:"column:os_axis"`
	OdAdd                     *string `gorm:"column:od_add"`
	OsAdd                     *string `gorm:"column:os_add"`
}

func (rxContactLens) TableName() string { return "contact_lens_prescription" }

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *Service) getEmployee(username string) (*empModel.Employee, error) {
	var login empLoginModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func (s *Service) getExam(examID int64) (*visionModel.EyeExam, error) {
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	return &exam, nil
}

func boolDeref(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func i64(i int) int64 { return int64(i) }

// ─── Input types ──────────────────────────────────────────────────────────────

type UnaidedVADistanceInput struct {
	Od20 *string `json:"od_20"`
	Os20 *string `json:"os_20"`
	Ou20 *string `json:"ou_20"`
}

type UnaidedPHDistanceInput struct {
	Od20 *string `json:"od_20"`
	Os20 *string `json:"os_20"`
}

type UnaidedVANearInput struct {
	Od20 *string `json:"od_20"`
	Os20 *string `json:"os_20"`
	Ou20 *string `json:"ou_20"`
}

type AidedVADistanceInput struct {
	Od20 *string `json:"od_20"`
	Os20 *string `json:"os_20"`
	Ou20 *string `json:"ou_20"`
}

type AidedPHDistanceInput struct {
	Od20 *string `json:"od_20"`
	Os20 *string `json:"os_20"`
}

type AidedVANearInput struct {
	Od20 *string `json:"od_20"`
	Os20 *string `json:"os_20"`
	Ou20 *string `json:"ou_20"`
}

type ConfrontationInput struct {
	Od *string `json:"od"`
	Os *string `json:"os"`
}

type AutomatedInput struct {
	Od *string `json:"od"`
	Os *string `json:"os"`
}

type MotilityInput struct {
	Od *string `json:"od"`
	Os *string `json:"os"`
}

type PupilsInput struct {
	OdMmDim   *string `json:"od_mm_dim"`
	OdMmBright *string `json:"od_mm_bright"`
	OsMmDim   *string `json:"os_mm_dim"`
	OsMmBright *string `json:"os_mm_bright"`
	Perrla    *bool   `json:"perrla"`
	PerrlaText *string `json:"perrla_text"`
	Apd       *bool   `json:"apd"`
	ApdText   *string `json:"apd_text"`
}

type ColorVisionInput struct {
	Od1 *string `json:"od1"`
	Od2 *string `json:"od2"`
	Os1 *string `json:"os1"`
	Os2 *string `json:"os2"`
}

type BrucknerInput struct {
	Od          *string `json:"od"`
	Os          *string `json:"os"`
	GoodReflex  *bool   `json:"good_reflex"`
}

type AmslerGridInput struct {
	Od *string `json:"od"`
	Os *string `json:"os"`
}

type DistanceVonGraefeInput struct {
	HDistVgp *string `json:"h_dist_vgp"`
	VDistVgp *string `json:"v_dist_vgp"`
}

type NearVonGraefeInput struct {
	HNearVgp *string `json:"h_near_vgp"`
	VNearVgp *string `json:"v_near_vgp"`
}

type AutorefractorInput struct {
	OdSph  *string `json:"od_sph"`
	OsSph  *string `json:"os_sph"`
	OdCyl  *string `json:"od_cyl"`
	OsCyl  *string `json:"os_cyl"`
	OdAxis *string `json:"od_axis"`
	OsAxis *string `json:"os_axis"`
	Pd     *string `json:"pd"`
}

type AutoKeratometerInput struct {
	OdPw1  *string `json:"od_pw1"`
	OsPw1  *string `json:"os_pw1"`
	OdPw2  *string `json:"od_pw2"`
	OsPw2  *string `json:"os_pw2"`
	OdAxis *string `json:"od_axis"`
	OsAxis *string `json:"os_axis"`
}

type BloodPressureInput struct {
	Sbp *string `json:"sbp"`
	Dbp *string `json:"dbp"`
}

type DistPhoriaInput struct {
	Horiz      *string `json:"horiz"`
	Vert       *string `json:"vert"`
	HorizExo   *bool   `json:"horiz_exo"`
	HorizEso   *bool   `json:"horiz_eso"`
	HorizOrtho *bool   `json:"horiz_ortho"`
	VertRh     *bool   `json:"vert_rh"`
	VertLn     *bool   `json:"vert_ln"`
	VertOrtho  *bool   `json:"vert_ortho"`
}

type NearPhoriaInput struct {
	Horiz            *string `json:"horiz"`
	Vert             *string `json:"vert"`
	GradientRatio1   *string `json:"gradient_ratio1"`
	CalculatedRatio1 *string `json:"calculated_ratio1"`
	GradientRatio2   *string `json:"gradient_ratio2"`
	CalculatedRatio2 *string `json:"calculated_ratio2"`
	HorizExo         *bool   `json:"horiz_exo"`
	HorizEso         *bool   `json:"horiz_eso"`
	HorizOrtho       *bool   `json:"horiz_ortho"`
	VertRh           *bool   `json:"vert_rh"`
	VertLn           *bool   `json:"vert_ln"`
	VertOrtho        *bool   `json:"vert_ortho"`
}

type DistVergenceInput struct {
	Bi1 *string `json:"bi1"`
	Bo1 *string `json:"bo1"`
	Bi2 *string `json:"bi2"`
	Bo2 *string `json:"bo2"`
	Bi3 *string `json:"bi3"`
	Bo3 *string `json:"bo3"`
}

type NearVergenceInput struct {
	Bi1 *string `json:"bi1"`
	Bo1 *string `json:"bo1"`
	Bi2 *string `json:"bi2"`
	Bo2 *string `json:"bo2"`
	Bi3 *string `json:"bi3"`
	Bo3 *string `json:"bo3"`
}

type AccommodationInput struct {
	Pra1                 *string `json:"pra1"`
	Nra1                 *string `json:"nra1"`
	Pra2                 *string `json:"pra2"`
	Nra2                 *string `json:"nra2"`
	MemOd                *string `json:"mem_od"`
	MemOs                *string `json:"mem_os"`
	Baf                  *string `json:"baf"`
	VergenceFacilityCpm  *string `json:"vergence_facility_cpm"`
	VergenceFacilityWith *string `json:"vergence_facility_with"`
	PushUpOd             *string `json:"push_up_od"`
	PushUpOs             *string `json:"push_up_os"`
	PushUpOu             *string `json:"push_up_ou"`
	SlowWith             *bool   `json:"slow_with"`
}

type NearPointTestingInput struct {
	DistPhoriaTesting   *DistPhoriaInput   `json:"dist_phoria_testing"`
	NearPhoriaTesting   *NearPhoriaInput   `json:"near_phoria_testing"`
	DistVergenceTesting *DistVergenceInput `json:"dist_vergence_testing"`
	NearVergenceTesting *NearVergenceInput `json:"near_vergence_testing"`
	Accommodation       *AccommodationInput `json:"accommodation"`
}

type SavePreliminaryInput struct {
	UnaidedVADistance          *UnaidedVADistanceInput `json:"unaided_va_distance"`
	UnaidedPHDistance          *UnaidedPHDistanceInput `json:"unaided_ph_distance"`
	UnaidedVANear              *UnaidedVANearInput     `json:"unaided_va_near"`
	AidedVADistance            *AidedVADistanceInput   `json:"aided_va_distance"`
	AidedPHDistance            *AidedPHDistanceInput   `json:"aided_ph_distance"`
	AidedVANear                *AidedVANearInput       `json:"aided_va_near"`
	AidedByGlasses             *bool                   `json:"aided_by_glasses"`
	AidedByContacts            *bool                   `json:"aided_by_contacts"`
	Confrontation              *ConfrontationInput     `json:"confrontation"`
	Automated                  *AutomatedInput         `json:"automated"`
	Motility                   *MotilityInput          `json:"motility"`
	Pupils                     *PupilsInput            `json:"pupils"`
	ColorVision                *ColorVisionInput       `json:"color_vision"`
	DistanceCoverTest          *string                 `json:"distance_cover_test"`
	NearCoverTest              *string                 `json:"near_cover_test"`
	NpcTest                    *string                 `json:"npc_test"`
	Bruckner                   *BrucknerInput          `json:"bruckner"`
	AmslerGrid                 *AmslerGridInput        `json:"amsler_grid"`
	Worth4Dot                  *string                 `json:"worth_4_dot"`
	StereoVision               *string                 `json:"stereo_vision"`
	FixationDisparity          *string                 `json:"fixation_disparity"`
	DistanceVonGraefePhoria    *DistanceVonGraefeInput `json:"distance_von_graefe_phoria"`
	NearVonGraefePhoria        *NearVonGraefeInput     `json:"near_von_graefe_phoria"`
	NearPointTesting           *NearPointTestingInput  `json:"near_point_testing"`
	AutorefractorPreliminary   *AutorefractorInput     `json:"autorefractor_preliminary"`
	AutoKeratometerPreliminary *AutoKeratometerInput   `json:"auto_keratometer_preliminary"`
	BloodPressure              *BloodPressureInput     `json:"blood_pressure"`
	IrisColor                  *string                 `json:"iris_color"`
	Note                       *string                 `json:"note"`
}

type UpdatePreliminaryInput = SavePreliminaryInput

type RxFillInput struct {
	Glasses *int64 `json:"glasses"`
	Contact *int64 `json:"contact"`
}

type FillEntranceRxInput struct {
	RxData RxFillInput `json:"rx_data"`
}

type NearPointTestingStandaloneInput struct {
	NearPointTesting NearPointTestingInput `json:"near_point_testing"`
}

// ─── SavePreliminary ─────────────────────────────────────────────────────────

func (s *Service) SavePreliminary(username string, examID int64, input SavePreliminaryInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}

	if exam.EmployeeID != i64(emp.IDEmployee) {
		return nil, errors.New("you are not authorized to update this exam")
	}
	if exam.Passed {
		return nil, errors.New("cannot create preliminary for a completed exam")
	}

	var existing preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&existing).Error; err == nil {
		return nil, errors.New("preliminary already exists for this exam")
	}

	// NearPointTesting sub-records
	if input.NearPointTesting == nil {
		input.NearPointTesting = &NearPointTestingInput{}
	}
	if input.NearPointTesting.DistPhoriaTesting == nil {
		input.NearPointTesting.DistPhoriaTesting = &DistPhoriaInput{}
	}
	if input.NearPointTesting.NearPhoriaTesting == nil {
		input.NearPointTesting.NearPhoriaTesting = &NearPhoriaInput{}
	}
	if input.NearPointTesting.DistVergenceTesting == nil {
		input.NearPointTesting.DistVergenceTesting = &DistVergenceInput{}
	}
	if input.NearPointTesting.NearVergenceTesting == nil {
		input.NearPointTesting.NearVergenceTesting = &NearVergenceInput{}
	}
	if input.NearPointTesting.Accommodation == nil {
		input.NearPointTesting.Accommodation = &AccommodationInput{}
	}

	dp := input.NearPointTesting.DistPhoriaTesting
	distPhoria := preliminary.DistPhoriaTest{
		Horiz:      defaults.Str(dp.Horiz),
		Vert:       defaults.Str(dp.Vert),
		HorizExo:   boolDeref(dp.HorizExo, false),
		HorizEso:   boolDeref(dp.HorizEso, false),
		HorizOrtho: boolDeref(dp.HorizOrtho, false),
		VertRh:     boolDeref(dp.VertRh, false),
		VertLn:     boolDeref(dp.VertLn, false),
		VertOrtho:  boolDeref(dp.VertOrtho, false),
	}
	if err := s.db.Create(&distPhoria).Error; err != nil {
		return nil, err
	}

	np := input.NearPointTesting.NearPhoriaTesting
	nearPhoria := preliminary.NearPhoriaTest{
		Horiz:            defaults.Str(np.Horiz),
		Vert:             defaults.Str(np.Vert),
		GradientRatio1:   defaults.Str(np.GradientRatio1),
		CalculatedRatio1: defaults.Str(np.CalculatedRatio1),
		GradientRatio2:   defaults.Str(np.GradientRatio2),
		CalculatedRatio2: defaults.Str(np.CalculatedRatio2),
		HorizExo:         boolDeref(np.HorizExo, false),
		HorizEso:         boolDeref(np.HorizEso, false),
		HorizOrtho:       boolDeref(np.HorizOrtho, false),
		VertRh:           boolDeref(np.VertRh, false),
		VertLn:           boolDeref(np.VertLn, false),
		VertOrtho:        boolDeref(np.VertOrtho, false),
	}
	if err := s.db.Create(&nearPhoria).Error; err != nil {
		return nil, err
	}

	dv := input.NearPointTesting.DistVergenceTesting
	distVergence := preliminary.DistVergenceTest{
		Bi1: defaults.Str(dv.Bi1), Bo1: defaults.Str(dv.Bo1),
		Bi2: defaults.Str(dv.Bi2), Bo2: defaults.Str(dv.Bo2),
		Bi3: defaults.Str(dv.Bi3), Bo3: defaults.Str(dv.Bo3),
	}
	if err := s.db.Create(&distVergence).Error; err != nil {
		return nil, err
	}

	nv := input.NearPointTesting.NearVergenceTesting
	nearVergence := preliminary.NearVergenceTest{
		Bi1: defaults.Str(nv.Bi1), Bo1: defaults.Str(nv.Bo1),
		Bi2: defaults.Str(nv.Bi2), Bo2: defaults.Str(nv.Bo2),
		Bi3: defaults.Str(nv.Bi3), Bo3: defaults.Str(nv.Bo3),
	}
	if err := s.db.Create(&nearVergence).Error; err != nil {
		return nil, err
	}

	ac := input.NearPointTesting.Accommodation
	accommodation := preliminary.Accommodation{
		Pra1:                 defaults.Str(ac.Pra1),
		Nra1:                 defaults.Str(ac.Nra1),
		Pra2:                 defaults.Str(ac.Pra2),
		Nra2:                 defaults.Str(ac.Nra2),
		MemOd:                defaults.Str(ac.MemOd),
		MemOs:                defaults.Str(ac.MemOs),
		Baf:                  defaults.Str(ac.Baf),
		VergenceFacilityCpm:  defaults.Str(ac.VergenceFacilityCpm),
		VergenceFacilityWith: defaults.Str(ac.VergenceFacilityWith),
		PushUpOd:             defaults.Str(ac.PushUpOd),
		PushUpOs:             defaults.Str(ac.PushUpOs),
		PushUpOu:             defaults.Str(ac.PushUpOu),
		SlowWith:             defaults.Bool(ac.SlowWith),
	}
	if err := s.db.Create(&accommodation).Error; err != nil {
		return nil, err
	}

	distPhoriaID := i64(distPhoria.IDDistPhoriaTest)
	nearPhoriaID := i64(nearPhoria.IDNearPhoriaTest)
	distVergenceID := i64(distVergence.IDDistVergenceTest)
	nearVergenceID := i64(nearVergence.IDNearVergenceTest)
	accommodationID := i64(accommodation.IDAccommodation)

	npt := preliminary.NearPointTesting{
		DistPhoriaTestingID:   &distPhoriaID,
		NearPhoriaTestingID:   &nearPhoriaID,
		DistVergenceTestingID: &distVergenceID,
		NearVergenceTestingID: &nearVergenceID,
		AccommodationID:       &accommodationID,
	}
	if err := s.db.Create(&npt).Error; err != nil {
		return nil, err
	}
	nptID := npt.IDNearPointTesting

	// Simple sub-records
	if input.UnaidedVADistance == nil {
		input.UnaidedVADistance = &UnaidedVADistanceInput{}
	}
	d1 := input.UnaidedVADistance
	unaidedVADist := preliminary.UnaidedVADistance{Od20: defaults.Str(d1.Od20), Os20: defaults.Str(d1.Os20), Ou20: defaults.Str(d1.Ou20)}
	if err := s.db.Create(&unaidedVADist).Error; err != nil {
		return nil, err
	}
	unaidedVADistID := unaidedVADist.IDUnaidedVADistance

	if input.UnaidedPHDistance == nil {
		input.UnaidedPHDistance = &UnaidedPHDistanceInput{}
	}
	d2 := input.UnaidedPHDistance
	unaidedPHDist := preliminary.UnaidedPHDistance{Od20: defaults.Str(d2.Od20), Os20: defaults.Str(d2.Os20)}
	if err := s.db.Create(&unaidedPHDist).Error; err != nil {
		return nil, err
	}
	unaidedPHDistID := unaidedPHDist.IDUnaidedPHDistance

	if input.UnaidedVANear == nil {
		input.UnaidedVANear = &UnaidedVANearInput{}
	}
	d3 := input.UnaidedVANear
	unaidedVANear := preliminary.UnaidedVANear{Od20: defaults.Str(d3.Od20), Os20: defaults.Str(d3.Os20), Ou20: defaults.Str(d3.Ou20)}
	if err := s.db.Create(&unaidedVANear).Error; err != nil {
		return nil, err
	}
	unaidedVANearID := unaidedVANear.IDUnaidedVANear

	if input.AidedVADistance == nil {
		input.AidedVADistance = &AidedVADistanceInput{}
	}
	d4 := input.AidedVADistance
	aidedVADist := preliminary.AidedVADistance{Od20: defaults.Str(d4.Od20), Os20: defaults.Str(d4.Os20), Ou20: defaults.Str(d4.Ou20)}
	if err := s.db.Create(&aidedVADist).Error; err != nil {
		return nil, err
	}
	aidedVADistID := aidedVADist.IDAidedVADistance

	if input.AidedPHDistance == nil {
		input.AidedPHDistance = &AidedPHDistanceInput{}
	}
	d5 := input.AidedPHDistance
	aidedPHDist := preliminary.AidedPHDistance{Od20: defaults.Str(d5.Od20), Os20: defaults.Str(d5.Os20)}
	if err := s.db.Create(&aidedPHDist).Error; err != nil {
		return nil, err
	}
	aidedPHDistID := aidedPHDist.IDAidedPHDistance

	if input.AidedVANear == nil {
		input.AidedVANear = &AidedVANearInput{}
	}
	d6 := input.AidedVANear
	aidedVANear := preliminary.AidedVANear{Od20: defaults.Str(d6.Od20), Os20: defaults.Str(d6.Os20), Ou20: defaults.Str(d6.Ou20)}
	if err := s.db.Create(&aidedVANear).Error; err != nil {
		return nil, err
	}
	aidedVANearID := aidedVANear.IDAidedVANear

	if input.Confrontation == nil {
		input.Confrontation = &ConfrontationInput{}
	}
	d7 := input.Confrontation
	confrontation := preliminary.Confrontation{Od: defaults.Str(d7.Od), Os: defaults.Str(d7.Os)}
	if err := s.db.Create(&confrontation).Error; err != nil {
		return nil, err
	}
	confrontationID := confrontation.IDConfrontation

	if input.Automated == nil {
		input.Automated = &AutomatedInput{}
	}
	d8 := input.Automated
	automated := preliminary.Automated{Od: defaults.Str(d8.Od), Os: defaults.Str(d8.Os)}
	if err := s.db.Create(&automated).Error; err != nil {
		return nil, err
	}
	automatedID := automated.IDAutomated

	if input.Motility == nil {
		input.Motility = &MotilityInput{}
	}
	d9 := input.Motility
	motility := preliminary.Motility{Od: defaults.Str(d9.Od), Os: defaults.Str(d9.Os)}
	if err := s.db.Create(&motility).Error; err != nil {
		return nil, err
	}
	motilityID := motility.IDMotility

	if input.Pupils == nil {
		input.Pupils = &PupilsInput{}
	}
	d10 := input.Pupils
	pupils := preliminary.Pupils{
		OdMmDim:    defaults.Str(d10.OdMmDim),
		OdMmBright: defaults.Str(d10.OdMmBright),
		OsMmDim:    defaults.Str(d10.OsMmDim),
		OsMmBright: defaults.Str(d10.OsMmBright),
		Perrla:     boolDeref(d10.Perrla, false),
		PerrlaText: defaults.Str(d10.PerrlaText),
		Apd:        boolDeref(d10.Apd, false),
		ApdText:    defaults.Str(d10.ApdText),
	}
	if err := s.db.Create(&pupils).Error; err != nil {
		return nil, err
	}
	pupilsID := pupils.IDPupils

	if input.ColorVision == nil {
		input.ColorVision = &ColorVisionInput{}
	}
	d11 := input.ColorVision
	colorVision := preliminary.ColorVision{Od1: defaults.Str(d11.Od1), Od2: defaults.Str(d11.Od2), Os1: defaults.Str(d11.Os1), Os2: defaults.Str(d11.Os2)}
	if err := s.db.Create(&colorVision).Error; err != nil {
		return nil, err
	}
	colorVisionID := colorVision.IDColorVision

	if input.Bruckner == nil {
		input.Bruckner = &BrucknerInput{}
	}
	d12 := input.Bruckner
	bruckner := preliminary.Bruckner{Od: defaults.Str(d12.Od), Os: defaults.Str(d12.Os), GoodReflex: defaults.Bool(d12.GoodReflex)}
	if err := s.db.Create(&bruckner).Error; err != nil {
		return nil, err
	}
	brucknerID := bruckner.IDBruckner

	if input.AmslerGrid == nil {
		input.AmslerGrid = &AmslerGridInput{}
	}
	d13 := input.AmslerGrid
	amslerGrid := preliminary.AmslerGrid{Od: defaults.Str(d13.Od), Os: defaults.Str(d13.Os)}
	if err := s.db.Create(&amslerGrid).Error; err != nil {
		return nil, err
	}
	amslerGridID := amslerGrid.IDAmslerGrid

	if input.DistanceVonGraefePhoria == nil {
		input.DistanceVonGraefePhoria = &DistanceVonGraefeInput{}
	}
	d14 := input.DistanceVonGraefePhoria
	distVGP := preliminary.DistanceVonGraefePhoria{HDistVgp: defaults.Str(d14.HDistVgp), VDistVgp: defaults.Str(d14.VDistVgp)}
	if err := s.db.Create(&distVGP).Error; err != nil {
		return nil, err
	}
	distVGPID := distVGP.IDDistanceVonGraefePhoria

	if input.NearVonGraefePhoria == nil {
		input.NearVonGraefePhoria = &NearVonGraefeInput{}
	}
	d15 := input.NearVonGraefePhoria
	nearVGP := preliminary.NearVonGraefePhoria{HNearVgp: defaults.Str(d15.HNearVgp), VNearVgp: defaults.Str(d15.VNearVgp)}
	if err := s.db.Create(&nearVGP).Error; err != nil {
		return nil, err
	}
	nearVGPID := nearVGP.IDNearVonGraefePhoria

	if input.AutorefractorPreliminary == nil {
		input.AutorefractorPreliminary = &AutorefractorInput{}
	}
	d16 := input.AutorefractorPreliminary
	autoRef := preliminary.AutorefractorPreliminary{
		OdSph: defaults.Str(d16.OdSph), OsSph: defaults.Str(d16.OsSph),
		OdCyl: defaults.Str(d16.OdCyl), OsCyl: defaults.Str(d16.OsCyl),
		OdAxis: defaults.Str(d16.OdAxis), OsAxis: defaults.Str(d16.OsAxis),
		Pd: defaults.Str(d16.Pd),
	}
	if err := s.db.Create(&autoRef).Error; err != nil {
		return nil, err
	}
	autoRefID := autoRef.IDAutorefractorPreliminary

	if input.AutoKeratometerPreliminary == nil {
		input.AutoKeratometerPreliminary = &AutoKeratometerInput{}
	}
	d17 := input.AutoKeratometerPreliminary
	autoKera := preliminary.AutoKeratometerPreliminary{
		OdPw1: defaults.Str(d17.OdPw1), OsPw1: defaults.Str(d17.OsPw1),
		OdPw2: defaults.Str(d17.OdPw2), OsPw2: defaults.Str(d17.OsPw2),
		OdAxis: defaults.Str(d17.OdAxis), OsAxis: defaults.Str(d17.OsAxis),
	}
	if err := s.db.Create(&autoKera).Error; err != nil {
		return nil, err
	}
	autoKeraID := autoKera.IDAutoKeratometerPreliminary

	if input.BloodPressure == nil {
		input.BloodPressure = &BloodPressureInput{}
	}
	d18 := input.BloodPressure
	bp := preliminary.BloodPressure{Sbp: defaults.Str(d18.Sbp), Dbp: defaults.Str(d18.Dbp)}
	if err := s.db.Create(&bp).Error; err != nil {
		return nil, err
	}
	bpIntID := int(bp.IDBloodPressure)

	prelim := preliminary.PreliminaryEyeExam{
		EyeExamID:                    examID,
		UnaidedVADistanceID:          &unaidedVADistID,
		UnaidedPHDistanceID:          &unaidedPHDistID,
		UnaidedVANearID:              &unaidedVANearID,
		AidedVADistanceID:            &aidedVADistID,
		AidedPHDistanceID:            &aidedPHDistID,
		AidedVANearID:                &aidedVANearID,
		AidedByGlasses:               boolDeref(input.AidedByGlasses, false),
		AidedByContacts:              boolDeref(input.AidedByContacts, false),
		ConfrontationID:              &confrontationID,
		AutomatedID:                  &automatedID,
		MotilityID:                   &motilityID,
		PupilsID:                     &pupilsID,
		ColorVisionID:                &colorVisionID,
		DistanceCoverTest:            defaults.Str(input.DistanceCoverTest),
		NearCoverTest:                defaults.Str(input.NearCoverTest),
		NpcTest:                      defaults.Str(input.NpcTest),
		BrucknerID:                   &brucknerID,
		AmslerGridID:                 &amslerGridID,
		Worth4Dot:                    defaults.Str(input.Worth4Dot),
		StereoVision:                 defaults.Str(input.StereoVision),
		FixationDisparity:            defaults.Str(input.FixationDisparity),
		DistanceVonGraefePhorialID:   &distVGPID,
		NearVonGraefePhorialID:       &nearVGPID,
		NearPointTestingID:           &nptID,
		AutorefractorPreliminaryID:   &autoRefID,
		AutoKeratometerPreliminaryID: &autoKeraID,
		BloodPressureID:              &bpIntID,
		IrisColor:                    irisColorOrDefault(input.IrisColor),
		Note:                         defaults.Str(input.Note),
	}
	if err := s.db.Create(&prelim).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "exam_preliminary", "save", activitylog.WithEntity(examID))

	return map[string]interface{}{
		"message":                "preliminary saved successfully",
		"id_preliminary_eye_exam": prelim.IDPreliminaryEyeExam,
		"data":                   prelim.ToMap(),
	}, nil
}

// ─── GetPreliminary ───────────────────────────────────────────────────────────

func (s *Service) GetPreliminary(examID int64) (map[string]interface{}, error) {
	if _, err := s.getExam(examID); err != nil {
		return nil, err
	}

	var prelim preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error; err != nil {
		return map[string]interface{}{
			"exam_id": examID,
			"exists":  false,
			"data":    nil,
		}, nil
	}

	result := map[string]interface{}{
		"exam_id":             examID,
		"exists":              true,
		"preliminary":         prelim.ToMap(),
		"iris_color":          prelim.IrisColor,
		"distance_cover_test": prelim.DistanceCoverTest,
		"near_cover_test":     prelim.NearCoverTest,
		"npc_test":            prelim.NpcTest,
		"worth_4_dot":         prelim.Worth4Dot,
		"stereo_vision":       prelim.StereoVision,
		"fixation_disparity":  prelim.FixationDisparity,
		"note":                prelim.Note,
	}

	loadSub := func(id *int64, dest interface{}) bool {
		if id == nil {
			return false
		}
		return s.db.First(dest, *id).Error == nil
	}

	var uvaDist preliminary.UnaidedVADistance
	if loadSub(prelim.UnaidedVADistanceID, &uvaDist) {
		result["unaided_va_distance"] = uvaDist.ToMap()
	} else {
		result["unaided_va_distance"] = nil
	}

	var uphDist preliminary.UnaidedPHDistance
	if loadSub(prelim.UnaidedPHDistanceID, &uphDist) {
		result["unaided_ph_distance"] = uphDist.ToMap()
	} else {
		result["unaided_ph_distance"] = nil
	}

	var uvaNear preliminary.UnaidedVANear
	if loadSub(prelim.UnaidedVANearID, &uvaNear) {
		result["unaided_va_near"] = uvaNear.ToMap()
	} else {
		result["unaided_va_near"] = nil
	}

	var avaDist preliminary.AidedVADistance
	if loadSub(prelim.AidedVADistanceID, &avaDist) {
		result["aided_va_distance"] = avaDist.ToMap()
	} else {
		result["aided_va_distance"] = nil
	}

	var aphDist preliminary.AidedPHDistance
	if loadSub(prelim.AidedPHDistanceID, &aphDist) {
		result["aided_ph_distance"] = aphDist.ToMap()
	} else {
		result["aided_ph_distance"] = nil
	}

	var avaNear preliminary.AidedVANear
	if loadSub(prelim.AidedVANearID, &avaNear) {
		result["aided_va_near"] = avaNear.ToMap()
	} else {
		result["aided_va_near"] = nil
	}

	var conf preliminary.Confrontation
	if loadSub(prelim.ConfrontationID, &conf) {
		result["confrontation"] = conf.ToMap()
	} else {
		result["confrontation"] = nil
	}

	var auto preliminary.Automated
	if loadSub(prelim.AutomatedID, &auto) {
		result["automated"] = auto.ToMap()
	} else {
		result["automated"] = nil
	}

	var mot preliminary.Motility
	if loadSub(prelim.MotilityID, &mot) {
		result["motility"] = mot.ToMap()
	} else {
		result["motility"] = nil
	}

	var pup preliminary.Pupils
	if loadSub(prelim.PupilsID, &pup) {
		result["pupils"] = pup.ToMap()
	} else {
		result["pupils"] = nil
	}

	var cv preliminary.ColorVision
	if loadSub(prelim.ColorVisionID, &cv) {
		result["color_vision"] = cv.ToMap()
	} else {
		result["color_vision"] = nil
	}

	var br preliminary.Bruckner
	if loadSub(prelim.BrucknerID, &br) {
		result["bruckner"] = br.ToMap()
	} else {
		result["bruckner"] = nil
	}

	var ag preliminary.AmslerGrid
	if loadSub(prelim.AmslerGridID, &ag) {
		result["amsler_grid"] = ag.ToMap()
	} else {
		result["amsler_grid"] = nil
	}

	var dvgp preliminary.DistanceVonGraefePhoria
	if loadSub(prelim.DistanceVonGraefePhorialID, &dvgp) {
		result["distance_von_graefe_phoria"] = dvgp.ToMap()
	} else {
		result["distance_von_graefe_phoria"] = nil
	}

	var nvgp preliminary.NearVonGraefePhoria
	if loadSub(prelim.NearVonGraefePhorialID, &nvgp) {
		result["near_von_graefe_phoria"] = nvgp.ToMap()
	} else {
		result["near_von_graefe_phoria"] = nil
	}

	var ar preliminary.AutorefractorPreliminary
	if loadSub(prelim.AutorefractorPreliminaryID, &ar) {
		result["autorefractor_preliminary"] = ar.ToMap()
	} else {
		result["autorefractor_preliminary"] = nil
	}

	var ak preliminary.AutoKeratometerPreliminary
	if loadSub(prelim.AutoKeratometerPreliminaryID, &ak) {
		result["auto_keratometer_preliminary"] = ak.ToMap()
	} else {
		result["auto_keratometer_preliminary"] = nil
	}

	// Blood pressure — FK is *int, convert for query
	if prelim.BloodPressureID != nil {
		bpID := int64(*prelim.BloodPressureID)
		var bp preliminary.BloodPressure
		if s.db.First(&bp, bpID).Error == nil {
			result["blood_pressure"] = bp.ToMap()
		} else {
			result["blood_pressure"] = nil
		}
	} else {
		result["blood_pressure"] = nil
	}

	// Entrance rx
	var eg preliminary.EntranceGlasses
	if loadSub(prelim.EntranceGlassesID, &eg) {
		result["entrance_glasses"] = eg.ToMap()
	} else {
		result["entrance_glasses"] = nil
	}

	var ecl preliminary.EntranceContLens
	if loadSub(prelim.EntranceContLensID, &ecl) {
		result["entrance_cont_lens"] = ecl.ToMap()
	} else {
		result["entrance_cont_lens"] = nil
	}

	// NearPointTesting with sub-models
	if prelim.NearPointTestingID != nil {
		var npt preliminary.NearPointTesting
		if s.db.First(&npt, *prelim.NearPointTestingID).Error == nil {
			nptMap := npt.ToMap()

			var dp preliminary.DistPhoriaTest
			if npt.DistPhoriaTestingID != nil && s.db.First(&dp, *npt.DistPhoriaTestingID).Error == nil {
				nptMap["dist_phoria_testing"] = dp.ToMap()
			} else {
				nptMap["dist_phoria_testing"] = nil
			}

			var np preliminary.NearPhoriaTest
			if npt.NearPhoriaTestingID != nil && s.db.First(&np, *npt.NearPhoriaTestingID).Error == nil {
				nptMap["near_phoria_testing"] = np.ToMap()
			} else {
				nptMap["near_phoria_testing"] = nil
			}

			var dv preliminary.DistVergenceTest
			if npt.DistVergenceTestingID != nil && s.db.First(&dv, *npt.DistVergenceTestingID).Error == nil {
				nptMap["dist_vergence_testing"] = dv.ToMap()
			} else {
				nptMap["dist_vergence_testing"] = nil
			}

			var nv preliminary.NearVergenceTest
			if npt.NearVergenceTestingID != nil && s.db.First(&nv, *npt.NearVergenceTestingID).Error == nil {
				nptMap["near_vergence_testing"] = nv.ToMap()
			} else {
				nptMap["near_vergence_testing"] = nil
			}

			var ac preliminary.Accommodation
			if npt.AccommodationID != nil && s.db.First(&ac, *npt.AccommodationID).Error == nil {
				nptMap["accommodation"] = ac.ToMap()
			} else {
				nptMap["accommodation"] = nil
			}

			result["near_point_testing"] = nptMap
		} else {
			result["near_point_testing"] = nil
		}
	} else {
		result["near_point_testing"] = nil
	}

	return result, nil
}

// ─── UpdatePreliminary ────────────────────────────────────────────────────────

func (s *Service) UpdatePreliminary(username string, examID int64, input UpdatePreliminaryInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}

	if exam.EmployeeID != i64(emp.IDEmployee) {
		return nil, errors.New("you are not authorized to update this exam")
	}
	if exam.Passed {
		return nil, errors.New("cannot update preliminary for a completed exam")
	}

	var prelim preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error; err != nil {
		return nil, errors.New("preliminary record not found for this exam")
	}

	// Update simple fields
	if input.AidedByGlasses != nil {
		prelim.AidedByGlasses = *input.AidedByGlasses
	}
	if input.AidedByContacts != nil {
		prelim.AidedByContacts = *input.AidedByContacts
	}
	if input.DistanceCoverTest != nil {
		prelim.DistanceCoverTest = input.DistanceCoverTest
	}
	if input.NearCoverTest != nil {
		prelim.NearCoverTest = input.NearCoverTest
	}
	if input.NpcTest != nil {
		prelim.NpcTest = input.NpcTest
	}
	if input.Worth4Dot != nil {
		prelim.Worth4Dot = input.Worth4Dot
	}
	if input.StereoVision != nil {
		prelim.StereoVision = input.StereoVision
	}
	if input.FixationDisparity != nil {
		prelim.FixationDisparity = input.FixationDisparity
	}
	if input.IrisColor != nil && *input.IrisColor != "" {
		prelim.IrisColor = *input.IrisColor
	}
	if input.Note != nil {
		prelim.Note = input.Note
	}

	// Unaided VA Distance
	if input.UnaidedVADistance != nil {
		d := input.UnaidedVADistance
		var r preliminary.UnaidedVADistance
		if prelim.UnaidedVADistanceID != nil {
			s.db.First(&r, *prelim.UnaidedVADistanceID)
		}
		if d.Od20 != nil { r.Od20 = d.Od20 }
		if d.Os20 != nil { r.Os20 = d.Os20 }
		if d.Ou20 != nil { r.Ou20 = d.Ou20 }
		s.db.Save(&r)
		id := r.IDUnaidedVADistance
		prelim.UnaidedVADistanceID = &id
	}

	// Unaided PH Distance
	if input.UnaidedPHDistance != nil {
		d := input.UnaidedPHDistance
		var r preliminary.UnaidedPHDistance
		if prelim.UnaidedPHDistanceID != nil {
			s.db.First(&r, *prelim.UnaidedPHDistanceID)
		}
		if d.Od20 != nil { r.Od20 = d.Od20 }
		if d.Os20 != nil { r.Os20 = d.Os20 }
		s.db.Save(&r)
		id := r.IDUnaidedPHDistance
		prelim.UnaidedPHDistanceID = &id
	}

	// Unaided VA Near
	if input.UnaidedVANear != nil {
		d := input.UnaidedVANear
		var r preliminary.UnaidedVANear
		if prelim.UnaidedVANearID != nil {
			s.db.First(&r, *prelim.UnaidedVANearID)
		}
		if d.Od20 != nil { r.Od20 = d.Od20 }
		if d.Os20 != nil { r.Os20 = d.Os20 }
		if d.Ou20 != nil { r.Ou20 = d.Ou20 }
		s.db.Save(&r)
		id := r.IDUnaidedVANear
		prelim.UnaidedVANearID = &id
	}

	// Aided VA Distance
	if input.AidedVADistance != nil {
		d := input.AidedVADistance
		var r preliminary.AidedVADistance
		if prelim.AidedVADistanceID != nil {
			s.db.First(&r, *prelim.AidedVADistanceID)
		}
		if d.Od20 != nil { r.Od20 = d.Od20 }
		if d.Os20 != nil { r.Os20 = d.Os20 }
		if d.Ou20 != nil { r.Ou20 = d.Ou20 }
		s.db.Save(&r)
		id := r.IDAidedVADistance
		prelim.AidedVADistanceID = &id
	}

	// Aided PH Distance
	if input.AidedPHDistance != nil {
		d := input.AidedPHDistance
		var r preliminary.AidedPHDistance
		if prelim.AidedPHDistanceID != nil {
			s.db.First(&r, *prelim.AidedPHDistanceID)
		}
		if d.Od20 != nil { r.Od20 = d.Od20 }
		if d.Os20 != nil { r.Os20 = d.Os20 }
		s.db.Save(&r)
		id := r.IDAidedPHDistance
		prelim.AidedPHDistanceID = &id
	}

	// Aided VA Near
	if input.AidedVANear != nil {
		d := input.AidedVANear
		var r preliminary.AidedVANear
		if prelim.AidedVANearID != nil {
			s.db.First(&r, *prelim.AidedVANearID)
		}
		if d.Od20 != nil { r.Od20 = d.Od20 }
		if d.Os20 != nil { r.Os20 = d.Os20 }
		if d.Ou20 != nil { r.Ou20 = d.Ou20 }
		s.db.Save(&r)
		id := r.IDAidedVANear
		prelim.AidedVANearID = &id
	}

	// Confrontation
	if input.Confrontation != nil {
		d := input.Confrontation
		var r preliminary.Confrontation
		if prelim.ConfrontationID != nil {
			s.db.First(&r, *prelim.ConfrontationID)
		}
		if d.Od != nil { r.Od = d.Od }
		if d.Os != nil { r.Os = d.Os }
		s.db.Save(&r)
		id := r.IDConfrontation
		prelim.ConfrontationID = &id
	}

	// Automated
	if input.Automated != nil {
		d := input.Automated
		var r preliminary.Automated
		if prelim.AutomatedID != nil {
			s.db.First(&r, *prelim.AutomatedID)
		}
		if d.Od != nil { r.Od = d.Od }
		if d.Os != nil { r.Os = d.Os }
		s.db.Save(&r)
		id := r.IDAutomated
		prelim.AutomatedID = &id
	}

	// Motility
	if input.Motility != nil {
		d := input.Motility
		var r preliminary.Motility
		if prelim.MotilityID != nil {
			s.db.First(&r, *prelim.MotilityID)
		}
		if d.Od != nil { r.Od = d.Od }
		if d.Os != nil { r.Os = d.Os }
		s.db.Save(&r)
		id := r.IDMotility
		prelim.MotilityID = &id
	}

	// Pupils
	if input.Pupils != nil {
		d := input.Pupils
		var r preliminary.Pupils
		if prelim.PupilsID != nil {
			s.db.First(&r, *prelim.PupilsID)
		}
		if d.OdMmDim != nil { r.OdMmDim = d.OdMmDim }
		if d.OdMmBright != nil { r.OdMmBright = d.OdMmBright }
		if d.OsMmDim != nil { r.OsMmDim = d.OsMmDim }
		if d.OsMmBright != nil { r.OsMmBright = d.OsMmBright }
		if d.Perrla != nil { r.Perrla = *d.Perrla }
		if d.PerrlaText != nil { r.PerrlaText = d.PerrlaText }
		if d.Apd != nil { r.Apd = *d.Apd }
		if d.ApdText != nil { r.ApdText = d.ApdText }
		s.db.Save(&r)
		id := r.IDPupils
		prelim.PupilsID = &id
	}

	// ColorVision
	if input.ColorVision != nil {
		d := input.ColorVision
		var r preliminary.ColorVision
		if prelim.ColorVisionID != nil {
			s.db.First(&r, *prelim.ColorVisionID)
		}
		if d.Od1 != nil { r.Od1 = d.Od1 }
		if d.Od2 != nil { r.Od2 = d.Od2 }
		if d.Os1 != nil { r.Os1 = d.Os1 }
		if d.Os2 != nil { r.Os2 = d.Os2 }
		s.db.Save(&r)
		id := r.IDColorVision
		prelim.ColorVisionID = &id
	}

	// Bruckner
	if input.Bruckner != nil {
		d := input.Bruckner
		var r preliminary.Bruckner
		if prelim.BrucknerID != nil {
			s.db.First(&r, *prelim.BrucknerID)
		}
		if d.Od != nil { r.Od = d.Od }
		if d.Os != nil { r.Os = d.Os }
		if d.GoodReflex != nil { r.GoodReflex = d.GoodReflex }
		s.db.Save(&r)
		id := r.IDBruckner
		prelim.BrucknerID = &id
	}

	// AmslerGrid
	if input.AmslerGrid != nil {
		d := input.AmslerGrid
		var r preliminary.AmslerGrid
		if prelim.AmslerGridID != nil {
			s.db.First(&r, *prelim.AmslerGridID)
		}
		if d.Od != nil { r.Od = d.Od }
		if d.Os != nil { r.Os = d.Os }
		s.db.Save(&r)
		id := r.IDAmslerGrid
		prelim.AmslerGridID = &id
	}

	// Distance Von Graefe
	if input.DistanceVonGraefePhoria != nil {
		d := input.DistanceVonGraefePhoria
		var r preliminary.DistanceVonGraefePhoria
		if prelim.DistanceVonGraefePhorialID != nil {
			s.db.First(&r, *prelim.DistanceVonGraefePhorialID)
		}
		if d.HDistVgp != nil { r.HDistVgp = d.HDistVgp }
		if d.VDistVgp != nil { r.VDistVgp = d.VDistVgp }
		s.db.Save(&r)
		id := r.IDDistanceVonGraefePhoria
		prelim.DistanceVonGraefePhorialID = &id
	}

	// Near Von Graefe
	if input.NearVonGraefePhoria != nil {
		d := input.NearVonGraefePhoria
		var r preliminary.NearVonGraefePhoria
		if prelim.NearVonGraefePhorialID != nil {
			s.db.First(&r, *prelim.NearVonGraefePhorialID)
		}
		if d.HNearVgp != nil { r.HNearVgp = d.HNearVgp }
		if d.VNearVgp != nil { r.VNearVgp = d.VNearVgp }
		s.db.Save(&r)
		id := r.IDNearVonGraefePhoria
		prelim.NearVonGraefePhorialID = &id
	}

	// AutorefractorPreliminary
	if input.AutorefractorPreliminary != nil {
		d := input.AutorefractorPreliminary
		var r preliminary.AutorefractorPreliminary
		if prelim.AutorefractorPreliminaryID != nil {
			s.db.First(&r, *prelim.AutorefractorPreliminaryID)
		}
		if d.OdSph != nil { r.OdSph = d.OdSph }
		if d.OsSph != nil { r.OsSph = d.OsSph }
		if d.OdCyl != nil { r.OdCyl = d.OdCyl }
		if d.OsCyl != nil { r.OsCyl = d.OsCyl }
		if d.OdAxis != nil { r.OdAxis = d.OdAxis }
		if d.OsAxis != nil { r.OsAxis = d.OsAxis }
		if d.Pd != nil { r.Pd = d.Pd }
		s.db.Save(&r)
		id := r.IDAutorefractorPreliminary
		prelim.AutorefractorPreliminaryID = &id
	}

	// AutoKeratometerPreliminary
	if input.AutoKeratometerPreliminary != nil {
		d := input.AutoKeratometerPreliminary
		var r preliminary.AutoKeratometerPreliminary
		if prelim.AutoKeratometerPreliminaryID != nil {
			s.db.First(&r, *prelim.AutoKeratometerPreliminaryID)
		}
		if d.OdPw1 != nil { r.OdPw1 = d.OdPw1 }
		if d.OsPw1 != nil { r.OsPw1 = d.OsPw1 }
		if d.OdPw2 != nil { r.OdPw2 = d.OdPw2 }
		if d.OsPw2 != nil { r.OsPw2 = d.OsPw2 }
		if d.OdAxis != nil { r.OdAxis = d.OdAxis }
		if d.OsAxis != nil { r.OsAxis = d.OsAxis }
		s.db.Save(&r)
		id := r.IDAutoKeratometerPreliminary
		prelim.AutoKeratometerPreliminaryID = &id
	}

	// BloodPressure
	if input.BloodPressure != nil {
		d := input.BloodPressure
		var r preliminary.BloodPressure
		if prelim.BloodPressureID != nil {
			s.db.First(&r, int64(*prelim.BloodPressureID))
		}
		if d.Sbp != nil { r.Sbp = d.Sbp }
		if d.Dbp != nil { r.Dbp = d.Dbp }
		s.db.Save(&r)
		bpIntID := int(r.IDBloodPressure)
		prelim.BloodPressureID = &bpIntID
	}

	// NearPointTesting
	if input.NearPointTesting != nil {
		nptInput := input.NearPointTesting
		var npt preliminary.NearPointTesting
		if prelim.NearPointTestingID != nil {
			s.db.First(&npt, *prelim.NearPointTestingID)
		}

		if nptInput.DistPhoriaTesting != nil {
			d := nptInput.DistPhoriaTesting
			var r preliminary.DistPhoriaTest
			if npt.DistPhoriaTestingID != nil {
				s.db.First(&r, *npt.DistPhoriaTestingID)
			}
			if d.Horiz != nil { r.Horiz = d.Horiz }
			if d.Vert != nil { r.Vert = d.Vert }
			if d.HorizExo != nil { r.HorizExo = *d.HorizExo }
			if d.HorizEso != nil { r.HorizEso = *d.HorizEso }
			if d.HorizOrtho != nil { r.HorizOrtho = *d.HorizOrtho }
			if d.VertRh != nil { r.VertRh = *d.VertRh }
			if d.VertLn != nil { r.VertLn = *d.VertLn }
			if d.VertOrtho != nil { r.VertOrtho = *d.VertOrtho }
			s.db.Save(&r)
			dpID := i64(r.IDDistPhoriaTest)
			npt.DistPhoriaTestingID = &dpID
		}

		if nptInput.NearPhoriaTesting != nil {
			d := nptInput.NearPhoriaTesting
			var r preliminary.NearPhoriaTest
			if npt.NearPhoriaTestingID != nil {
				s.db.First(&r, *npt.NearPhoriaTestingID)
			}
			if d.Horiz != nil { r.Horiz = d.Horiz }
			if d.Vert != nil { r.Vert = d.Vert }
			if d.GradientRatio1 != nil { r.GradientRatio1 = d.GradientRatio1 }
			if d.CalculatedRatio1 != nil { r.CalculatedRatio1 = d.CalculatedRatio1 }
			if d.GradientRatio2 != nil { r.GradientRatio2 = d.GradientRatio2 }
			if d.CalculatedRatio2 != nil { r.CalculatedRatio2 = d.CalculatedRatio2 }
			if d.HorizExo != nil { r.HorizExo = *d.HorizExo }
			if d.HorizEso != nil { r.HorizEso = *d.HorizEso }
			if d.HorizOrtho != nil { r.HorizOrtho = *d.HorizOrtho }
			if d.VertRh != nil { r.VertRh = *d.VertRh }
			if d.VertLn != nil { r.VertLn = *d.VertLn }
			if d.VertOrtho != nil { r.VertOrtho = *d.VertOrtho }
			s.db.Save(&r)
			npID := i64(r.IDNearPhoriaTest)
			npt.NearPhoriaTestingID = &npID
		}

		if nptInput.DistVergenceTesting != nil {
			d := nptInput.DistVergenceTesting
			var r preliminary.DistVergenceTest
			if npt.DistVergenceTestingID != nil {
				s.db.First(&r, *npt.DistVergenceTestingID)
			}
			if d.Bi1 != nil { r.Bi1 = d.Bi1 }
			if d.Bo1 != nil { r.Bo1 = d.Bo1 }
			if d.Bi2 != nil { r.Bi2 = d.Bi2 }
			if d.Bo2 != nil { r.Bo2 = d.Bo2 }
			if d.Bi3 != nil { r.Bi3 = d.Bi3 }
			if d.Bo3 != nil { r.Bo3 = d.Bo3 }
			s.db.Save(&r)
			dvID := i64(r.IDDistVergenceTest)
			npt.DistVergenceTestingID = &dvID
		}

		if nptInput.NearVergenceTesting != nil {
			d := nptInput.NearVergenceTesting
			var r preliminary.NearVergenceTest
			if npt.NearVergenceTestingID != nil {
				s.db.First(&r, *npt.NearVergenceTestingID)
			}
			if d.Bi1 != nil { r.Bi1 = d.Bi1 }
			if d.Bo1 != nil { r.Bo1 = d.Bo1 }
			if d.Bi2 != nil { r.Bi2 = d.Bi2 }
			if d.Bo2 != nil { r.Bo2 = d.Bo2 }
			if d.Bi3 != nil { r.Bi3 = d.Bi3 }
			if d.Bo3 != nil { r.Bo3 = d.Bo3 }
			s.db.Save(&r)
			nvID := i64(r.IDNearVergenceTest)
			npt.NearVergenceTestingID = &nvID
		}

		if nptInput.Accommodation != nil {
			d := nptInput.Accommodation
			var r preliminary.Accommodation
			if npt.AccommodationID != nil {
				s.db.First(&r, *npt.AccommodationID)
			}
			if d.Pra1 != nil { r.Pra1 = d.Pra1 }
			if d.Nra1 != nil { r.Nra1 = d.Nra1 }
			if d.Pra2 != nil { r.Pra2 = d.Pra2 }
			if d.Nra2 != nil { r.Nra2 = d.Nra2 }
			if d.MemOd != nil { r.MemOd = d.MemOd }
			if d.MemOs != nil { r.MemOs = d.MemOs }
			if d.Baf != nil { r.Baf = d.Baf }
			if d.VergenceFacilityCpm != nil { r.VergenceFacilityCpm = d.VergenceFacilityCpm }
			if d.VergenceFacilityWith != nil { r.VergenceFacilityWith = d.VergenceFacilityWith }
			if d.PushUpOd != nil { r.PushUpOd = d.PushUpOd }
			if d.PushUpOs != nil { r.PushUpOs = d.PushUpOs }
			if d.PushUpOu != nil { r.PushUpOu = d.PushUpOu }
			if d.SlowWith != nil { r.SlowWith = d.SlowWith }
			s.db.Save(&r)
			acID := i64(r.IDAccommodation)
			npt.AccommodationID = &acID
		}

		s.db.Save(&npt)
		nptID := npt.IDNearPointTesting
		prelim.NearPointTestingID = &nptID
	}

	if err := s.db.Save(&prelim).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "exam_preliminary", "update", activitylog.WithEntity(examID))

	return map[string]interface{}{
		"message": "preliminary updated successfully",
		"data":    prelim.ToMap(),
	}, nil
}

// ─── GetPrescriptionList ──────────────────────────────────────────────────────

func (s *Service) GetPrescriptionList(patientID int64) ([]map[string]interface{}, error) {
	var count int64
	s.db.Table("patient").Where("id_patient = ?", patientID).Count(&count)
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	var prescriptions []rxPrescription
	s.db.Where("patient_id = ?", patientID).Find(&prescriptions)

	var result []map[string]interface{}
	for _, rx := range prescriptions {
		var dateStr *string
		if rx.PrescriptionDate != nil {
			s := rx.PrescriptionDate.Format("2006-01-02")
			dateStr = &s
		}

		var glasses rxGlasses
		if err := s.db.Where("prescription_id = ?", rx.IDPatientPrescription).First(&glasses).Error; err == nil {
			entry := map[string]interface{}{
				"id_rx":  glasses.IDGlassesPrescription,
				"date":   dateStr,
				"g_or_c": "glasses",
				"details": map[string]interface{}{
					"od_sph":     glasses.OdSph,
					"os_sph":     glasses.OsSph,
					"od_cyl":     glasses.OdCyl,
					"os_cyl":     glasses.OsCyl,
					"od_axis":    glasses.OdAxis,
					"os_axis":    glasses.OsAxis,
					"od_add":     glasses.OdAdd,
					"os_add":     glasses.OsAdd,
					"od_h_prism": glasses.OdHPrism,
					"os_h_prism": glasses.OsHPrism,
				},
			}
			result = append(result, entry)
		}

		var contact rxContactLens
		if err := s.db.Where("prescription_id = ?", rx.IDPatientPrescription).First(&contact).Error; err == nil {
			entry := map[string]interface{}{
				"id_rx":  contact.IDContactLensPrescription,
				"date":   dateStr,
				"g_or_c": "contact",
				"details": map[string]interface{}{
					"od_cont_lens": contact.OdContLens,
					"os_cont_lens": contact.OsContLens,
					"od_bc":        contact.OdBc,
					"os_bc":        contact.OsBc,
					"od_dia":       contact.OdDia,
					"os_dia":       contact.OsDia,
					"od_pwr":       contact.OdPwr,
					"os_pwr":       contact.OsPwr,
					"od_cyl":       contact.OdCyl,
					"os_cyl":       contact.OsCyl,
					"od_axis":      contact.OdAxis,
					"os_axis":      contact.OsAxis,
					"od_add":       contact.OdAdd,
					"os_add":       contact.OsAdd,
				},
			}
			result = append(result, entry)
		}
	}

	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

// ─── FillEntranceRx ───────────────────────────────────────────────────────────

func (s *Service) FillEntranceRx(examID int64, input FillEntranceRxInput) (map[string]interface{}, error) {
	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}
	if exam.Passed {
		return nil, errors.New("cannot modify entrance rx for a completed exam")
	}

	// Find or create preliminary record
	var prelim preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error; err != nil {
		// No preliminary yet — create one
		prelim = preliminary.PreliminaryEyeExam{
			EyeExamID: examID,
			IrisColor: "n/a",
		}
		if err := s.db.Create(&prelim).Error; err != nil {
			return nil, err
		}
	}

	var entranceGlassesID *int64
	var entranceContLensID *int64

	if input.RxData.Glasses != nil {
		var gRx rxGlasses
		if err := s.db.Where("prescription_id = ?", *input.RxData.Glasses).First(&gRx).Error; err != nil {
			return nil, errors.New("glasses prescription not found")
		}
		var presc rxPrescription
		s.db.First(&presc, gRx.PrescriptionID)
		eg := preliminary.EntranceGlasses{
			Data:     presc.PrescriptionDate,
			OdSph:    gRx.OdSph,
			OsSph:    gRx.OsSph,
			OdCyl:    gRx.OdCyl,
			OsCyl:    gRx.OsCyl,
			OdAxis:   gRx.OdAxis,
			OsAxis:   gRx.OsAxis,
			OdAdd:    gRx.OdAdd,
			OsAdd:    gRx.OsAdd,
			OdHPrism: gRx.OdHPrism,
			OsHPrism: gRx.OsHPrism,
		}
		if err := s.db.Create(&eg).Error; err != nil {
			return nil, err
		}
		entranceGlassesID = &eg.IDEntranceGlasses
		prelim.EntranceGlassesID = entranceGlassesID
	}

	if input.RxData.Contact != nil {
		var cRx rxContactLens
		if err := s.db.Where("prescription_id = ?", *input.RxData.Contact).First(&cRx).Error; err != nil {
			return nil, errors.New("contact lens prescription not found")
		}
		var presc rxPrescription
		s.db.First(&presc, cRx.PrescriptionID)
		ecl := preliminary.EntranceContLens{
			Data:    presc.PrescriptionDate,
			OdBrand: cRx.OdContLens,
			OsBrand: cRx.OsContLens,
			OdBaseC: cRx.OdBc,
			OsBaseC: cRx.OsBc,
			OdDia:   cRx.OdDia,
			OsDia:   cRx.OsDia,
			OdPwr:   cRx.OdPwr,
			OsPwr:   cRx.OsPwr,
			OdCyl:   cRx.OdCyl,
			OsCyl:   cRx.OsCyl,
			OdAxis:  cRx.OdAxis,
			OsAxis:  cRx.OsAxis,
			OdAdd:   cRx.OdAdd,
			OsAdd:   cRx.OsAdd,
		}
		if err := s.db.Create(&ecl).Error; err != nil {
			return nil, err
		}
		entranceContLensID = &ecl.IDEntranceContLens
		prelim.EntranceContLensID = entranceContLensID
	}

	// Update existing preliminary with entrance IDs
	if err := s.db.Save(&prelim).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":                 "entrance rx saved successfully",
		"id_preliminary_eye_exam": prelim.IDPreliminaryEyeExam,
		"entrance_glasses_id":     entranceGlassesID,
		"entrance_cont_lens_id":   entranceContLensID,
	}, nil
}

// ─── UpdateEntranceRx ─────────────────────────────────────────────────────────

func (s *Service) UpdateEntranceRx(examID int64, input FillEntranceRxInput) (map[string]interface{}, error) {
	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}
	if exam.Passed {
		return nil, errors.New("cannot modify entrance rx for a completed exam")
	}

	var prelim preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error; err != nil {
		return nil, errors.New("preliminary record not found for this exam")
	}

	if input.RxData.Glasses != nil {
		var gRx rxGlasses
		if err := s.db.Where("prescription_id = ?", *input.RxData.Glasses).First(&gRx).Error; err != nil {
			return nil, errors.New("glasses prescription not found")
		}
		var presc rxPrescription
		s.db.First(&presc, gRx.PrescriptionID)

		var eg preliminary.EntranceGlasses
		if prelim.EntranceGlassesID != nil {
			s.db.First(&eg, *prelim.EntranceGlassesID)
		}
		eg.Data = presc.PrescriptionDate
		eg.OdSph = gRx.OdSph
		eg.OsSph = gRx.OsSph
		eg.OdCyl = gRx.OdCyl
		eg.OsCyl = gRx.OsCyl
		eg.OdAxis = gRx.OdAxis
		eg.OsAxis = gRx.OsAxis
		eg.OdAdd = gRx.OdAdd
		eg.OsAdd = gRx.OsAdd
		eg.OdHPrism = gRx.OdHPrism
		eg.OsHPrism = gRx.OsHPrism
		if err := s.db.Save(&eg).Error; err != nil {
			return nil, err
		}
		prelim.EntranceGlassesID = &eg.IDEntranceGlasses
	}

	if input.RxData.Contact != nil {
		var cRx rxContactLens
		if err := s.db.Where("prescription_id = ?", *input.RxData.Contact).First(&cRx).Error; err != nil {
			return nil, errors.New("contact lens prescription not found")
		}
		var presc rxPrescription
		s.db.First(&presc, cRx.PrescriptionID)

		var ecl preliminary.EntranceContLens
		if prelim.EntranceContLensID != nil {
			s.db.First(&ecl, *prelim.EntranceContLensID)
		}
		ecl.Data = presc.PrescriptionDate
		ecl.OdBrand = cRx.OdContLens
		ecl.OsBrand = cRx.OsContLens
		ecl.OdBaseC = cRx.OdBc
		ecl.OsBaseC = cRx.OsBc
		ecl.OdDia = cRx.OdDia
		ecl.OsDia = cRx.OsDia
		ecl.OdPwr = cRx.OdPwr
		ecl.OsPwr = cRx.OsPwr
		ecl.OdCyl = cRx.OdCyl
		ecl.OsCyl = cRx.OsCyl
		ecl.OdAxis = cRx.OdAxis
		ecl.OsAxis = cRx.OsAxis
		ecl.OdAdd = cRx.OdAdd
		ecl.OsAdd = cRx.OsAdd
		if err := s.db.Save(&ecl).Error; err != nil {
			return nil, err
		}
		prelim.EntranceContLensID = &ecl.IDEntranceContLens
	}

	if err := s.db.Save(&prelim).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":               "entrance rx updated successfully",
		"entrance_glasses_id":   prelim.EntranceGlassesID,
		"entrance_cont_lens_id": prelim.EntranceContLensID,
	}, nil
}

// ─── GetEntranceRx ────────────────────────────────────────────────────────────

func (s *Service) GetEntranceRx(examID int64) (map[string]interface{}, error) {
	if _, err := s.getExam(examID); err != nil {
		return nil, err
	}

	var prelim preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error; err != nil {
		return map[string]interface{}{
			"exam_id": examID,
			"rx_data": nil,
		}, nil
	}

	if prelim.EntranceGlassesID == nil && prelim.EntranceContLensID == nil {
		return map[string]interface{}{
			"exam_id": examID,
			"rx_data": nil,
		}, nil
	}

	rxData := map[string]interface{}{
		"glasses": nil,
		"contact": nil,
	}

	if prelim.EntranceGlassesID != nil {
		var eg preliminary.EntranceGlasses
		if s.db.First(&eg, *prelim.EntranceGlassesID).Error == nil {
			rxData["glasses"] = eg.ToMap()
		}
	}

	if prelim.EntranceContLensID != nil {
		var ecl preliminary.EntranceContLens
		if s.db.First(&ecl, *prelim.EntranceContLensID).Error == nil {
			rxData["contact"] = ecl.ToMap()
		}
	}

	return map[string]interface{}{
		"exam_id": examID,
		"rx_data": rxData,
	}, nil
}

// ─── DeleteEntranceRx ─────────────────────────────────────────────────────────

func (s *Service) DeleteEntranceRx(examID int64, input FillEntranceRxInput) (map[string]interface{}, error) {
	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}
	if exam.Passed {
		return nil, errors.New("cannot modify entrance rx for a completed exam")
	}

	var prelim preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error; err != nil {
		return nil, errors.New("preliminary record not found for this exam")
	}

	if input.RxData.Glasses != nil && prelim.EntranceGlassesID != nil {
		s.db.Delete(&preliminary.EntranceGlasses{}, *prelim.EntranceGlassesID)
		prelim.EntranceGlassesID = nil
	}

	if input.RxData.Contact != nil && prelim.EntranceContLensID != nil {
		s.db.Delete(&preliminary.EntranceContLens{}, *prelim.EntranceContLensID)
		prelim.EntranceContLensID = nil
	}

	if prelim.EntranceGlassesID == nil && prelim.EntranceContLensID == nil {
		s.db.Delete(&prelim)
	} else {
		s.db.Save(&prelim)
	}

	return map[string]interface{}{
		"message": "entrance rx deleted successfully",
		"exam_id": examID,
	}, nil
}

// ─── CreateNearPointTesting (standalone) ──────────────────────────────────────

func (s *Service) CreateNearPointTesting(examID int64, input NearPointTestingInput) (map[string]interface{}, error) {
	if _, err := s.getExam(examID); err != nil {
		return nil, err
	}

	var distPhoria preliminary.DistPhoriaTest
	if d := input.DistPhoriaTesting; d != nil {
		distPhoria = preliminary.DistPhoriaTest{
			Horiz:      defaults.Str(d.Horiz),
			Vert:       defaults.Str(d.Vert),
			HorizExo:   boolDeref(d.HorizExo, false),
			HorizEso:   boolDeref(d.HorizEso, false),
			HorizOrtho: boolDeref(d.HorizOrtho, false),
			VertRh:     boolDeref(d.VertRh, false),
			VertLn:     boolDeref(d.VertLn, false),
			VertOrtho:  boolDeref(d.VertOrtho, false),
		}
	}
	if err := s.db.Create(&distPhoria).Error; err != nil {
		return nil, err
	}

	var nearPhoria preliminary.NearPhoriaTest
	if d := input.NearPhoriaTesting; d != nil {
		nearPhoria = preliminary.NearPhoriaTest{
			Horiz:            defaults.Str(d.Horiz),
			Vert:             defaults.Str(d.Vert),
			GradientRatio1:   defaults.Str(d.GradientRatio1),
			CalculatedRatio1: defaults.Str(d.CalculatedRatio1),
			GradientRatio2:   defaults.Str(d.GradientRatio2),
			CalculatedRatio2: defaults.Str(d.CalculatedRatio2),
			HorizExo:         boolDeref(d.HorizExo, false),
			HorizEso:         boolDeref(d.HorizEso, false),
			HorizOrtho:       boolDeref(d.HorizOrtho, false),
			VertRh:           boolDeref(d.VertRh, false),
			VertLn:           boolDeref(d.VertLn, false),
			VertOrtho:        boolDeref(d.VertOrtho, false),
		}
	}
	if err := s.db.Create(&nearPhoria).Error; err != nil {
		return nil, err
	}

	var distVergence preliminary.DistVergenceTest
	if d := input.DistVergenceTesting; d != nil {
		distVergence = preliminary.DistVergenceTest{
			Bi1: defaults.Str(d.Bi1), Bo1: defaults.Str(d.Bo1),
			Bi2: defaults.Str(d.Bi2), Bo2: defaults.Str(d.Bo2),
			Bi3: defaults.Str(d.Bi3), Bo3: defaults.Str(d.Bo3),
		}
	}
	if err := s.db.Create(&distVergence).Error; err != nil {
		return nil, err
	}

	var nearVergence preliminary.NearVergenceTest
	if d := input.NearVergenceTesting; d != nil {
		nearVergence = preliminary.NearVergenceTest{
			Bi1: defaults.Str(d.Bi1), Bo1: defaults.Str(d.Bo1),
			Bi2: defaults.Str(d.Bi2), Bo2: defaults.Str(d.Bo2),
			Bi3: defaults.Str(d.Bi3), Bo3: defaults.Str(d.Bo3),
		}
	}
	if err := s.db.Create(&nearVergence).Error; err != nil {
		return nil, err
	}

	var accommodation preliminary.Accommodation
	if d := input.Accommodation; d != nil {
		accommodation = preliminary.Accommodation{
			Pra1:                 defaults.Str(d.Pra1),
			Nra1:                 defaults.Str(d.Nra1),
			Pra2:                 defaults.Str(d.Pra2),
			Nra2:                 defaults.Str(d.Nra2),
			MemOd:                defaults.Str(d.MemOd),
			MemOs:                defaults.Str(d.MemOs),
			Baf:                  defaults.Str(d.Baf),
			VergenceFacilityCpm:  defaults.Str(d.VergenceFacilityCpm),
			VergenceFacilityWith: defaults.Str(d.VergenceFacilityWith),
			PushUpOd:             defaults.Str(d.PushUpOd),
			PushUpOs:             defaults.Str(d.PushUpOs),
			PushUpOu:             defaults.Str(d.PushUpOu),
			SlowWith:             defaults.Bool(d.SlowWith),
		}
	}
	if err := s.db.Create(&accommodation).Error; err != nil {
		return nil, err
	}

	dpID := i64(distPhoria.IDDistPhoriaTest)
	npID := i64(nearPhoria.IDNearPhoriaTest)
	dvID := i64(distVergence.IDDistVergenceTest)
	nvID := i64(nearVergence.IDNearVergenceTest)
	acID := i64(accommodation.IDAccommodation)

	npt := preliminary.NearPointTesting{
		DistPhoriaTestingID:   &dpID,
		NearPhoriaTestingID:   &npID,
		DistVergenceTestingID: &dvID,
		NearVergenceTestingID: &nvID,
		AccommodationID:       &acID,
	}
	if err := s.db.Create(&npt).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":              "near point testing created successfully",
		"id_near_point_testing": npt.IDNearPointTesting,
		"data":                 npt.ToMap(),
	}, nil
}

// ─── UpdateNearPointTesting ───────────────────────────────────────────────────

func (s *Service) UpdateNearPointTesting(username string, examID int64, input NearPointTestingInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}

	if exam.EmployeeID != i64(emp.IDEmployee) {
		return nil, errors.New("you are not authorized to update this exam")
	}
	if exam.Passed {
		return nil, errors.New("cannot update near point testing for a completed exam")
	}

	var prelim preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error; err != nil {
		return nil, errors.New("preliminary record not found for this exam")
	}

	if prelim.NearPointTestingID == nil {
		return nil, errors.New("near point testing record not found for this exam")
	}

	var npt preliminary.NearPointTesting
	if err := s.db.First(&npt, *prelim.NearPointTestingID).Error; err != nil {
		return nil, errors.New("near point testing record not found")
	}

	if input.DistPhoriaTesting != nil {
		d := input.DistPhoriaTesting
		var r preliminary.DistPhoriaTest
		if npt.DistPhoriaTestingID != nil {
			s.db.First(&r, *npt.DistPhoriaTestingID)
		}
		if d.Horiz != nil { r.Horiz = d.Horiz }
		if d.Vert != nil { r.Vert = d.Vert }
		if d.HorizExo != nil { r.HorizExo = *d.HorizExo }
		if d.HorizEso != nil { r.HorizEso = *d.HorizEso }
		if d.HorizOrtho != nil { r.HorizOrtho = *d.HorizOrtho }
		if d.VertRh != nil { r.VertRh = *d.VertRh }
		if d.VertLn != nil { r.VertLn = *d.VertLn }
		if d.VertOrtho != nil { r.VertOrtho = *d.VertOrtho }
		s.db.Save(&r)
		dpID := i64(r.IDDistPhoriaTest)
		npt.DistPhoriaTestingID = &dpID
	}

	if input.NearPhoriaTesting != nil {
		d := input.NearPhoriaTesting
		var r preliminary.NearPhoriaTest
		if npt.NearPhoriaTestingID != nil {
			s.db.First(&r, *npt.NearPhoriaTestingID)
		}
		if d.Horiz != nil { r.Horiz = d.Horiz }
		if d.Vert != nil { r.Vert = d.Vert }
		if d.GradientRatio1 != nil { r.GradientRatio1 = d.GradientRatio1 }
		if d.CalculatedRatio1 != nil { r.CalculatedRatio1 = d.CalculatedRatio1 }
		if d.GradientRatio2 != nil { r.GradientRatio2 = d.GradientRatio2 }
		if d.CalculatedRatio2 != nil { r.CalculatedRatio2 = d.CalculatedRatio2 }
		if d.HorizExo != nil { r.HorizExo = *d.HorizExo }
		if d.HorizEso != nil { r.HorizEso = *d.HorizEso }
		if d.HorizOrtho != nil { r.HorizOrtho = *d.HorizOrtho }
		if d.VertRh != nil { r.VertRh = *d.VertRh }
		if d.VertLn != nil { r.VertLn = *d.VertLn }
		if d.VertOrtho != nil { r.VertOrtho = *d.VertOrtho }
		s.db.Save(&r)
		npID := i64(r.IDNearPhoriaTest)
		npt.NearPhoriaTestingID = &npID
	}

	if input.DistVergenceTesting != nil {
		d := input.DistVergenceTesting
		var r preliminary.DistVergenceTest
		if npt.DistVergenceTestingID != nil {
			s.db.First(&r, *npt.DistVergenceTestingID)
		}
		if d.Bi1 != nil { r.Bi1 = d.Bi1 }
		if d.Bo1 != nil { r.Bo1 = d.Bo1 }
		if d.Bi2 != nil { r.Bi2 = d.Bi2 }
		if d.Bo2 != nil { r.Bo2 = d.Bo2 }
		if d.Bi3 != nil { r.Bi3 = d.Bi3 }
		if d.Bo3 != nil { r.Bo3 = d.Bo3 }
		s.db.Save(&r)
		dvID := i64(r.IDDistVergenceTest)
		npt.DistVergenceTestingID = &dvID
	}

	if input.NearVergenceTesting != nil {
		d := input.NearVergenceTesting
		var r preliminary.NearVergenceTest
		if npt.NearVergenceTestingID != nil {
			s.db.First(&r, *npt.NearVergenceTestingID)
		}
		if d.Bi1 != nil { r.Bi1 = d.Bi1 }
		if d.Bo1 != nil { r.Bo1 = d.Bo1 }
		if d.Bi2 != nil { r.Bi2 = d.Bi2 }
		if d.Bo2 != nil { r.Bo2 = d.Bo2 }
		if d.Bi3 != nil { r.Bi3 = d.Bi3 }
		if d.Bo3 != nil { r.Bo3 = d.Bo3 }
		s.db.Save(&r)
		nvID := i64(r.IDNearVergenceTest)
		npt.NearVergenceTestingID = &nvID
	}

	if input.Accommodation != nil {
		d := input.Accommodation
		var r preliminary.Accommodation
		if npt.AccommodationID != nil {
			s.db.First(&r, *npt.AccommodationID)
		}
		if d.Pra1 != nil { r.Pra1 = d.Pra1 }
		if d.Nra1 != nil { r.Nra1 = d.Nra1 }
		if d.Pra2 != nil { r.Pra2 = d.Pra2 }
		if d.Nra2 != nil { r.Nra2 = d.Nra2 }
		if d.MemOd != nil { r.MemOd = d.MemOd }
		if d.MemOs != nil { r.MemOs = d.MemOs }
		if d.Baf != nil { r.Baf = d.Baf }
		if d.VergenceFacilityCpm != nil { r.VergenceFacilityCpm = d.VergenceFacilityCpm }
		if d.VergenceFacilityWith != nil { r.VergenceFacilityWith = d.VergenceFacilityWith }
		if d.PushUpOd != nil { r.PushUpOd = d.PushUpOd }
		if d.PushUpOs != nil { r.PushUpOs = d.PushUpOs }
		if d.PushUpOu != nil { r.PushUpOu = d.PushUpOu }
		if d.SlowWith != nil { r.SlowWith = d.SlowWith }
		s.db.Save(&r)
		acID := i64(r.IDAccommodation)
		npt.AccommodationID = &acID
	}

	if err := s.db.Save(&npt).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "near point testing updated successfully",
		"data":    npt.ToMap(),
	}, nil
}

// ─── GetNearPointTesting ──────────────────────────────────────────────────────

func (s *Service) GetNearPointTesting(examID int64) (map[string]interface{}, error) {
	if _, err := s.getExam(examID); err != nil {
		return nil, err
	}

	var prelim preliminary.PreliminaryEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error; err != nil {
		return map[string]interface{}{
			"exam_id":           examID,
			"exists":            false,
			"near_point_testing": nil,
		}, nil
	}

	if prelim.NearPointTestingID == nil {
		return map[string]interface{}{
			"exam_id":           examID,
			"exists":            false,
			"near_point_testing": nil,
		}, nil
	}

	var npt preliminary.NearPointTesting
	if err := s.db.First(&npt, *prelim.NearPointTestingID).Error; err != nil {
		return map[string]interface{}{
			"exam_id":           examID,
			"exists":            false,
			"near_point_testing": nil,
		}, nil
	}

	nptMap := npt.ToMap()

	var dp preliminary.DistPhoriaTest
	if npt.DistPhoriaTestingID != nil && s.db.First(&dp, *npt.DistPhoriaTestingID).Error == nil {
		nptMap["dist_phoria_testing"] = dp.ToMap()
	} else {
		nptMap["dist_phoria_testing"] = nil
	}

	var np preliminary.NearPhoriaTest
	if npt.NearPhoriaTestingID != nil && s.db.First(&np, *npt.NearPhoriaTestingID).Error == nil {
		nptMap["near_phoria_testing"] = np.ToMap()
	} else {
		nptMap["near_phoria_testing"] = nil
	}

	var dv preliminary.DistVergenceTest
	if npt.DistVergenceTestingID != nil && s.db.First(&dv, *npt.DistVergenceTestingID).Error == nil {
		nptMap["dist_vergence_testing"] = dv.ToMap()
	} else {
		nptMap["dist_vergence_testing"] = nil
	}

	var nv preliminary.NearVergenceTest
	if npt.NearVergenceTestingID != nil && s.db.First(&nv, *npt.NearVergenceTestingID).Error == nil {
		nptMap["near_vergence_testing"] = nv.ToMap()
	} else {
		nptMap["near_vergence_testing"] = nil
	}

	var ac preliminary.Accommodation
	if npt.AccommodationID != nil && s.db.First(&ac, *npt.AccommodationID).Error == nil {
		nptMap["accommodation"] = ac.ToMap()
	} else {
		nptMap["accommodation"] = nil
	}

	return map[string]interface{}{
		"exam_id":           examID,
		"exists":            true,
		"near_point_testing": nptMap,
	}, nil
}
