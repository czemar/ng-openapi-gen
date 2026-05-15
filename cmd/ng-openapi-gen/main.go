package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/generate"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := config.ParseOptionsFromArgs(os.Args[1:])
	if err != nil {
		return fmt.Errorf("parse options: %w", err)
	}

	input := opts.Input
	timeout := opts.FetchTimeout
	if timeout <= 0 {
		timeout = 20000
	}

	// Handle URL input
	if isURL(input) {
		tempFile, err := downloadSpec(input, timeout)
		if err != nil {
			return fmt.Errorf("download spec: %w", err)
		}
		defer os.Remove(tempFile)
		input = tempFile
	}

	// Parse the OpenAPI spec
	spec, err := openapi.ParseSpec(input)
	if err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}

	// Filter paths
	if len(opts.ExcludePaths) > 0 || len(opts.ExcludeTags) > 0 || len(opts.IncludeTags) > 0 {
		filterPaths(spec, opts)
	}

	// Determine templates directory
	execPath, _ := os.Executable()
	templatesDir := filepath.Join(filepath.Dir(execPath), "..", "..", "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// Try relative to working directory
		templatesDir = "templates"
	}

	// Create and run generator
	generator := generate.NewGenerator(spec, opts, templatesDir)
	if err := generator.Generate(); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	return nil
}

func isURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}

func downloadSpec(url string, timeout int) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("ng-openapi-gen-%d.json", time.Now().UnixNano()))

	f, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	if _, err := f.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	return tempFile, nil
}

func filterPaths(spec *openapi.Spec, opts *config.Options) {
	if spec.Paths == nil {
		return
	}

	excludePaths := makeSet(opts.ExcludePaths)
	excludeTags := makeSet(opts.ExcludeTags)
	includeTags := opts.IncludeTags

	for path, pathItem := range spec.Paths {
		if pathItem == nil {
			continue
		}
		if excludePaths[path] {
			delete(spec.Paths, path)
			continue
		}

		methods := []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}
		hasRemaining := false

		for _, method := range methods {
			op := getOperation(pathItem, method)
			if op == nil {
				continue
			}

			tags := op.Tags
			if len(tags) == 0 {
				continue
			}

			shouldExclude := false
			if len(excludeTags) > 0 {
				for _, tag := range tags {
					if excludeTags[tag] {
						shouldExclude = true
						break
					}
				}
			}
			if len(includeTags) > 0 {
				included := false
				for _, inclTag := range includeTags {
					for _, tag := range tags {
						if tag == inclTag {
							included = true
							break
						}
					}
					if included {
						break
					}
				}
				if !included {
					shouldExclude = true
				}
			}

			if shouldExclude {
				setOperationNil(pathItem, method)
			} else {
				hasRemaining = true
			}
		}

		if !hasRemaining {
			delete(spec.Paths, path)
		}
	}
}

func getOperation(pi *openapi.PathItem, method string) *openapi.Operation {
	switch method {
	case "get":
		return pi.Get
	case "put":
		return pi.Put
	case "post":
		return pi.Post
	case "delete":
		return pi.Delete
	case "options":
		return pi.Options
	case "head":
		return pi.Head
	case "patch":
		return pi.Patch
	case "trace":
		return pi.Trace
	}
	return nil
}

func setOperationNil(pi *openapi.PathItem, method string) {
	switch method {
	case "get":
		pi.Get = nil
	case "put":
		pi.Put = nil
	case "post":
		pi.Post = nil
	case "delete":
		pi.Delete = nil
	case "options":
		pi.Options = nil
	case "head":
		pi.Head = nil
	case "patch":
		pi.Patch = nil
	case "trace":
		pi.Trace = nil
	}
}

func makeSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}
