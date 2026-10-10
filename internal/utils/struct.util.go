package utils

import (
	"time"
)

type APIKey struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"userId" db:"user_id"`
	Secret      string    `json:"secret" db:"secret"`
	Public      string    `json:"public" db:"public"`
	SecretToken string    `json:"secretToken" db:"secret_token"`
	PublicToken string    `json:"publicToken" db:"public_token"`
	Domain      string    `json:"domain" db:"domain"`
	IsActive    bool      `json:"isActive" db:"is_active"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type CountryTimezone struct {
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	DisplayName  string   `json:"displayName"`
	Countries    []string `json:"countries"`
	UTCOffset    int      `json:"utcOffset"`
	UTCOffsetStr string   `json:"utcOffsetStr"`
	DSTOffset    int      `json:"dstOffset"`
	DSTOffsetStr string   `json:"dstOffsetStr"`
	AliasOf      any      `json:"aliasOf"`
}

type Country struct {
	Name         string            `json:"name"`
	Code2        string            `json:"code2"`
	Code3        string            `json:"code3"`
	Capital      string            `json:"capital"`
	Region       string            `json:"region"`
	CurrencyCode string            `json:"currencyCode"`
	PhoneCode    string            `json:"phoneCode"`
	Timezones    []CountryTimezone `json:"timezones"`
}

// PermissionAction is the shared {name, description} action shape stored in
// permissions.actions JSONB and seeded from permissions.json.
// It mirrors IPermissionAction: a shared shape, not a model.
type PermissionAction struct {
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

type UserPermission struct {
	Group   string   `json:"group"`
	Actions []string `json:"actions"`
}

type UserLogin struct {
	Last   string `json:"last" db:"last"`
	Method string `json:"method" db:"method"`
}

type UserOnboard struct {
	Stage  string `json:"stage" db:"stage"`
	Status string `json:"status" db:"status"`
}

type UserRegister struct {
	Date   time.Time `json:"date" db:"date"`
	Method string    `json:"method" db:"method"`
}

type UserStatus struct {
	Profile       string `json:"profile" db:"profile"`
	ExistOnInvite bool   `json:"existOnInvite" db:"exist_on_invite"`
}

type OAuthCredentials struct {
	AccessToken  string     `json:"accessToken" db:"access_token"`
	RefreshToken string     `json:"refreshToken" db:"refresh_token"`
	TokenType    string     `json:"tokenType" db:"token_type"`
	IDToken      string     `json:"idToken" db:"id_token"`
	ExpiryDate   *time.Time `json:"expiryDate" db:"expiry_date"`
	Scope        string     `json:"scope" db:"scope"`

	// If Data is structured, give it a proper struct later.
	Data map[string]any `json:"data" db:"-"`
}

type OAuthProvider struct {
	ID     string           `json:"id" db:"id"`
	UserID string           `json:"userId" db:"user_id"`
	Brand  string           `json:"brand" db:"brand"`
	Creds  OAuthCredentials `json:"creds" db:"-"`
}

type LoginSettings struct {
	Email     bool `json:"email"`
	SMS       bool `json:"sms"`
	Phone     bool `json:"phone"`
	TwoFA     bool `json:"twofa"`
	Biometric bool `json:"biometric"`
}

type UsageLimit struct {
	Enabled bool `json:"enabled"`
	Value   int  `json:"value"`
}

type UsageSettings struct {
	Branches   UsageLimit `json:"branches"`
	Categories UsageLimit `json:"categories"`
	Customers  UsageLimit `json:"customers"`
	Sales      UsageLimit `json:"sales"`
	Invoices   UsageLimit `json:"invoices"`
	Team       UsageLimit `json:"team"`
}

type NotificationChannels struct {
	Email    bool `json:"email"`
	WhatsApp bool `json:"whatsapp"`
}

type InvoiceReminder struct {
	Enabled       bool                 `json:"enabled"`
	DaysBeforeDue int                  `json:"daysBeforeDue"`
	CadenceDays   int                  `json:"cadenceDays"`
	Channels      NotificationChannels `json:"channels"`
}

type BillingSettings struct {
	InvoiceReminder InvoiceReminder `json:"invoiceReminder"`
}

type LogType string

const (
	LogInfo    LogType = "info"
	LogWarning LogType = "warning"
	LogError   LogType = "error"
	LogSuccess LogType = "success"
	LogAny     LogType = "any"
)

// APIReport is an optional report payload attached to Result responses.
type APIReport struct {
	Total   int `json:"total,omitempty"`
	Success int `json:"success,omitempty"`
	Failed  int `json:"failed,omitempty"`
}

// Pagination is an optional paging payload attached to Result responses.
type Pagination struct {
	Page    int   `json:"page"`
	Limit   int   `json:"limit"`
	Total   int64 `json:"total"`
	Pages   int   `json:"pages,omitempty"`
	HasNext bool  `json:"hasNext,omitempty"`
	HasPrev bool  `json:"hasPrev,omitempty"`
}

// Log DTOs live in utils (the leaf package) so that the Logger can use them
// without importing the higher-level dtos package (which would create an
// import cycle: utils -> dtos -> models -> utils).
// The dtos package re-exports these as type aliases for backward compatibility.

type LogRequestDTO struct {
	ClassName string `json:"className,omitempty"`
	Type      string `json:"type,omitempty"`
	Payload   any    `json:"payload,omitempty"`
	Monitor   bool   `json:"monitor,omitempty"`
	Display   *bool  `json:"display,omitempty"`
}

type LogDirectDTO struct {
	ClassName string `json:"className,omitempty"`
	Payload   any    `json:"payload,omitempty"`
	Monitor   bool   `json:"monitor,omitempty"`
	Display   *bool  `json:"display,omitempty"`
}

type LogAuditRequestDTO struct {
	Name     string `json:"name,omitempty"`
	Path     string `json:"path"`
	Headers  any    `json:"headers"`
	IP       string `json:"ip"`
	Device   any    `json:"device"`
	Body     any    `json:"body"`
	Metadata any    `json:"metadata,omitempty"`
}

type LogAPICallDTO struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Payload  any    `json:"payload"`
	IsError  bool   `json:"isError,omitempty"`
}

