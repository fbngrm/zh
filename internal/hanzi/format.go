package hanzi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Format int

const (
	Format_JSON = iota
	Format_YAML
	Format_text
)

type Formatter struct {
	format Format
}

func NewFormatter(fmt, fields string) *Formatter {
	return &Formatter{
		format: format(fmt),
	}
}

func (f *Formatter) FormatTemplate(h *Hanzi, fields, tmplPath string) (string, error) {
	i, err := f.filter(h, fields)
	if err != nil {
		return "", fmt.Errorf("could not filter fields: %w", err)
	}
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, []*Hanzi{i.(*Hanzi)})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (f *Formatter) Format(h *Hanzi, fields string) (string, error) {
	i, err := f.filter(h, fields)
	if err != nil {
		return "", fmt.Errorf("could not filter fields: %w", err)
	}
	if f.format == Format_JSON {
		return formatJSON(i)
	}
	if f.format == Format_YAML {
		return formatYAML(i)
	}
	return formatText(i)
}

func formatJSON(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func formatYAML(data interface{}) (string, error) {
	b, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func formatText(data interface{}) (string, error) {
	var result string
	d, ok := data.(*Hanzi)
	if !ok {
		return "", fmt.Errorf("could not format; expected type %T but got %T", &Hanzi{}, data)
	}
	if d.Ideograph != "" {
		result += d.Ideograph
	}
	result += "\t"
	result += strings.Join(d.Readings, ", ")
	result += "\t"
	result += strings.Join(d.Definitions, ", ")
	return result, nil
}

func format(fmt string) Format {
	if fmt == "json" {
		return Format_JSON
	} else if fmt == "yaml" {
		return Format_YAML
	}
	return Format_text
}

func (f *Formatter) filter(h *Hanzi, fields string) (interface{}, error) {
	var filterFields []string
	if len(fields) > 0 {
		filterFields = strings.Split(fields, ",")
	}
	for i, field := range filterFields {
		filterFields[i] = strings.TrimSpace(field)
	}

	if len(fields) != 0 {
		return h.GetFields(filterFields)
	}
	return h, nil
}
