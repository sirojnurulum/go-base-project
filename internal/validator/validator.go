package validator

import (
	"fmt"
	"net/http"

	goValidator "github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator adalah wrapper untuk library go-playground/validator.
type CustomValidator struct {
	validator *goValidator.Validate
}

// NewCustomValidator membuat instance baru dari CustomValidator.
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: goValidator.New()}
}

// Validate memvalidasi struct yang diberikan dan mengembalikan error yang ramah HTTP.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Ubah error validasi menjadi format yang lebih mudah dibaca.
		validationErrors := err.(goValidator.ValidationErrors)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Validation failed for field '%s' with tag '%s'", validationErrors[0].Field(), validationErrors[0].Tag()))
	}
	return nil
}
