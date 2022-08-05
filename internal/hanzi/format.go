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
	Format_TXT = iota
	Format_verbose_TXT
	Format_JSON
	Format_YAML
)

type Formatter struct {
	format   Format
	fields   []string
	template string
}

func (f *Formatter) WithFormat(format string) *Formatter {
	if format == "json" {
		f.format = Format_JSON
	} else if format == "yaml" {
		f.format = Format_YAML
	} else if format == "verbose" {
		f.format = Format_verbose_TXT
	} else {
		f.format = Format_TXT
	}
	return f
}

func (f *Formatter) WithFields(fields string) *Formatter {
	f.fields = strings.Split(fields, ",")
	return f
}

func (f *Formatter) WithTemplate(tmpl string) *Formatter {
	f.template = tmpl
	return f
}

func (f *Formatter) Format(h []*Hanzi) (string, error) {
	if h == nil {
		return "no results :(", nil
	}

	if f.template != "" {
		return f.FormatTemplate(h)
	}

	var i interface{}
	i = h
	if len(f.fields) > 0 {
		var err error
		i, err = f.filter(h)
		if err != nil {
			return "", fmt.Errorf("could not filter fields: %w", err)
		}
	}

	if f.format == Format_JSON {
		return formatJSON(i)
	}
	if f.format == Format_YAML {
		return formatYAML(i)
	}
	if f.format == Format_verbose_TXT {
		return formatTextVerbose(h)
	}
	return formatText(h)
}

func (f *Formatter) FormatTemplate(h []*Hanzi) (string, error) {
	// tplFuncMap := make(template.FuncMap)
	// tplFuncMap["definitions"] = func(definitions []string) string {
	// 	defs := ""
	// 	if len(definitions) == 0 {
	// 		return ""
	// 	}
	// 	if len(definitions) == 1 {
	// 		definitions = strings.Split(definitions[0], ",")
	// 	}
	// 	for i, s := range definitions {
	// 		defs += s
	// 		if i == 4 {
	// 			break
	// 		}
	// 		if i == len(definitions)-1 {
	// 			break
	// 		}
	// 		defs += ", "
	// 		defs += "\n"
	// 	}
	// 	return defs
	// }
	// tmpl, err := template.New("anki.tmpl").Funcs(tplFuncMap).ParseFiles(tmplPath)

	// tmplName needs to be name of argument to ParseFiles
	tmplName := "anki.tmpl"
	tmpl, err := template.New(tmplName).ParseFiles(f.template)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, h)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
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

func formatText(hs []*Hanzi) (string, error) {
	var result string
	for _, h := range hs {
		if h.Ideograph != "" {
			result += h.Ideograph
		}
		result += "\t"
		result += strings.Join(h.Readings, ", ")
		result += "\t"
		result += strings.Join(h.Definitions, ", ")
		result += "\n"
	}
	return result, nil
}

func formatTextVerbose(hs []*Hanzi) (string, error) {
	var result string
	for _, h := range hs {
		if h.Ideograph != "" {
			result += h.Ideograph
		}
		result += "\t"
		result += strings.Join(h.Readings, ", ")
		result += "\t"
		result += strings.Join(h.Definitions, ", ")
		result += "\n"
		for _, c := range h.ComponentsDecompositions {
			if c.Ideograph != "" {
				result += c.Ideograph
			}
			result += "\t"
			result += strings.Join(c.Readings, ", ")
			result += "\t"
			result += strings.Join(c.Definitions, ", ")
			result += "\n"
		}
	}
	return result, nil
}

func (f *Formatter) filter(hs []*Hanzi) ([]map[string]interface{}, error) {
	var filtered []map[string]interface{}
	for _, h := range hs {
		var filterFields []string
		for i, field := range filterFields {
			filterFields[i] = strings.TrimSpace(field)
		}

		if len(f.fields) != 0 {
			fields, err := h.GetFields(filterFields)
			if err != nil {
				return nil, err
			}
			filtered = append(filtered, fields)
		}
	}
	return filtered, nil
}
