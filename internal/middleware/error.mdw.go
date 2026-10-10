package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

// ErrorHandler is the single error translator.
// Controllers attach errors with ctx.Error(err) and return; this middleware
// maps them to a consistent JSON response after the handler chain.
func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		err := ctx.Errors.Last().Err

		statusCode := http.StatusInternalServerError
		message := "Internal Server Error"
		errorsList := []any{}
		data := any(nil)

		var customError *utils.ErrorResponse
		if errors.As(err, &customError) {
			statusCode = customError.Status
			message = customError.Message
			errorsList = utils.ErrorsAsAny(customError.Errors)
			data = customError.Data

		} else if errors.Is(err, pgx.ErrNoRows) {
			statusCode = http.StatusNotFound
			message = "Resource not found"

		} else {

			// handle postgres errors
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505": // unique_violation
					statusCode = http.StatusBadRequest
					message = "Duplicate field value entered"
				case "23503": // foreign_key_violation
					statusCode = http.StatusBadRequest
					message = "Related resource not found"
				case "23502": // not_null_violation
					statusCode = http.StatusBadRequest
					message = "Missing required field"
				case "23514": // check_violation
					statusCode = http.StatusBadRequest
					message = "Invalid field value"
				case "22P02": // invalid_text_representation
					statusCode = http.StatusNotFound
					message = "Resource not found"
				default:
					statusCode = http.StatusBadRequest
					message = "Database error"
				}
			}

		}

		if data == nil {
			data = map[string]any{}
		}

		ctx.JSON(statusCode, utils.Result[any]{
			Error:   true,
			Message: message,
			Code:    statusCode,
			Data:    data,
			Errors:  errorsList,
			Status:  statusCode,
		})
	}
}
