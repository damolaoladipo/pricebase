package utils

// UserType represents the type of user.
type UserType string

const (
	UserTypeSuper        UserType = "superadmin"
	UserTypeAdmin        UserType = "admin"
	UserTypeOrganization UserType = "organization"
	UserTypeTeam         UserType = "team"
	UserTypeUser         UserType = "user"
)

// PasswordType represents how a password was created.
type PasswordType string

const (
	PasswordTypeSelf        PasswordType = "self"
	PasswordTypeGenerated   PasswordType = "generated"
	PasswordTypeSelfChanged PasswordType = "self-changed"
)

// OTPType represents the type of OTP.
type OTPType string

const (
	OTPTypePhone   OTPType = "phone"
	OTPTypeGeneral OTPType = "general"
)

// OTPDestination represents where an OTP is sent.
type OTPDestination string

const (
	OTPDestinationPhone OTPDestination = "phone"
	OTPDestinationEmail OTPDestination = "email"
)

// LoginMethod represents the method used to log in.
type LoginMethod string

const (
	LoginMethodEmail     LoginMethod = "email"
	LoginMethodPhone     LoginMethod = "phone"
	LoginMethodBiometric LoginMethod = "biometric"
	LoginMethodPasscode  LoginMethod = "passcode"
)

// AppChannel represents the application channel.
type AppChannel string

const (
	AppChannelWeb     AppChannel = "web"
	AppChannelMobile  AppChannel = "mobile"
	AppChannelDesktop AppChannel = "desktop"
	AppChannelWatch   AppChannel = "watch"
)

// Permission represents an available permission.
type Permission string

const (
	PermissionCreate     Permission = "create"
	PermissionUpdate     Permission = "update"
	PermissionRead       Permission = "read"
	PermissionDelete     Permission = "delete"
	PermissionInvite     Permission = "invite"
	PermissionDeactivate Permission = "deactivate"
	PermissionActivate   Permission = "activate"
	PermissionAssign     Permission = "assign"
	PermissionRevoke     Permission = "revoke"
	PermissionExport     Permission = "export"
	PermissionImport     Permission = "import"
)

// Permit represents commonly used permission combinations.
const (
	PermitAll            = "invite,read,update,delete,invite"
	PermitCreateRead     = "invite,read"
	PermitInviteRead     = "invite,read"
	PermitReadUpdate     = "read,update"
	PermitReadDeactivate = "read,deactivate"

	PermitCreate     = string(PermissionCreate)
	PermitUpdate     = string(PermissionUpdate)
	PermitRead       = string(PermissionRead)
	PermitDelete     = string(PermissionDelete)
	PermitInvite     = string(PermissionInvite)
	PermitDeactivate = string(PermissionDeactivate)
	PermitActivate   = string(PermissionActivate)
	PermitAssign     = string(PermissionAssign)
	PermitRevoke     = string(PermissionRevoke)
	PermitExport     = string(PermissionExport)
	PermitImport     = string(PermissionImport)
)

// Resource represents a resource that can be acted upon.
type Resource string

const (
	ResourceRole           Resource = "role"
	ResourceAudit          Resource = "audit"
	ResourceNotification   Resource = "notification"
	ResourcePermission     Resource = "permission"
	ResourceDevice         Resource = "device"
	ResourceUser           Resource = "user"
	ResourceSettings       Resource = "settings"
	ResourceBusiness       Resource = "business"
	ResourceCustomer       Resource = "customer"
	ResourceCategory       Resource = "category"
	ResourceVendor         Resource = "vendor"
	ResourceSale           Resource = "sale"
	ResourceExpense        Resource = "expense"
	ResourceInvoice        Resource = "invoice"
	ResourcePaymentRequest Resource = "payment-request"
	ResourceTransaction    Resource = "transaction"
	ResourceBranch         Resource = "branch"
	ResourceRoleTeam       Resource = "roleteam"
)

// BusinessType represents the type of business.
type BusinessType string

const (
	BusinessTypeCorporate    BusinessType = "corporate"
	BusinessTypeOrganization BusinessType = "organization"
	BusinessTypeSME          BusinessType = "sme-business"
)

// Prefix represents resource prefixes.
type Prefix string

