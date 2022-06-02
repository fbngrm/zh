package zh

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type OutputFormat int

const (
	Format_JSON = iota
	Format_YAML
	Format_plain
)

type Formatter struct {
	Dict         LookupDict
	FilterFields []string
	Format       OutputFormat
}

func (f *Formatter) Print(index int) (string, error) {
	if index >= len(f.Dict) {
		return "", fmt.Errorf("could not find match at index %d", index)
	}

	var data interface{}
	if len(f.FilterFields) != 0 {
		var err error
		data, err = f.Dict[index].GetFields(f.FilterFields)
		if err != nil {
			return "", fmt.Errorf("could not filter fields: %w", err)
		}
	} else {
		data = f.Dict[index]
	}

	if f.Format == Format_JSON {
		return f.formatJSON(data)
	}
	if f.Format == Format_YAML {
		return f.formatYAML(data)
	}
	return f.format(data)
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

func (f *Formatter) format(data interface{}) (string, error) {
	var result string
	d, ok := data.(*Hanzi)
	if !ok {
		return "", fmt.Errorf("could not format; expected type %T but got %T", &Hanzi{}, data)
	}
	if d.Ideograph != "" {
		result += d.Ideograph
	}
	result += "\t"
	if r, ok := d.Readings["kMandarin"]; ok {
		result += r
	}
	result += "\t"
	result += d.Definition
	result += "\t"
	result += fmt.Sprintf("source=%s", d.Source)
	return result, nil
}
