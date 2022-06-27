package hanzi

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Format int

const (
	Format_JSON = iota
	Format_YAML
	Format_plain
)

type Formatter struct {
	format       Format
	filterFields []string
}

func NewFormatter(f Format, fields string) *Formatter {
	return &Formatter{
		format:       f,
		filterFields: prepareFilter(fields),
	}
}

func (f *Formatter) Format(hanzi *Hanzi) (string, error) {
	var data interface{}
	if len(f.filterFields) != 0 {
		var err error
		data, err = hanzi.GetFields(f.filterFields)
		if err != nil {
			return "", fmt.Errorf("could not filter fields: %w", err)
		}
	} else {
		data = hanzi
	}
	if f.format == Format_JSON {
		return f.formatJSON(data)
	}
	if f.format == Format_YAML {
		return f.formatYAML(data)
	}
	return f.formatPlain(data)
}

func (f *Formatter) formatJSON(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (f *Formatter) formatYAML(data interface{}) (string, error) {
	b, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (f *Formatter) formatPlain(data interface{}) (string, error) {
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
	result += "\t"
	result += fmt.Sprintf("source=%s", d.Source)
	return result, nil
}

func prepareFilter(fields string) []string {
	var filterFields []string
	if len(fields) > 0 {
		filterFields = strings.Split(fields, ",")
	}
	for i, field := range filterFields {
		filterFields[i] = strings.TrimSpace(field)
	}
	return filterFields
}