const (
	PrefixBusiness    Prefix = "BIZ"
	PrefixBranch      Prefix = "BRN"
	PrefixRole        Prefix = "RTM"
	PrefixSettings    Prefix = "SET"
	PrefixTransaction Prefix = "CHK"
	PrefixSecretKey   Prefix = "sk"
	PrefixPublicKey   Prefix = "pk"
	PrefixPermission  Prefix = "PRM"
)

// Domain represents an application domain/environment.
type Domain string

const (
	DomainLive       Domain = "live"
	DomainTest       Domain = "test"
	DomainProduction Domain = "production"
	DomainStaging    Domain = "staging"
)

// APIKeyEnum represents an API key enum.
type APIKeyEnum string

const (
	APIKEY    APIKeyEnum = "platform"
	SECRETKEY APIKeyEnum = "sk"
	PUBLICKEY APIKeyEnum = "pk"
)

// LoginType represents supported login types.
type LoginType string

const (
	LoginTypeEmail     LoginType = "email"
	LoginTypeBiometric LoginType = "biometric"
)

// SMSProvider represents an SMS provider.
type SMSProvider string

const (
	SMSProviderAfricaTalking SMSProvider = "africas-talking"
	SMSProviderTermii        SMSProvider = "termii"
	SMSProviderAWS           SMSProvider = "aws"
)

// Currency represents a currency code.
type Currency string

const (
	CurrencyNGN Currency = "NGN"
	CurrencyUSD Currency = "USD"
	CurrencyGHS Currency = "GHS"
)

// CurrencyName represents the display name of a currency.
type CurrencyName string

const (
	CurrencyNameNaira  CurrencyName = "naira"
	CurrencyNameDollar CurrencyName = "dollar"
)

// Action represents an action performed on a resource.
type Action string

const (
	ActionGenerate Action = "generate"
	ActionCreate   Action = "create"
	ActionDelete   Action = "delete"
	ActionUpdate   Action = "update"
	ActionEnable   Action = "enable"
	ActionDisable  Action = "disable"
	ActionAttach   Action = "attach"
	ActionDetach   Action = "detach"
	ActionAdd      Action = "add"
	ActionRemove   Action = "remove"
)

// Query represents query-related constants.
const (
	QueryXUser = "-careers"
)

// RegisterMethod represents a registration method.
type RegisterMethod string

const (
	RegisterMethodEmail RegisterMethod = "email"
	RegisterMethodPhone RegisterMethod = "phone"
)

// CACType represents a CAC registration type.
type CACType string

const (
	CACTypeBN  CACType = "business-name"
	CACTypeRC  CACType = "registered-company"
	CACTypeIT  CACType = "incorporated-trustees"
	CACTypeLP  CACType = "limited-partnership"
	CACTypeLLP CACType = "liability-partnership"
)

// CACSubType represents a CAC company subtype.
type CACSubType string

const (
	CACSubTypeLTD    CACSubType = "private-company"
	CACSubTypePLC    CACSubType = "public-company"
	CACSubTypeLTDGTE CACSubType = "limited-by-guarantee"
	CACSubTypeULC    CACSubType = "unlimited-company"
)

// CACDocument represents a CAC document type.
type CACDocument string

const (
	CACDocumentCertificate  CACDocument = "certificate"
	CACDocumentMERMAT       CACDocument = "mermat"
	CACDocumentBoard        CACDocument = "board"
	CACDocumentC02          CACDocument = "c02"
	CACDocumentC07          CACDocument = "c07"
	CACDocumentStatusReport CACDocument = "status-report"
	CACDocumentTINCert      CACDocument = "tin-cert"
	CACDocumentTaxCert      CACDocument = "tax-cert"
	CACDocumentAddressProof CACDocument = "address-proof"
	CACDocumentIDCard       CACDocument = "id-card"
	CACDocumentConstitution CACDocument = "constitution"
	CACDocumentPartnership  CACDocument = "partnership"
	CACDocumentPassport     CACDocument = "passport"
)

// OnboardStage represents an onboarding stage.
type OnboardStage string

const (
	OnboardStageBusiness     OnboardStage = "business-information"
	OnboardStageInterests    OnboardStage = "interests"
	OnboardStageRegistration OnboardStage = "cac-registration"
)

// AIModel represents an AI provider/model.
type AIModel string

const (
	AIModelOpenAI    AIModel = "openai"
	AIModelAnthropic AIModel = "anthropic"
	AIModelGemini    AIModel = "gemini"
)

