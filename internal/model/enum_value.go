package model

import (
	"fmt"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
)

// EnumValue represents a single enum entry
type EnumValue struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// NewEnumValue creates a new enum value
func NewEnumValue(typeName string, name string, description string, rawValue any, opts *config.Options) *EnumValue {
	rawStr := fmt.Sprintf("%v", rawValue)
	var value string
	if typeName == "string" {
		value = "'" + gen.EscapeJS(rawStr) + "'"
	} else {
		value = rawStr
	}
	ev := &EnumValue{
		Type:  typeName,
		Value: value,
	}

	if name != "" {
		ev.Name = name
	} else {
		ev.Name = gen.EnumName(fmt.Sprintf("%v", rawValue), opts)
	}

	if description != "" {
		ev.Description = description
	} else {
		ev.Description = ev.Name
	}

	if ev.Name == "" {
		ev.Name = "_"
	}
	if ev.Description == "" {
		ev.Description = ev.Name
	}

	return ev
}
