//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/generate"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

func generateWrapper(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return `{"error":"missing spec argument"}`
	}

	specBytes := []byte(args[0].String())
	configBytes := []byte("{}")
	if len(args) > 1 && args[1].String() != "" {
		configBytes = []byte(args[1].String())
	}

	opts := &config.Options{
		Output: "src/app/api",
	}
	if len(configBytes) > 0 {
		if err := json.Unmarshal(configBytes, opts); err != nil {
			return `{"error":"invalid config JSON: ` + err.Error() + `"}`
		}
	}
	opts.Silent = true

	spec, err := openapi.ParseSpecBytes(specBytes, "")
	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}

	gen := generate.NewGenerator(spec, opts)
	gen.Files = make(map[string][]byte)

	if err := gen.Generate(); err != nil {
		return `{"error":"` + err.Error() + `"}`
	}

	result := make(map[string]string, len(gen.Files))
	for path, content := range gen.Files {
		result[path] = string(content)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	return string(jsonBytes)
}

func main() {
	js.Global().Set("generateOpenAPIDemo", js.FuncOf(generateWrapper))
	select {}
}
