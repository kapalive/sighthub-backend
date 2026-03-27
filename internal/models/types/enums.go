package types

// ===== Gender =====
type Gender string

const (
	GenderMale           Gender = "Male"
	GenderFemale         Gender = "Female"
	GenderOther          Gender = "Other"
	GenderUnisexNotToSay Gender = "Prefer not to say"
	GenderUnisexUnknown  Gender = "Unknown"
)

// ===== TypeProducts =====
type TypeProducts string

const (
	// canonical
	TypeProductsEyeglasses    TypeProducts = "eyeglasses"
	TypeProductsSunglasses    TypeProducts = "sunglasses"
	TypeProductsSafetyGlasses TypeProducts = "safety_glasses"
	// backward-compat aliases (если где-то уже юзались старые имена)
	FrameTypeEyeglasses TypeProducts = TypeProductsEyeglasses
	FrameTypeSunglasses TypeProducts = TypeProductsSunglasses
)

// ===== LensType =====
type LensType string

const (
	LensTypeSingleVision LensType = "single_vision"
	LensTypeBifocal      LensType = "bifocal"
	LensTypeProgressive  LensType = "progressive"
	LensTypeTrifocal     LensType = "trifocal"
)

// ===== Prism directions =====
type HPrismDirection string

const (
	// правильные горизонтальные направления
	HPrismDirectionBI HPrismDirection = "BI"
	HPrismDirectionBO HPrismDirection = "BO"
	// алиасы для совместимости (раньше были перепутаны названия)
	HPrismDirectionBU HPrismDirection = HPrismDirectionBI // DEPRECATED
	HPrismDirectionBD HPrismDirection = HPrismDirectionBO // DEPRECATED
)

type VPrismDirection string

const (
	VPrismDirectionBU VPrismDirection = "BU"
	VPrismDirectionBD VPrismDirection = "BD"
)

// ===== Commission keys =====
type CommissionPBKey string

const (
	CommissionPBKeyFrames      CommissionPBKey = "Frames"
	CommissionPBKeyLens        CommissionPBKey = "Lens"
	CommissionPBKeyContactLens CommissionPBKey = "Contact Lens"
	CommissionPBKeyProfService CommissionPBKey = "Prof. service"
	CommissionPBKeyTreatment   CommissionPBKey = "Treatment"
	CommissionPBKeyAddService  CommissionPBKey = "Add service"
)

// ===== Contact lens type =====
type ContactLensTypeEnum string

const (
	ContactLensTypePatientRx ContactLensTypeEnum = "Patient Rx"
	ContactLensTypeTrial     ContactLensTypeEnum = "Trial"
)

// ===== Inventory status =====
type StatusItemsInventory string

const (
	StatusInventoryReadyForSale          StatusItemsInventory = "Ready for Sale"
	StatusInventoryDefective             StatusItemsInventory = "Defective"
	StatusInventoryOnReturn              StatusItemsInventory = "On Return"
	StatusInventoryICTToReceiveInMN      StatusItemsInventory = "ICT (to receive in MN)"
	StatusInventoryICTSentAndNotReceived StatusItemsInventory = "ICT (sent and not received)"
	StatusInventorySOLD                  StatusItemsInventory = "SOLD"
	StatusInventoryMissing               StatusItemsInventory = "Missing"
	StatusInventoryRemoved               StatusItemsInventory = "Removed"
	StatusInventoryOrdered               StatusItemsInventory = "Ordered"
)

// ===== Return reason (vendor) =====
type ReasonReturnVendor string

const (
	ReasonReturnVendorDefective  ReasonReturnVendor = "Defective"
	ReasonReturnVendorNotForSale ReasonReturnVendor = "Not for Sale"
)

// ===== Lens material =====
type LensMaterial string

const (
	LensMaterialPlastic       LensMaterial = "Plastic"
	LensMaterialGlass         LensMaterial = "Glass"
	LensMaterialPolycarbonate LensMaterial = "Polycarbonate"
	LensMaterial160I          LensMaterial = "1.6O Index" // так в БД
	LensMaterialCR39          LensMaterial = "CR39 - Plastic"
	LensMaterialHI160         LensMaterial = "High Index 1.60"
	LensMaterialHI166         LensMaterial = "High Index 1.66/1.67"
	LensMaterialHI174         LensMaterial = "High Index 1.74"
	LensMaterialTrivex        LensMaterial = "Trivex"
	LensMaterialUnknown       LensMaterial = "n/a"
)

