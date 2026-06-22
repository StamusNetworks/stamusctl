package validation

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// MaxYAMLDepth is the maximum nesting depth allowed in YAML files
	MaxYAMLDepth = 50

	// MaxIncludeDepth is the maximum depth for recursive includes
	MaxIncludeDepth = 10

	// MaxPathLength is the maximum allowed path length
	MaxPathLength = 4096
)

var (
	// ValidProjectNameRegex matches valid project names (alphanumeric, dash, underscore only)
	ValidProjectNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// ValidVersionRegex matches semantic versioning and simple version strings
	ValidVersionRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// ValidParameterKeyRegex matches valid parameter keys
	ValidParameterKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// ForbiddenPathChars are characters that should never appear in file paths
	ForbiddenPathChars = []string{"\x00", "\r", "\n", "\t"}

	// PathTraversalSequences are patterns that indicate path traversal attempts
	PathTraversalSequences = []string{"..", "~", "$"}
)

// ValidateProjectName validates that a project name is safe to use
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("project name too long (max 255 characters)")
	}

	if !ValidProjectNameRegex.MatchString(name) {
		return fmt.Errorf("project name contains invalid characters (only alphanumeric, dash, and underscore allowed)")
	}

	// Check for path traversal attempts
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("project name contains path traversal sequences")
	}

	return nil
}

// ValidateVersion validates that a version string is safe to use
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if len(version) > 255 {
		return fmt.Errorf("version string too long (max 255 characters)")
	}

	if !ValidVersionRegex.MatchString(version) {
		return fmt.Errorf("version contains invalid characters")
	}

	// Check for path traversal attempts
	if strings.Contains(version, "..") || strings.Contains(version, "/") || strings.Contains(version, "\\") {
		return fmt.Errorf("version contains path traversal sequences")
	}

	return nil
}

// ValidateParameterKey validates that a parameter key is safe to use
func ValidateParameterKey(key string) error {
	if key == "" {
		return fmt.Errorf("parameter key cannot be empty")
	}

	if len(key) > 255 {
		return fmt.Errorf("parameter key too long (max 255 characters)")
	}

	if !ValidParameterKeyRegex.MatchString(key) {
		return fmt.Errorf("parameter key contains invalid characters")
	}

	return nil
}

// ValidateParameterValue validates that a parameter value is safe to use
func ValidateParameterValue(value string) error {
	if len(value) > 65536 {
		return fmt.Errorf("parameter value too long (max 65536 characters)")
	}

	// Check for null bytes and control characters
	for _, forbidden := range ForbiddenPathChars {
		if strings.Contains(value, forbidden) {
			return fmt.Errorf("parameter value contains forbidden control characters")
		}
	}

	return nil
}

// SanitizePath validates and sanitizes a file path to prevent traversal attacks
func SanitizePath(path string, baseDir string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	if len(path) > MaxPathLength {
		return "", fmt.Errorf("path too long (max %d characters)", MaxPathLength)
	}

	// Check for null bytes and control characters
	for _, forbidden := range ForbiddenPathChars {
		if strings.Contains(path, forbidden) {
			return "", fmt.Errorf("path contains forbidden control characters")
		}
	}

	// Clean the path (removes .., . and duplicate slashes)
	cleaned := filepath.Clean(path)

	// If baseDir is provided, make path absolute relative to baseDir
	if baseDir != "" {
		// Get absolute path of baseDir
		absBase, err := filepath.Abs(baseDir)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute base path: %w", err)
		}

		// Join with cleaned path
		var absPath string
		if filepath.IsAbs(cleaned) {
			absPath = cleaned
		} else {
			absPath = filepath.Join(absBase, cleaned)
		}

		// Get absolute path
		absPath, err = filepath.Abs(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Verify the resulting path is within baseDir
		relPath, err := filepath.Rel(absBase, absPath)
		if err != nil {
			return "", fmt.Errorf("failed to get relative path: %w", err)
		}

		// Check if path escapes baseDir (starts with ..)
		if strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." {
			return "", fmt.Errorf("path traversal attempt detected: path escapes base directory")
		}

		return absPath, nil
	}

	// If no baseDir, there is no containment root to verify against, so reject
	// any traversal or expansion sequence outright. Check the original path as
	// well as the cleaned form: filepath.Clean collapses sequences like
	// "a/../../b" and would otherwise hide them from the post-clean check.
	for _, seq := range PathTraversalSequences {
		if strings.Contains(path, seq) || strings.Contains(cleaned, seq) {
			return "", fmt.Errorf("path contains traversal sequences")
		}
	}

	return cleaned, nil
}

// ValidateScriptPath validates that a script path is within allowed directories
func ValidateScriptPath(scriptPath string, allowedDirs []string) error {
	if scriptPath == "" {
		return fmt.Errorf("script path cannot be empty")
	}

	// Get absolute path
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute script path: %w", err)
	}

	// Check if path is within any allowed directory
	allowed := false
	for _, allowedDir := range allowedDirs {
		absAllowedDir, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}

		relPath, err := filepath.Rel(absAllowedDir, absPath)
		if err != nil {
			continue
		}

		// Check if path is within this allowed directory
		if !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) && relPath != ".." {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("script path is not within allowed directories")
	}

	return nil
}

// ValidateMapDepth checks if a map has nested depth greater than max
func ValidateMapDepth(data map[string]interface{}, maxDepth int) error {
	return validateMapDepthRecursive(data, 0, maxDepth)
}

func validateMapDepthRecursive(data map[string]interface{}, currentDepth int, maxDepth int) error {
	if currentDepth > maxDepth {
		return fmt.Errorf("map nesting depth exceeds maximum allowed depth of %d", maxDepth)
	}

	for _, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			if err := validateMapDepthRecursive(v, currentDepth+1, maxDepth); err != nil {
				return err
			}
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if err := validateMapDepthRecursive(itemMap, currentDepth+1, maxDepth); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// SanitizeForTemplate removes potentially dangerous content from template data
func SanitizeForTemplate(value string) string {
	// Remove any template directive sequences
	value = strings.ReplaceAll(value, "{{", "")
	value = strings.ReplaceAll(value, "}}", "")
	value = strings.ReplaceAll(value, "/*", "")
	value = strings.ReplaceAll(value, "*/", "")

	return value
}
