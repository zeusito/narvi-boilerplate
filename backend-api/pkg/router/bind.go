package router

import (
	"backend-api/pkg/terrors"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-playground/validator/v10"
)

const (
	MIMEApplicationJSON = "application/json"
)

// use a single instance of Validate, it caches struct info
var validate = validator.New(validator.WithRequiredStructEnabled())

// BindBody binds the request body to the target struct.
func BindBody[T any](r *http.Request, target *T) error {
	if r.ContentLength == 0 {
		return terrors.PreconditionFailed("empty request body")
	}

	// Extract the media type from the request.
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return terrors.PreconditionFailed("invalid content type")
	}

	// We only support JSON for now.
	if mediaType != MIMEApplicationJSON {
		return terrors.PreconditionFailed("unsupported content type")
	}

	// Decode the request body into the provided struct.
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return terrors.PreconditionFailed("invalid json payload")
	}

	// Validate the struct using the validator.
	err = validate.Struct(target)
	if err != nil {
		var ve validator.ValidationErrors
		if ok := errors.As(err, &ve); ok {
			log.Info().Msgf("validation error: %d", len(ve))
			// Get the first error message.
			formattedErr := fmt.Sprintf("%s %s", ve[0].Field(), msgForTag(ve[0].Tag()))
			return terrors.PreconditionFailed(formattedErr)
		}

		return terrors.PreconditionFailed("validation failed")
	}

	return nil
}

func Validate[T any](target *T) error {
	err := validate.Struct(target)
	if err != nil {
		var ve validator.ValidationErrors
		if ok := errors.As(err, &ve); ok {
			log.Info().Msgf("validation error: %d", len(ve))
			// Get the first error message.
			formattedErr := fmt.Sprintf("%s %s", ve[0].Field(), msgForTag(ve[0].Tag()))
			return terrors.PreconditionFailed(formattedErr)
		}

		return terrors.PreconditionFailed("validation failed")
	}

	return nil
}

func msgForTag(tag string) string {
	switch tag {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "len":
		return "length must be exactly %s characters"
	case "min":
		return "length must be at least %s characters"
	case "max":
		return "length must be at most %s characters"
	case "gt":
		return "must be greater than %s"
	case "gte":
		return "must be greater than or equal to %s"
	case "lt":
		return "must be less than %s"
	case "lte":
		return "must be less than or equal to %s"
	default:
		return tag
	}
}