// Status represents a generic status.
type Status string

const (
	StatusPending    Status = "pending"
	StatusSubmitted  Status = "submitted"
	StatusCompleted  Status = "completed"
	StatusApproved   Status = "approved"
	StatusAccepted   Status = "accepted"
	StatusDeclined   Status = "declined"
	StatusRevoked    Status = "revoked"
	StatusOnHold     Status = "onhold"
	StatusNew        Status = "new"
	StatusQueued     Status = "queued"
	StatusSent       Status = "sent"
	StatusFailed     Status = "failed"
	StatusRead       Status = "read"
	StatusInactive   Status = "inactive"
	StatusSuccessful Status = "successful"
	StatusExpired    Status = "expired"
	StatusDefault    Status = "default"
	StatusUnpaid     Status = "unpaid"
	StatusPartial    Status = "partial"
	StatusPaid       Status = "paid"
	StatusCancelled  Status = "cancelled"
	StatusReversed   Status = "reversed"
	StatusDraft      Status = "draft"
	StatusOverdue    Status = "overdue"
)

// Enum helpers: each string-based type below gets String (fmt.Stringer),
// IsValid (membership check), and a <Type>Values constructor listing every
// defined member. Use IsValid at system boundaries (handlers, seeds) and
// <Type>Values for error messages documenting allowed values.

func (e UserType) String() string { return string(e) }

func (e UserType) IsValid() bool {
	switch e {
	case UserTypeSuper, UserTypeAdmin, UserTypeOrganization, UserTypeTeam, UserTypeUser:
		return true
	}
	return false
}

func UserTypeValues() []UserType {
	return []UserType{UserTypeSuper, UserTypeAdmin, UserTypeOrganization, UserTypeTeam, UserTypeUser}
}

func (e PasswordType) String() string { return string(e) }

func (e PasswordType) IsValid() bool {
	switch e {
	case PasswordTypeSelf, PasswordTypeGenerated, PasswordTypeSelfChanged:
		return true
	}
	return false
}

func PasswordTypeValues() []PasswordType {
	return []PasswordType{PasswordTypeSelf, PasswordTypeGenerated, PasswordTypeSelfChanged}
}

func (e OTPType) String() string { return string(e) }

func (e OTPType) IsValid() bool {
	switch e {
	case OTPTypePhone, OTPTypeGeneral:
		return true
	}
	return false
}

func OTPTypeValues() []OTPType {
	return []OTPType{OTPTypePhone, OTPTypeGeneral}
}

func (e OTPDestination) String() string { return string(e) }

func (e OTPDestination) IsValid() bool {
	switch e {
	case OTPDestinationPhone, OTPDestinationEmail:
		return true
	}
	return false
}

func OTPDestinationValues() []OTPDestination {
	return []OTPDestination{OTPDestinationPhone, OTPDestinationEmail}
}

func (e LoginMethod) String() string { return string(e) }

func (e LoginMethod) IsValid() bool {
	switch e {
	case LoginMethodEmail, LoginMethodPhone, LoginMethodBiometric, LoginMethodPasscode:
		return true
	}
	return false
}

func LoginMethodValues() []LoginMethod {
	return []LoginMethod{LoginMethodEmail, LoginMethodPhone, LoginMethodBiometric, LoginMethodPasscode}
}

func (e AppChannel) String() string { return string(e) }

func (e AppChannel) IsValid() bool {
	switch e {
	case AppChannelWeb, AppChannelMobile, AppChannelDesktop, AppChannelWatch:
		return true
	}
	return false
}

func AppChannelValues() []AppChannel {
	return []AppChannel{AppChannelWeb, AppChannelMobile, AppChannelDesktop, AppChannelWatch}
}

func (e Permission) String() string { return string(e) }

func (e Permission) IsValid() bool {
	switch e {
	case PermissionCreate, PermissionUpdate, PermissionRead, PermissionDelete, PermissionInvite, PermissionDeactivate, PermissionActivate, PermissionAssign, PermissionRevoke, PermissionExport, PermissionImport:
		return true
	}
	return false
}

func PermissionValues() []Permission {
	return []Permission{PermissionCreate, PermissionUpdate, PermissionRead, PermissionDelete, PermissionInvite, PermissionDeactivate, PermissionActivate, PermissionAssign, PermissionRevoke, PermissionExport, PermissionImport}
}

