package notification

import (
	"bytes"
	"html/template"
)

const sightingEmailTpl = `
<div style="margin: 5px auto; text-align:center; padding: 10px; font-family: 'Raleway', sans-serif;">
    <h1>Traceo</h1>
    <p>Alguém informou que a pessoa desaparecida foi avistada.</p>
    <h3>Observação</h3>
    <p style="background-color: #0097D6; padding: 5px; border-radius: 5px; color: white; font-weight: bold">
        {{.Observation}}
    </p>
    <p>Acesse a plataforma para ver a localização no mapa.</p>
</div>
`

func renderTemplate(tplStr string, data interface{}) (string, error) {
	tpl, err := template.New("email").Parse(tplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