type LogAuditDTO struct {
	Name string `json:"name,omitempty"`

	User struct {
		ID    any    `json:"id,omitempty"`
		Email string `json:"email,omitempty"`
	} `json:"user"`

	Action struct {
		Function    string `json:"function"`
		Description string `json:"description,omitempty"`
		Status      string `json:"status"`
	} `json:"action"`

	Entity struct {
		Model string `json:"model"`
	} `json:"entity"`

	Request struct {
		Body     any    `json:"body"`
		Path     string `json:"path,omitempty"`
		Metadata any    `json:"metadata,omitempty"`
	} `json:"request"`
}

// QueryParam represents a single query filter parameter.
type QueryParam struct {
	Key      string
	Value    any
	Operator string // eq, gt, gte, lt, lte, in, neq
}

// QueryParams is a collection of query parameters.
type QueryParams []QueryParam

// AdvancedQuery holds all parameters for an advanced query.
//
// Like the TS advanced({ model, populate, sortRef, req, q, paginate, fields }),
// it carries no DB handle — the pool is passed explicitly as its own argument
// to Advanced/Populate/Search/GeoSearch.
type AdvancedQuery struct {
	Table      string
	Columns    []string
	Fields     []string // forced select fields
	Ref        string   // column to filter on (e.g. "organization_id")
	Value      any      // value for ref filter
	Data       QueryParams
	Operator   string              // "and", "or", "in", "andor", "orand"
	Query      QueryParams         // extra query filters (TS `q.query`)
	QueryParam map[string][]string // raw query params (from HTTP request)
	Paginate   string              // "", "absolute", "relative"
}

// PopulateQuery holds parameters for a populate (join) query.
type PopulateQuery struct {
	Table       string
	Columns     []string
	Fields      []string
	Ref         string
	Value       any
	Path        string
	PathColumns []string
	Populate    []PopulateQuery
	QueryParam  map[string][]string
	Paginate    string
}

// SearchQuery holds parameters for a search query.
type SearchQuery struct {
	Table      string
	Columns    []string
	Fields     []string
	Ref        string
	Value      any
	Data       QueryParams
	Query      QueryParams
	Operator   string
	QueryParam map[string][]string
	Paginate   string
}

// GeoSearchQuery holds parameters for a geo-spatial search.
type GeoSearchQuery struct {
	Table      string
	Columns    []string
	Ref        string
	Value      any
	Data       QueryParams
	Operator   string // center, centersphere, range, within-poly
	GeoData    GeoData
	QueryParam map[string][]string
	Paginate   string
}

// GeoData holds geo-spatial search parameters.
type GeoData struct {
	Longitude   float64
	Latitude    float64
	Radius      float64
	RadiusUnit  string // "km" or "ml"
	MinDistance float64
	MaxDistance float64
	Coordinates [][]float64 // polygon coordinates
}

// PageRef holds a page reference.
type PageRef struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// PaginationResult holds pagination metadata.
type PaginationResult struct {
	Next *PageRef `json:"next,omitempty"`
	Prev *PageRef `json:"prev,omitempty"`
}

// PaginatedResult is the standard paginated response.
type PaginatedResult struct {
	Total      int64             `json:"total"`
	Count      int               `json:"count"`
	Pagination *PaginationResult `json:"pagination,omitempty"`
	Data       []map[string]any  `json:"data"`
}

// DateRange holds a formatted date range.
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ResultsMiddleware configures the middleware.
// Mirrors TS IResultMiddleware { model, populate, key, sortRef, cache, query, fields }.
type ResultsMiddleware struct {
	// Table is the TS `model`: the SQL table to query.
	Table    string
	Columns  []string
	Fields   []string
	SortRef  string
	Cache    bool
	CacheKey string
	// Data maps to TS `query`/`q`: extra filter conditions applied with Operator.
	Data QueryParams
	// Query maps to TS `q.query` used by andor/orand operators.
	Query    QueryParams
	Operator string
	Paginate string
}

type DeviceOS struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Platform  string `json:"platform"`
	Version   string `json:"version"`
}

type DeviceClient struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Version   string `json:"version"`
}

type DeviceDetails struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Brand string `json:"brand"`
	Model string `json:"model"`
	Code  string `json:"code"`
}