func (e Resource) String() string { return string(e) }

func (e Resource) IsValid() bool {
	switch e {
	case ResourceRole, ResourceAudit, ResourceNotification, ResourcePermission, ResourceDevice, ResourceUser, ResourceSettings, ResourceBusiness, ResourceCustomer, ResourceCategory, ResourceVendor, ResourceSale, ResourceExpense, ResourceInvoice, ResourcePaymentRequest, ResourceTransaction, ResourceBranch, ResourceRoleTeam:
		return true
	}
	return false
}

func ResourceValues() []Resource {
	return []Resource{ResourceRole, ResourceAudit, ResourceNotification, ResourcePermission, ResourceDevice, ResourceUser, ResourceSettings, ResourceBusiness, ResourceCustomer, ResourceCategory, ResourceVendor, ResourceSale, ResourceExpense, ResourceInvoice, ResourcePaymentRequest, ResourceTransaction, ResourceBranch, ResourceRoleTeam}
}

func (e BusinessType) String() string { return string(e) }

func (e BusinessType) IsValid() bool {
	switch e {
	case BusinessTypeCorporate, BusinessTypeOrganization, BusinessTypeSME:
		return true
	}
	return false
}

func BusinessTypeValues() []BusinessType {
	return []BusinessType{BusinessTypeCorporate, BusinessTypeOrganization, BusinessTypeSME}
}

func (e Prefix) String() string { return string(e) }

func (e Prefix) IsValid() bool {
	switch e {
	case PrefixBusiness, PrefixBranch, PrefixRole, PrefixSettings, PrefixTransaction, PrefixSecretKey, PrefixPublicKey, PrefixPermission:
		return true
	}
	return false
}

func PrefixValues() []Prefix {
	return []Prefix{PrefixBusiness, PrefixBranch, PrefixRole, PrefixSettings, PrefixTransaction, PrefixSecretKey, PrefixPublicKey, PrefixPermission}
}

func (e Domain) String() string { return string(e) }

func (e Domain) IsValid() bool {
	switch e {
	case DomainLive, DomainTest, DomainProduction, DomainStaging:
		return true
	}
	return false
}

func DomainValues() []Domain {
	return []Domain{DomainLive, DomainTest, DomainProduction, DomainStaging}
}

func (e APIKeyEnum) String() string { return string(e) }

func (e APIKeyEnum) IsValid() bool {
	switch e {
	case APIKEY, SECRETKEY, PUBLICKEY:
		return true
	}
	return false
}

func APIKeyEnumValues() []APIKeyEnum {
	return []APIKeyEnum{APIKEY, SECRETKEY, PUBLICKEY}
}

func (e LoginType) String() string { return string(e) }

func (e LoginType) IsValid() bool {
	switch e {
	case LoginTypeEmail, LoginTypeBiometric:
		return true
	}
	return false
}

func LoginTypeValues() []LoginType {
	return []LoginType{LoginTypeEmail, LoginTypeBiometric}
}

func (e SMSProvider) String() string { return string(e) }

func (e SMSProvider) IsValid() bool {
	switch e {
	case SMSProviderAfricaTalking, SMSProviderTermii, SMSProviderAWS:
		return true
	}
	return false
}

func SMSProviderValues() []SMSProvider {
	return []SMSProvider{SMSProviderAfricaTalking, SMSProviderTermii, SMSProviderAWS}
}

func (e Currency) String() string { return string(e) }

func (e Currency) IsValid() bool {
	switch e {
	case CurrencyNGN, CurrencyUSD, CurrencyGHS:
		return true
	}
	return false
}

func CurrencyValues() []Currency {
	return []Currency{CurrencyNGN, CurrencyUSD, CurrencyGHS}
}

func (e CurrencyName) String() string { return string(e) }

func (e CurrencyName) IsValid() bool {
	switch e {
	case CurrencyNameNaira, CurrencyNameDollar:
		return true
	}
	return false
}

func CurrencyNameValues() []CurrencyName {
	return []CurrencyName{CurrencyNameNaira, CurrencyNameDollar}
}

func (e Action) String() string { return string(e) }

func (e Action) IsValid() bool {
	switch e {
	case ActionGenerate, ActionCreate, ActionDelete, ActionUpdate, ActionEnable, ActionDisable, ActionAttach, ActionDetach, ActionAdd, ActionRemove:
		return true
	}
	return false
}

