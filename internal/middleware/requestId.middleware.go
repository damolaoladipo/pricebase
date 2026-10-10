package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header read and written for request tracking.
	RequestIDHeader = "X-Request-ID"
	// ContextRequestID is the context key holding the request ID.
	// The logger picks it up as meta.requestId.
	ContextRequestID = "requestId"
)

// RequestID assigns an ID to every request, like gin-contrib/requestid.
// An incoming X-Request-ID is reused (proxy-safe); otherwise a UUID is
// generated. The ID is set as the response header and stored on the context.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		c.Header(RequestIDHeader, id)
		c.Set(ContextRequestID, id)
		c.Next()
	}
}
