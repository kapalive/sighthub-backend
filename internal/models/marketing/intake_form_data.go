package marketing

import "time"

type IntakeFormData struct {
	IDIntakeFormData    int64      `gorm:"column:id_intake_form_data;primaryKey"                    json:"id_intake_form_data"`
	LastName            *string    `gorm:"column:last_name;type:text"                               json:"last_name,omitempty"`
	FirstName           *string    `gorm:"column:first_name;type:text"                              json:"first_name,omitempty"`
	MiddleInitial       *string    `gorm:"column:middle_initial;type:text"                          json:"middle_initial,omitempty"`
	DateOfBirth         *time.Time `gorm:"column:date_of_birth;type:date"                           json:"date_of_birth,omitempty"`
	Age                 *int16     `gorm:"column:age"                                               json:"age,omitempty"`
	Gender              *string    `gorm:"column:gender;type:text"                                  json:"gender,omitempty"`
	CellPhone           *string    `gorm:"column:cell_phone;type:text"                              json:"cell_phone,omitempty"`
	HomePhone           *string    `gorm:"column:home_phone;type:text"                              json:"home_phone,omitempty"`
	Address             *string    `gorm:"column:address;type:text"                                 json:"address,omitempty"`
	City                *string    `gorm:"column:city;type:text"                                    json:"city,omitempty"`
	Zip                 *string    `gorm:"column:zip;type:text"                                     json:"zip,omitempty"`
	ReferredBy          *string    `gorm:"column:referred_by;type:text"                             json:"referred_by,omitempty"`
	EmailAddress        *string    `gorm:"column:email_address;type:text"                           json:"email_address,omitempty"`
	Insurance           *string    `gorm:"column:insurance;type:text"                               json:"insurance,omitempty"`
	PolicyHolderName    *string    `gorm:"column:policy_holder_name;type:text"                      json:"policy_holder_name,omitempty"`
	PolicyHolderDob     *time.Time `gorm:"column:policy_holder_dob;type:date"                       json:"policy_holder_dob,omitempty"`
	LastEyeExam         *string    `gorm:"column:last_eye_exam;type:text"                           json:"last_eye_exam,omitempty"`
	BeenHereBefore      *string    `gorm:"column:been_here_before;type:text"                        json:"been_here_before,omitempty"`
	WhenBefore          *string    `gorm:"column:when_before;type:text"                             json:"when_before,omitempty"`
	EmergencyContact    *string    `gorm:"column:emergency_contact;type:text"                       json:"emergency_contact,omitempty"`
	WearsGlasses        *string    `gorm:"column:wears_glasses;type:text"                           json:"wears_glasses,omitempty"`
	PrimaryCareProvider *string    `gorm:"column:primary_care_provider;type:text"                   json:"primary_care_provider,omitempty"`
	PrimaryCarePhone    *string    `gorm:"column:primary_care_phone;type:text"                      json:"primary_care_phone,omitempty"`
	Surgeries           *string    `gorm:"column:surgeries;type:text"                               json:"surgeries,omitempty"`
	DiagnosisType       *string    `gorm:"column:diagnosis_type;type:text"                          json:"diagnosis_type,omitempty"`
	DiagnosisDate       *time.Time `gorm:"column:diagnosis_date;type:date"                          json:"diagnosis_date,omitempty"`
	OtherConditions     *string    `gorm:"column:other_conditions;type:text"                        json:"other_conditions,omitempty"`
	BlurryVision        *bool      `gorm:"column:blurry_vision"                                     json:"blurry_vision,omitempty"`
	BlurryVisionDistance *string   `gorm:"column:blurry_vision_distance;type:text"                  json:"blurry_vision_distance,omitempty"`
	BlurryVisionNear    *string    `gorm:"column:blurry_vision_near;type:text"                      json:"blurry_vision_near,omitempty"`
	Itching             *bool      `gorm:"column:itching"                                           json:"itching,omitempty"`
	DoubleVision        *bool      `gorm:"column:double_vision"                                     json:"double_vision,omitempty"`
	Burning             *bool      `gorm:"column:burning"                                           json:"burning,omitempty"`
	EyeInjury           *bool      `gorm:"column:eye_injury"                                        json:"eye_injury,omitempty"`
	EyeInfection        *bool      `gorm:"column:eye_infection"                                     json:"eye_infection,omitempty"`
	Tearing             *bool      `gorm:"column:tearing"                                           json:"tearing,omitempty"`
	FloatersSpots       *bool      `gorm:"column:floaters_spots"                                    json:"floaters_spots,omitempty"`
	FlashesOfLight      *bool      `gorm:"column:flashes_of_light"                                  json:"flashes_of_light,omitempty"`
	Pain                *bool      `gorm:"column:pain"                                              json:"pain,omitempty"`
	LightSensitivity    *bool      `gorm:"column:light_sensitivity"                                 json:"light_sensitivity,omitempty"`
	Cataracts           *bool      `gorm:"column:cataracts"                                         json:"cataracts,omitempty"`
	Glaucoma            *bool      `gorm:"column:glaucoma"                                          json:"glaucoma,omitempty"`
	RetinalProblems     *bool      `gorm:"column:retinal_problems"                                  json:"retinal_problems,omitempty"`
	CornealProblems     *bool      `gorm:"column:corneal_problems"                                  json:"corneal_problems,omitempty"`
	CataractSurgery     *bool      `gorm:"column:cataract_surgery"                                  json:"cataract_surgery,omitempty"`
	CataractSurgeryEye  *string    `gorm:"column:cataract_surgery_eye;type:text"                    json:"cataract_surgery_eye,omitempty"`
	DateAndDoctorName   *string    `gorm:"column:date_and_doctor_name;type:text"                    json:"date_and_doctor_name,omitempty"`
	SmokingStatus       *string    `gorm:"column:smoking_status;type:text"                          json:"smoking_status,omitempty"`
	DrinkingStatus      *string    `gorm:"column:drinking_status;type:text"                         json:"drinking_status,omitempty"`
	Signature           *string    `gorm:"column:signature;type:text"                               json:"signature,omitempty"`
	SignatureDate       *time.Time `gorm:"column:signature_date;type:date"                          json:"signature_date,omitempty"`
	PatientOldRx        *string    `gorm:"column:patient_old_rx;type:text"                          json:"patient_old_rx,omitempty"`
	CreatedAt           *time.Time `gorm:"column:created_at;type:timestamptz;default:now()"         json:"created_at,omitempty"`
	SignaturePath       *string    `gorm:"column:signature_path;type:text"                          json:"signature_path,omitempty"`
	AppointmentID       int64      `gorm:"column:appointment_id;not null"                           json:"appointment_id"`

	MedicalHistory []IntakeFormMedicalHistory `gorm:"foreignKey:RequestID" json:"medical_history,omitempty"`
	Medications    []IntakeFormMedications    `gorm:"foreignKey:RequestID" json:"medications,omitempty"`
	Allergies      []IntakeFormAllergies      `gorm:"foreignKey:RequestID" json:"allergies,omitempty"`
}

func (IntakeFormData) TableName() string { return "intake_form_data" }
