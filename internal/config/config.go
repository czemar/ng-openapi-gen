// Package config handles loading and validating ng-openapi-gen configuration
// from JSON files and CLI arguments.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Options mirrors the TypeScript Options interface
type Options struct {
	Input                    string                         `json:"input"`
	Output                   string                         `json:"output,omitempty"`
	FetchTimeout             int                            `json:"fetchTimeout,omitempty"`
	RemoveStaleFiles         *bool                          `json:"removeStaleFiles,omitempty"`
	UseTempDir               bool                           `json:"useTempDir,omitempty"`
	Silent                   bool                           `json:"silent,omitempty"`
	DefaultTag               string                         `json:"defaultTag,omitempty"`
	IncludeTags              []string                       `json:"includeTags,omitempty"`
	ExcludeTags              []string                       `json:"excludeTags,omitempty"`
	ExcludePaths             []string                       `json:"excludePaths,omitempty"`
	Configuration            string                         `json:"configuration,omitempty"`
	ApiService               any                            `json:"apiService,omitempty"` // string or bool
	ExcludeParameters        []string                       `json:"excludeParameters,omitempty"`
	FunctionIndex            any                            `json:"functionIndex,omitempty"` // string or bool
	CamelizeModelNames       *bool                          `json:"camelizeModelNames,omitempty"`
	IgnoreUnusedModels       *bool                          `json:"ignoreUnusedModels,omitempty"`
	ModelIndex               any                            `json:"modelIndex,omitempty"` // string or bool
	ModelPrefix              string                         `json:"modelPrefix,omitempty"`
	ModelSuffix              string                         `json:"modelSuffix,omitempty"`
	EnumStyle                string                         `json:"enumStyle,omitempty"`
	EnumArray                *bool                          `json:"enumArray,omitempty"`
	Promises                 *bool                          `json:"promises,omitempty"`
	Services                 *bool                          `json:"services,omitempty"`
	ServiceIndex             any                            `json:"serviceIndex,omitempty"` // string or bool
	ServicePrefix            string                         `json:"servicePrefix,omitempty"`
	ServiceSuffix            string                         `json:"serviceSuffix,omitempty"`
	BaseService              string                         `json:"baseService,omitempty"`
	RequestBuilder           string                         `json:"requestBuilder,omitempty"`
	Response                 string                         `json:"response,omitempty"`
	Module                   any                            `json:"module,omitempty"` // string or bool
	EndOfLineStyle           string                         `json:"endOfLineStyle,omitempty"`
	Templates                string                         `json:"templates,omitempty"`
	SkipJsonSuffix           bool                           `json:"skipJsonSuffix,omitempty"`
	CustomizedResponseType   map[string]CustomizedResponse  `json:"customizedResponseType,omitempty"`
	KeepFullResponseMediaType any                           `json:"keepFullResponseMediaType,omitempty"` // bool or []KeepMediaType
	IndexFile                bool                           `json:"indexFile,omitempty"`
	ApiModule                bool                           `json:"apiModule,omitempty"`
}

type CustomizedResponse struct {
	ToUse string `json:"toUse"`
}

type KeepMediaType struct {
	MediaType string `json:"mediaType,omitempty"`
	Use       string `json:"use"`
}

// LoadOptions reads and parses options from a JSON config file
func LoadOptions(configPath string) (*Options, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var opts Options
	if err := json.Unmarshal(data, &opts); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	opts.setDefaults()
	return &opts, nil
}

func (o *Options) setDefaults() {
	if o.Output == "" {
		o.Output = "src/app/api"
	}
	if o.FetchTimeout <= 0 {
		o.FetchTimeout = 20000
	}
	if o.RemoveStaleFiles == nil {
		t := true
		o.RemoveStaleFiles = &t
	}
	if o.DefaultTag == "" {
		o.DefaultTag = "Api"
	}
	if o.EnumStyle == "" {
		o.EnumStyle = "alias"
	}
	if o.EnumArray == nil && o.EnumStyle == "alias" {
		t := true
		o.EnumArray = &t
	}
	if o.ApiService == nil {
		o.ApiService = "Api"
	}
	if o.Module == nil {
		o.Module = false
	} else if b, ok := o.Module.(bool); ok && b {
		o.Module = "ApiModule"
	}
	if o.EndOfLineStyle == "" {
		o.EndOfLineStyle = "auto"
	}
	if o.CamelizeModelNames == nil {
		t := true
		o.CamelizeModelNames = &t
	}
}

