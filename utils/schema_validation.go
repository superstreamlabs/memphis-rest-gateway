package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var Validator *validator.Validate

type ValidationError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func InitializeValidations() {
	Validator = validator.New()
}

func Validate(schema any) []ValidationError {
	errors := []ValidationError{}
	err := Validator.Struct(schema)
	if err != nil {
		for _, f := range err.(validator.ValidationErrors) {
			tErr := f.ActualTag()
			if f.Param() != "" {
				tErr = fmt.Sprintf("%s=%s", err, f.Param())
			}
			errors = append(errors, ValidationError{Field: f.Field(), Reason: tErr})
		}
		return errors
	}
	return nil
}
