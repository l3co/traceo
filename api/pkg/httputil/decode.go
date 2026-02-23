package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

var (
	validate  = validator.New()
	sanitizer = bluemonday.StrictPolicy()
)

func DecodeAndValidate(r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if err := validate.Struct(dst); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return fmt.Errorf("validation: %s", formatValidationErrors(ve))
		}
		return fmt.Errorf("validation: %w", err)
	}

	return nil
}

func SanitizeString(s string) string {
	return sanitizer.Sanitize(strings.TrimSpace(s))
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, fmt.Sprintf("%s: failed on '%s'", e.Field(), e.Tag()))
	}
	return strings.Join(msgs, "; ")
}