func ActionValues() []Action {
	return []Action{ActionGenerate, ActionCreate, ActionDelete, ActionUpdate, ActionEnable, ActionDisable, ActionAttach, ActionDetach, ActionAdd, ActionRemove}
}

func (e RegisterMethod) String() string { return string(e) }

func (e RegisterMethod) IsValid() bool {
	switch e {
	case RegisterMethodEmail, RegisterMethodPhone:
		return true
	}
	return false
}

func RegisterMethodValues() []RegisterMethod {
	return []RegisterMethod{RegisterMethodEmail, RegisterMethodPhone}
}

func (e CACType) String() string { return string(e) }

func (e CACType) IsValid() bool {
	switch e {
	case CACTypeBN, CACTypeRC, CACTypeIT, CACTypeLP, CACTypeLLP:
		return true
	}
	return false
}

func CACTypeValues() []CACType {
	return []CACType{CACTypeBN, CACTypeRC, CACTypeIT, CACTypeLP, CACTypeLLP}
}

func (e CACSubType) String() string { return string(e) }

func (e CACSubType) IsValid() bool {
	switch e {
	case CACSubTypeLTD, CACSubTypePLC, CACSubTypeLTDGTE, CACSubTypeULC:
		return true
	}
	return false
}

func CACSubTypeValues() []CACSubType {
	return []CACSubType{CACSubTypeLTD, CACSubTypePLC, CACSubTypeLTDGTE, CACSubTypeULC}
}

func (e CACDocument) String() string { return string(e) }

func (e CACDocument) IsValid() bool {
	switch e {
	case CACDocumentCertificate, CACDocumentMERMAT, CACDocumentBoard, CACDocumentC02, CACDocumentC07, CACDocumentStatusReport, CACDocumentTINCert, CACDocumentTaxCert, CACDocumentAddressProof, CACDocumentIDCard, CACDocumentConstitution, CACDocumentPartnership, CACDocumentPassport:
		return true
	}
	return false
}

func CACDocumentValues() []CACDocument {
	return []CACDocument{CACDocumentCertificate, CACDocumentMERMAT, CACDocumentBoard, CACDocumentC02, CACDocumentC07, CACDocumentStatusReport, CACDocumentTINCert, CACDocumentTaxCert, CACDocumentAddressProof, CACDocumentIDCard, CACDocumentConstitution, CACDocumentPartnership, CACDocumentPassport}
}

func (e OnboardStage) String() string { return string(e) }

func (e OnboardStage) IsValid() bool {
	switch e {
	case OnboardStageBusiness, OnboardStageInterests, OnboardStageRegistration:
		return true
	}
	return false
}

func OnboardStageValues() []OnboardStage {
	return []OnboardStage{OnboardStageBusiness, OnboardStageInterests, OnboardStageRegistration}
}

func (e AIModel) String() string { return string(e) }

func (e AIModel) IsValid() bool {
	switch e {
	case AIModelOpenAI, AIModelAnthropic, AIModelGemini:
		return true
	}
	return false
}

func AIModelValues() []AIModel {
	return []AIModel{AIModelOpenAI, AIModelAnthropic, AIModelGemini}
}

func (e Status) String() string { return string(e) }

func (e Status) IsValid() bool {
	switch e {
	case StatusPending, StatusSubmitted, StatusCompleted, StatusApproved, StatusAccepted, StatusDeclined, StatusRevoked, StatusOnHold, StatusNew, StatusQueued, StatusSent, StatusFailed, StatusRead, StatusInactive, StatusSuccessful, StatusExpired, StatusDefault, StatusUnpaid, StatusPartial, StatusPaid, StatusCancelled, StatusReversed, StatusDraft, StatusOverdue:
		return true
	}
	return false
}

func StatusValues() []Status {
	return []Status{StatusPending, StatusSubmitted, StatusCompleted, StatusApproved, StatusAccepted, StatusDeclined, StatusRevoked, StatusOnHold, StatusNew, StatusQueued, StatusSent, StatusFailed, StatusRead, StatusInactive, StatusSuccessful, StatusExpired, StatusDefault, StatusUnpaid, StatusPartial, StatusPaid, StatusCancelled, StatusReversed, StatusDraft, StatusOverdue}
}