// ===== Pronoun =====
type Pronoun string

const (
	PronounHe    Pronoun = "He/Him"
	PronounShe   Pronoun = "She/Her"
	PronounThey  Pronoun = "They/Them"
	PronounOther Pronoun = "Other"
)

// ===== Insurance status =====
type PaidInsuranceStatus string

const (
	PaidInsuranceStatusBilled  PaidInsuranceStatus = "Billed"
	PaidInsuranceStatusPending PaidInsuranceStatus = "Pending"
	PaidInsuranceStatusPaid    PaidInsuranceStatus = "Paid"
)

// ===== Dominant eye =====
type DominantEye string

const (
	DominantEyeRight   DominantEye = "Right"
	DominantEyeLeft    DominantEye = "Left"
	DominantEyeUnknown DominantEye = "n/a"
)

// ===== Angle estimation eye =====
type AngleEstimationEye string

const (
	AngleEstimationEye1       AngleEstimationEye = "1"
	AngleEstimationEye2       AngleEstimationEye = "2"
	AngleEstimationEye3       AngleEstimationEye = "3"
	AngleEstimationEye4       AngleEstimationEye = "4"
	AngleEstimationEyeUnknown AngleEstimationEye = "n/a"
)

// ===== Iris color =====
type IrisColor string

const (
	IrisColorBlue          IrisColor = "Blue"
	IrisColorGreen         IrisColor = "Green"
	IrisColorHazel         IrisColor = "Hazel"
	IrisColorBrown         IrisColor = "Brown"
	IrisColorHeterochromia IrisColor = "Heterochromia"
	IrisColorUnknown       IrisColor = "n/a"
)

// ===== Lab ticket: lens status =====
type LabTicketLensStatus string

const (
	LabTicketLensStatusUncut    LabTicketLensStatus = "Uncut"
	LabTicketLensStatusCut      LabTicketLensStatus = "Cut"
	LabTicketLensStatusNoLenses LabTicketLensStatus = "No Lenses"
)

// ===== Lab ticket: lens order =====
type LabTicketLensOrder string

const (
	LabTicketLensOrderPair              LabTicketLensOrder = "Pair"
	LabTicketLensOrderRtEyeOnly         LabTicketLensOrder = "Rt Eye Only"
	LabTicketLensOrderLtEyeOnly         LabTicketLensOrder = "Lt Eye Only"
	LabTicketLensOrderRtEyeOnlyLtEyeBal LabTicketLensOrder = "Rt Eye Only, Lt Eye Bal"
	LabTicketLensOrderLtEyeOnlyRtEyeBal LabTicketLensOrder = "Lt Eye Only, Rt Eye Bal"
)

// ===== Lab ticket: frame status =====
type LabTicketFrameStatus string

const (
	LabTicketFrameStatusFrameInStore      LabTicketFrameStatus = "Frame in Store"
	LabTicketFrameStatusSpecialOrderFrame LabTicketFrameStatus = "Special Order Frame"
	LabTicketFrameStatusCallForFrame      LabTicketFrameStatus = "Call for Frame"
	LabTicketFrameStatusFrameToOurLab     LabTicketFrameStatus = "Frame to Our Lab"
	LabTicketFrameStatusFrameToVendorLab  LabTicketFrameStatus = "Frame to Vendor Lab"
)

// ===== Communication types =====
type CommunicationType string

const (
	CommunicationTypeCall  CommunicationType = "call"
	CommunicationTypeEmail CommunicationType = "email"
	CommunicationTypeMail  CommunicationType = "mail"
	CommunicationTypeVisit CommunicationType = "visit"
)

// ===== Contact W type (DW/EW) =====
type ContactWType string

const (
	ContactWTypeDW ContactWType = "DW"
	ContactWTypeEW ContactWType = "EW"
)
