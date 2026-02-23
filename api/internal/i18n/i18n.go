package i18n

import (
	"context"
	"embed"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var localeFS embed.FS

type contextKey string

const localizerKey contextKey = "localizer"

var bundle *i18n.Bundle

func Init(defaultLang string) error {
	tag, err := language.Parse(defaultLang)
	if err != nil {
		tag = language.BrazilianPortuguese
	}

	bundle = i18n.NewBundle(tag)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	files := []string{"locales/pt-BR.toml", "locales/en.toml"}
	for _, f := range files {
		data, err := localeFS.ReadFile(f)
		if err != nil {
			return err
		}
		bundle.MustParseMessageFileBytes(data, f)
	}

	return nil
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept-Language")
		localizer := i18n.NewLocalizer(bundle, accept)
		ctx := context.WithValue(r.Context(), localizerKey, localizer)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) *i18n.Localizer {
	l, ok := ctx.Value(localizerKey).(*i18n.Localizer)
	if !ok {
		return i18n.NewLocalizer(bundle, "pt-BR")
	}
	return l
}

func T(ctx context.Context, messageID string) string {
	localizer := FromContext(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		return messageID
	}
	return msg
}

func TWithData(ctx context.Context, messageID string, data map[string]interface{}) string {
	localizer := FromContext(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return msg
}
