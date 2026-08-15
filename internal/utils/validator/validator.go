package validator

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	Validate   *validator.Validate
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[0-9\s\-\(\)]{10,20}$`)
)

type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Tag     string      `json:"tag"`
	Value   interface{} `json:"value,omitempty"`
	Param   string      `json:"param,omitempty"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	var messages []string
	for _, e := range ve {
		messages = append(messages, e.Message)
	}
	return strings.Join(messages, "; ")
}

func Init() {
	Validate = validator.New()

	// Register custom validators
	registerCustomValidators()

	// Register custom tag name function to use JSON tags
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func registerCustomValidators() {
	_ = Validate.RegisterValidation("uuid", validateUUID)
	_ = Validate.RegisterValidation("phone", validatePhone)
	_ = Validate.RegisterValidation("password_strength", validatePasswordStrength)
	_ = Validate.RegisterValidation("no_spaces", validateNoSpaces)
	_ = Validate.RegisterValidation("not_empty", validateNotEmpty)
	_ = Validate.RegisterValidation("valid_name", validateName)
	_ = Validate.RegisterValidation("future_date", validateFutureDate)
	_ = Validate.RegisterValidation("past_date", validatePastDate)
	_ = Validate.RegisterValidation("valid_url", validateURL)
	_ = Validate.RegisterValidation("alphanumeric", validateAlphanumeric)
}

func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Let required validator handle empty
	}
	_, err := uuid.Parse(value)
	return err == nil
}

func validatePhone(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	return phoneRegex.MatchString(value)
}

func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return true
	}

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateNoSpaces(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return !strings.ContainsAny(value, " \t\r\n")
}

func validateNotEmpty(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return strings.TrimSpace(value) != ""
}

func validateName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if name == "" {
		return true
	}

	// Name should only contain letters, spaces, hyphens, and apostrophes
	for _, char := range name {
		if !unicode.IsLetter(char) && char != ' ' && char != '-' && char != '\'' {
			return false
		}
	}
	return true
}

func validateFutureDate(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return date.After(time.Now())
}

func validatePastDate(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return date.Before(time.Now())
}

func validateURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true
	}
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

func validateAlphanumeric(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			return false
		}
	}
	return true
}

// Struct validates a struct and returns validation errors
func Struct(s interface{}) ValidationErrors {
	err := Validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors ValidationErrors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getValidationMessage(e),
				Tag:     e.Tag(),
				Value:   e.Value(),
				Param:   e.Param(),
			})
		}
	}
	return errors
}

// StructCtx validates a struct with context
func StructCtx(ctx context.Context, s interface{}) ValidationErrors {
	err := Validate.StructCtx(ctx, s)
	if err == nil {
		return nil
	}

	var errors ValidationErrors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getValidationMessage(e),
				Tag:     e.Tag(),
				Value:   e.Value(),
				Param:   e.Param(),
			})
		}
	}
	return errors
}

func getValidationMessage(e validator.FieldError) string {
	field := e.Field()
	param := e.Param()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "required_if":
		return fmt.Sprintf("%s is required when %s", field, param)
	case "required_unless":
		return fmt.Sprintf("%s is required unless %s", field, param)
	case "required_with":
		return fmt.Sprintf("%s is required when %s is present", field, param)
	case "required_without":
		return fmt.Sprintf("%s is required when %s is not present", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "numeric":
		return fmt.Sprintf("%s must contain only numbers", field)
	case "alphanumeric":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", field, param)
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "nefield":
		return fmt.Sprintf("%s must not be equal to %s", field, param)
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", field, param)
	case "jwt":
		return fmt.Sprintf("%s must be a valid JWT token", field)
	case "password_strength":
		return fmt.Sprintf("%s must be at least 8 characters and contain uppercase, lowercase, number, and special character", field)
	case "no_spaces":
		return fmt.Sprintf("%s must not contain spaces", field)
	case "valid_name":
		return fmt.Sprintf("%s can only contain letters, spaces, hyphens, and apostrophes", field)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "future_date":
		return fmt.Sprintf("%s must be a future date", field)
	case "past_date":
		return fmt.Sprintf("%s must be a past date", field)
	default:
		return fmt.Sprintf("%s failed validation for %s", field, e.Tag())
	}
}

// ValidateEmail checks if email is valid
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidatePhone checks if phone number is valid
func ValidatePhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}
