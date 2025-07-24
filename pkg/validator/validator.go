package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) []ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, ValidationError{
			Field:   err.Field(),
			Message: getErrorMessage(err),
		})
	}

	return validationErrors
}

// getErrorMessage returns a human-readable error message
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s characters", err.Field(), err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", err.Field(), err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", err.Field(), err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", err.Field(), err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
	case "gtfield":
		return fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}
