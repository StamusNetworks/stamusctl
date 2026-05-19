package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "myproject", false},
		{"valid with dash", "my-project", false},
		{"valid with underscore", "my_project", false},
		{"valid alphanumeric", "project123", false},
		{"empty name", "", true},
		{"path traversal dots", "../etc", true},
		{"path traversal slash", "my/project", true},
		{"path traversal backslash", "my\\project", true},
		{"special characters", "my@project", true},
		{"spaces", "my project", true},
		{"too long", strings.Repeat("a", 256), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid semver", "1.2.3", false},
		{"valid with v prefix", "v1.2.3", false},
		{"valid latest", "latest", false},
		{"valid with dash", "1.2.3-beta", false},
		{"empty version", "", true},
		{"path traversal", "../etc", true},
		{"with slash", "1.2/3", true},
		{"with backslash", "1.2\\3", true},
		{"too long", strings.Repeat("1", 256), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateParameterKey(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple key", "mykey", false},
		{"valid with dots", "my.key.name", false},
		{"valid with dash", "my-key", false},
		{"valid with underscore", "my_key", false},
		{"empty key", "", true},
		{"special characters", "my@key", true},
		{"spaces", "my key", true},
		{"too long", strings.Repeat("a", 256), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParameterKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateParameterKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateParameterValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple value", "myvalue", false},
		{"valid with spaces", "my value", false},
		{"valid with special chars", "my@value!", false},
		{"null byte", "my\x00value", true},
		{"newline", "my\nvalue", true},
		{"carriage return", "my\rvalue", true},
		{"tab", "my\tvalue", true},
		{"too long", strings.Repeat("a", 65537), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParameterValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateParameterValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "validation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		path    string
		baseDir string
		wantErr bool
	}{
		{"valid relative path", "subdir/file.txt", tmpDir, false},
		{"valid absolute path within base", filepath.Join(tmpDir, "file.txt"), tmpDir, false},
		{"path traversal attempt", "../etc/passwd", tmpDir, true},
		{"path traversal with clean", "subdir/../../etc/passwd", tmpDir, true},
		{"null byte in path", "file\x00.txt", tmpDir, true},
		{"newline in path", "file\n.txt", tmpDir, true},
		{"empty path", "", tmpDir, true},
		{"too long path", strings.Repeat("a", MaxPathLength+1), tmpDir, true},
		{"no baseDir with dots", "../etc", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizePath(tt.path, tt.baseDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.baseDir != "" {
				// Verify result is within baseDir
				absBase, _ := filepath.Abs(tt.baseDir)
				relPath, err := filepath.Rel(absBase, result)
				if err != nil || strings.HasPrefix(relPath, "..") {
					t.Errorf("SanitizePath() result escapes baseDir: %v", result)
				}
			}
		})
	}
}

func TestValidateScriptPath(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "validation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	allowedDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(allowedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	validScript := filepath.Join(allowedDir, "script.sh")
	if err := os.WriteFile(validScript, []byte("#!/bin/bash"), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		scriptPath  string
		allowedDirs []string
		wantErr     bool
	}{
		{"valid script in allowed dir", validScript, []string{allowedDir}, false},
		{"script outside allowed dir", "/tmp/malicious.sh", []string{allowedDir}, true},
		{"path traversal attempt", filepath.Join(allowedDir, "../../../etc/passwd"), []string{allowedDir}, true},
		{"empty script path", "", []string{allowedDir}, true},
		{"no allowed dirs", validScript, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScriptPath(tt.scriptPath, tt.allowedDirs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScriptPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMapDepth(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		maxDepth int
		wantErr  bool
	}{
		{
			name: "shallow map",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			maxDepth: 5,
			wantErr:  false,
		},
		{
			name: "nested map within limit",
			data: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "value",
					},
				},
			},
			maxDepth: 5,
			wantErr:  false,
		},
		{
			name: "nested map exceeding limit",
			data: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"level4": "value",
						},
					},
				},
			},
			maxDepth: 2,
			wantErr:  true,
		},
		{
			name: "array of maps within limit",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"nested": "value",
					},
				},
			},
			maxDepth: 5,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMapDepth(tt.data, tt.maxDepth)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMapDepth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeForTemplate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no special chars", "normal value", "normal value"},
		{"with template brackets", "{{ .Value }}", " .Value "},
		{"with comment start", "/* comment", " comment"},
		{"with comment end", "comment */", "comment "},
		{"mixed special chars", "{{ /* test */ }}", "  test  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeForTemplate(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeForTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Additional coverage tests
// ---------------------------------------------------------------------------

// TestSanitizePath_ValidNoBaseDir covers the success return path (line 176) when
// no baseDir is provided and the path has no traversal sequences.
func TestSanitizePath_ValidNoBaseDir(t *testing.T) {
	result, err := SanitizePath("/tmp/valid/path", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "/tmp/valid/path" {
		t.Errorf("expected /tmp/valid/path, got %s", result)
	}
}

// TestSanitizePath_RelativeNoBaseDir covers the relative-path + no-baseDir success path.
func TestSanitizePath_RelativeNoBaseDir(t *testing.T) {
	result, err := SanitizePath("subdir/file.txt", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "subdir/file.txt" {
		t.Errorf("expected subdir/file.txt, got %s", result)
	}
}

// TestValidateMapDepth_ArrayWithDeepMap covers the []interface{} path (line 234-239)
// where an array element is a map that exceeds the maximum depth.
func TestValidateMapDepth_ArrayWithDeepMap(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "value",
				},
			},
		},
	}
	err := ValidateMapDepth(data, 1)
	if err == nil {
		t.Error("expected error for depth-exceeding array element, got nil")
	}
}

// TestValidateScriptPath_NoAllowedDirs covers the case with an empty allowedDirs slice.
func TestValidateScriptPath_EmptyAllowedDirs(t *testing.T) {
	err := ValidateScriptPath("/tmp/script.sh", []string{})
	if err == nil {
		t.Error("expected error for empty allowed dirs, got nil")
	}
}

// TestValidateVersion_DoubleDotPassesRegexButFailsTraversal covers line 75-77:
// "1..2" passes the ValidVersionRegex (dots are allowed) but contains ".." which
// triggers the path-traversal check after the regex.
func TestValidateVersion_DoubleDotPassesRegexButFailsTraversal(t *testing.T) {
	err := ValidateVersion("1..2")
	if err == nil {
		t.Error("expected error for version '1..2' containing '..', got nil")
	}
}