// GetStringOrBool returns the string value or default if bool
func GetStringOrBool(v any, defaultStr string) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		if val == "" {
			return defaultStr
		}
		return val
	case bool:
		if val {
			return defaultStr
		}
		return ""
	default:
		return ""
	}
}

// GetBool returns the bool value from any
func GetBool(v any) bool {
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// ParseOptionsFromArgs parses options from CLI args and JSON config
func ParseOptionsFromArgs(args []string) (*Options, error) {
	var configPath string
	var input string

	// Simple arg parsing - look for --config and --input
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--config", "-c":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--input", "-i":
			if i+1 < len(args) {
				input = args[i+1]
				i++
			}
		}
	}

	if configPath == "" {
		configPath = "ng-openapi-gen.json"
	}

	var opts *Options
	var err error

	if fileExists(configPath) {
		opts, err = LoadOptions(configPath)
		if err != nil {
			return nil, err
		}
	} else {
		opts = &Options{}
		opts.setDefaults()
	}

	if input != "" {
		opts.Input = input
	}

	// Process remaining args as overrides
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") && arg != "--config" && arg != "--input" {
			key := strings.TrimPrefix(arg, "--")
			key = kebabToCamel(key)
			value := ""
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				value = args[i+1]
				i++
			}
			applyOverride(opts, key, value)
		}
	}

	if opts.Input == "" {
		return nil, fmt.Errorf("no input (OpenAPI specification) defined")
	}
	return opts, nil
}

func kebabToCamel(s string) string {
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func applyOverride(opts *Options, key, value string) {
	switch key {
	case "input":
		opts.Input = value
	case "output":
		opts.Output = value
	case "fetchTimeout":
		if n, err := strconv.Atoi(value); err == nil {
			opts.FetchTimeout = n
		}
	case "silent":
		opts.Silent = value == "true"
	case "defaultTag":
		opts.DefaultTag = value
	case "enumStyle":
		opts.EnumStyle = value
	case "modelPrefix":
		opts.ModelPrefix = value
	case "modelSuffix":
		opts.ModelSuffix = value
	case "servicePrefix":
		opts.ServicePrefix = value
	case "serviceSuffix":
		opts.ServiceSuffix = value
	case "configuration":
		opts.Configuration = value
	case "baseService":
		opts.BaseService = value
	case "requestBuilder":
		opts.RequestBuilder = value
	case "response":
		opts.Response = value
	case "skipJsonSuffix":
		opts.SkipJsonSuffix = value == "true"
	case "indexFile":
		opts.IndexFile = value == "true"
	case "useTempDir":
		opts.UseTempDir = value == "true"
	case "apiModule":
		opts.ApiModule = value == "true"
	case "includeTags":
		opts.IncludeTags = splitAndTrim(value)
	case "excludeTags":
		opts.ExcludeTags = splitAndTrim(value)
	case "excludePaths":
		opts.ExcludePaths = splitAndTrim(value)
	case "excludeParameters":
		opts.ExcludeParameters = splitAndTrim(value)
	case "promises":
		b := value == "true"
		opts.Promises = &b
	case "services":
		b := value == "true"
		opts.Services = &b
	case "ignoreUnusedModels":
		b := value == "true"
		opts.IgnoreUnusedModels = &b
	case "camelizeModelNames":
		b := value == "true"
		opts.CamelizeModelNames = &b
	case "enumArray":
		b := value == "true"
		opts.EnumArray = &b
	case "removeStaleFiles":
		b := value == "true"
		opts.RemoveStaleFiles = &b
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
