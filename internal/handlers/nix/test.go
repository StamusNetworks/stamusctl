package handlers

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/nix"
	"stamus-ctl/internal/validation"

	"github.com/spf13/afero"
)

type NixTestHandlerInputs struct {
	Config string
	Filter string
}

// TestResult holds the outcome of a single test file execution.
type TestResult struct {
	Name   string
	Passed bool
	Err    error
}

func NixTestHandler(params NixTestHandlerInputs) error {
	logger := logging.Sugar.With("Config", params.Config, "Filter", params.Filter)

	if !nix.IsNixOS() {
		return fmt.Errorf("this system is not running NixOS")
	}

	configPath := params.Config
	if !app.IsCtl() {
		configPath = app.GetConfigsFolder(params.Config)
	}

	exists, err := afero.DirExists(app.FS, configPath)
	if err != nil || !exists {
		return fmt.Errorf("configuration %q not found — run 'stamusctl nix init' first", params.Config)
	}

	testsDir := filepath.Join(configPath, "tests")
	exists, err = afero.DirExists(app.FS, testsDir)
	if err != nil || !exists {
		return fmt.Errorf("no tests directory found at %q", testsDir)
	}

	testFiles, err := discoverTests(testsDir, params.Filter)
	if err != nil {
		return fmt.Errorf("failed to discover tests: %w", err)
	}

	if len(testFiles) == 0 {
		if params.Filter != "" {
			return fmt.Errorf("no test files matching filter %q in %s", params.Filter, testsDir)
		}
		return fmt.Errorf("no test files (.sh or .nix) found in %s", testsDir)
	}

	logger.Infof("Found %d test(s) in %s", len(testFiles), testsDir)

	var results []TestResult
	for _, tf := range testFiles {
		name := filepath.Base(tf)

		if err := validation.ValidateScriptPath(tf, []string{testsDir}); err != nil {
			results = append(results, TestResult{Name: name, Passed: false, Err: err})
			continue
		}

		var runErr error
		ext := strings.ToLower(filepath.Ext(tf))
		switch ext {
		case ".sh":
			runErr = nix.RunShellTest(tf, configPath)
		case ".nix":
			runErr = nix.RunNixTest(tf, configPath)
		default:
			continue
		}

		results = append(results, TestResult{
			Name:   name,
			Passed: runErr == nil,
			Err:    runErr,
		})
	}

	return printSummary(results)
}

func discoverTests(testsDir string, filter string) ([]string, error) {
	entries, err := afero.ReadDir(app.FS, testsDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".sh" && ext != ".nix" {
			continue
		}
		if filter != "" {
			matched, matchErr := filepath.Match(filter, entry.Name())
			if matchErr != nil {
				return nil, fmt.Errorf("invalid filter pattern %q: %w", filter, matchErr)
			}
			if !matched {
				continue
			}
		}
		files = append(files, filepath.Join(testsDir, entry.Name()))
	}

	sort.Strings(files)
	return files, nil
}

func printSummary(results []TestResult) error {
	passed := 0
	failed := 0

	fmt.Println()
	for _, r := range results {
		if r.Passed {
			fmt.Printf("  PASS  %s\n", r.Name)
			passed++
		} else {
			fmt.Printf("  FAIL  %s: %v\n", r.Name, r.Err)
			failed++
		}
	}

	fmt.Println()
	fmt.Printf("Results: %d passed, %d failed, %d total\n", passed, failed, len(results))

	if failed > 0 {
		return fmt.Errorf("%d test(s) failed", failed)
	}
	return nil
}
