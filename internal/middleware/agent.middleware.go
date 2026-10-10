package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
)

const ContextUserAgent = "userAgent"

type UserAgent struct {
	Source    string `json:"source"`
	Browser   string `json:"browser"`
	Version   string `json:"version"`
	OS        string `json:"os"`
	Platform  string `json:"platform"`
	IsMobile  bool   `json:"isMobile"`
	IsTablet  bool   `json:"isTablet"`
	IsDesktop bool   `json:"isDesktop"`
	IsBot     bool   `json:"isBot"`
}

// Agent returns Gin middleware that parses the incoming User-Agent request
// header and stores a structured UserAgent value on the Gin context before
// downstream handlers run.
//
// Register Agent early in the middleware chain—typically after request
// identification and security headers but before handlers that need browser,
// OS, or device-type metadata for logging, analytics, feature gating, or
// response shaping. The middleware only inspects the request; it does not
// modify the response body or status and never aborts the chain.
//
// Context and request interaction:
//   - Reads ctx.GetHeader("User-Agent") from the HTTP request.
//   - Writes ctx.Set(ContextUserAgent, UserAgent) using the ContextUserAgent key.
//
// An empty or absent User-Agent header is still passed to ParseUserAgent, which
// yields a value with empty or default fields rather than failing. Handlers
// should read the parsed result via GetUserAgent instead of re-parsing the raw
// header. The middleware always calls ctx.Next().
func Agent() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(ContextUserAgent, ParseUserAgent(ctx.GetHeader("User-Agent")))
		ctx.Next()
	}
}

// GetUserAgent returns the UserAgent value previously stored on the Gin
// context by the Agent middleware.
//
// Use this helper in handlers or downstream middleware when device, browser,
// or bot classification is needed. It performs a type-safe lookup under the
// ContextUserAgent key and avoids duplicate parsing of the User-Agent header.
//
// When Agent has not run, the context key is missing, or the stored value is
// not a UserAgent, GetUserAgent returns a zero UserAgent (empty strings and
// false flags). It does not read from the request directly and never panics.
func GetUserAgent(ctx *gin.Context) UserAgent {
	value, exists := ctx.Get(ContextUserAgent)
	if !exists {
		return UserAgent{}
	}

	agent, ok := value.(UserAgent)
	if !ok {
		return UserAgent{}
	}

	return agent
}

// ParseUserAgent converts a raw User-Agent header string into a structured
// UserAgent using github.com/mssola/user_agent.
//
// The result includes Source (full UA string), Browser, Version, OS, Platform,
// and boolean flags IsMobile, IsTablet, IsDesktop, and IsBot. Tablet detection
// augments the library's Mobile() result: platform equal to "iPad"
// (case-insensitive) or the substring "tablet" in the source marks the client
// as a tablet. IsMobile is true only when the parser reports mobile and the
// device is not a tablet. IsDesktop is true when the client is not mobile, not
// tablet, and not classified as a bot.
//
// This function is pure—it does not read or write Gin context or HTTP
// responses. Agent middleware invokes it on each request; it may also be used
// in tests or when parsing User-Agent strings outside the middleware pipeline.
func ParseUserAgent(source string) UserAgent {
	parsed := user_agent.New(source)
	browser, version := parsed.Browser()
	platform := parsed.Platform()
	isMobile := parsed.Mobile()
	isBot := parsed.Bot()
	isTablet := strings.EqualFold(platform, "iPad") ||
		strings.Contains(strings.ToLower(source), "tablet")

	return UserAgent{
		Source:    parsed.UA(),
		Browser:   browser,
		Version:   version,
		OS:        parsed.OS(),
		Platform:  platform,
		IsMobile:  isMobile && !isTablet,
		IsTablet:  isTablet,
		IsDesktop: !isMobile && !isTablet && !isBot,
		IsBot:     isBot,
	}
}
