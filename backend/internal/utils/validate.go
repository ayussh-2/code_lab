package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ValidationErrors(err error) map[string]string {
	out := map[string]string{}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		out["request"] = "invalid request payload"
		return out
	}

	for _, fe := range validationErrs {
		field := fe.Field()
		switch fe.Tag() {
		case "required":
			out[field] = fmt.Sprintf("%s is required", field)
		case "email":
			out[field] = fmt.Sprintf("%s must be a valid email", field)
		case "min":
			out[field] = fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		case "max":
			out[field] = fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
		default:
			out[field] = fmt.Sprintf("%s is invalid", field)
		}
	}

	return out
}