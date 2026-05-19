package models

import (
	"strings"
	"testing"
)

func TestIsRemoteInclude(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "local relative path with dot",
			path:     "./local/file.yaml",
			expected: false,
		},
		{
			name:     "local relative path without dot",
			path:     "local/file.yaml",
			expected: false,
		},
		{
			name:     "local absolute path",
			path:     "/absolute/path.yaml",
			expected: false,
		},
		{
			name:     "local home path",
			path:     "~/home/file.yaml",
			expected: false,
		},
		{
			name:     "remote with tag",
			path:     "ghcr.io/org/image:v1/file.yaml",
			expected: true,
		},
		{
			name:     "remote with digest",
			path:     "registry.com:5000/image@sha256:abc123/file.yaml",
			expected: true,
		},
		{
			name:     "remote docker hub",
			path:     "docker.io/library/nginx:latest/config.yaml",
			expected: true,
		},
		{
			name:     "remote with multi-level namespace",
			path:     "ghcr.io/stamusnetworks/configs:v1.2.3/nginx/config.yaml",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRemoteInclude(tt.path)
			if result != tt.expected {
				t.Errorf("isRemoteInclude(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestParseRemoteInclude(t *testing.T) {
	tests := []struct {
		name         string
		include      string
		wantImageRef string
		wantFilePath string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "simple tag with single file",
			include:      "ghcr.io/org/image:v1/file.yaml",
			wantImageRef: "ghcr.io/org/image:v1",
			wantFilePath: "file.yaml",
			wantErr:      false,
		},
		{
			name:         "tag with nested path",
			include:      "ghcr.io/stamusnetworks/configs:v1.2.3/nginx/config.yaml",
			wantImageRef: "ghcr.io/stamusnetworks/configs:v1.2.3",
			wantFilePath: "nginx/config.yaml",
			wantErr:      false,
		},
		{
			name:         "digest reference",
			include:      "registry.com/image@sha256:abc123def456/path/to/file.yaml",
			wantImageRef: "registry.com/image@sha256:abc123def456",
			wantFilePath: "path/to/file.yaml",
			wantErr:      false,
		},
		{
			name:         "registry with port",
			include:      "localhost:5000/myimage:v1/config.yaml",
			wantImageRef: "localhost:5000/myimage:v1",
			wantFilePath: "config.yaml",
			wantErr:      false,
		},
		{
			name:         "multi-level namespace",
			include:      "ghcr.io/org/team/project:tag/folder/subfolder/file.yaml",
			wantImageRef: "ghcr.io/org/team/project:tag",
			wantFilePath: "folder/subfolder/file.yaml",
			wantErr:      false,
		},
		{
			name:        "missing tag or digest",
			include:     "ghcr.io/org/image/file.yaml",
			wantErr:     true,
			errContains: "missing tag or digest",
		},
		{
			name:        "missing file path",
			include:     "ghcr.io/org/image:v1",
			wantErr:     true,
			errContains: "missing file path",
		},
		{
			name:        "path traversal attempt",
			include:     "ghcr.io/org/image:v1/../../../etc/passwd",
			wantErr:     true,
			errContains: "contains ..",
		},
		{
			name:        "absolute file path",
			include:     "ghcr.io/org/image:v1//etc/passwd",
			wantErr:     true,
			errContains: "starts with /",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageRef, filePath, err := parseRemoteInclude(tt.include)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseRemoteInclude(%q) expected error containing %q, got nil", tt.include, tt.errContains)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseRemoteInclude(%q) error = %v, want error containing %q", tt.include, err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("parseRemoteInclude(%q) unexpected error: %v", tt.include, err)
				return
			}

			if imageRef != tt.wantImageRef {
				t.Errorf("parseRemoteInclude(%q) imageRef = %q, want %q", tt.include, imageRef, tt.wantImageRef)
			}

			if filePath != tt.wantFilePath {
				t.Errorf("parseRemoteInclude(%q) filePath = %q, want %q", tt.include, filePath, tt.wantFilePath)
			}
		})
	}
}

func TestHashURL(t *testing.T) {
	tests := []struct {
		name string
		url1 string
		url2 string
		same bool
	}{
		{
			name: "same URL produces same hash",
			url1: "ghcr.io/org/image:v1/file.yaml",
			url2: "ghcr.io/org/image:v1/file.yaml",
			same: true,
		},
		{
			name: "different URLs produce different hashes",
			url1: "ghcr.io/org/image:v1/file.yaml",
			url2: "ghcr.io/org/image:v2/file.yaml",
			same: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hashURL(tt.url1)
			hash2 := hashURL(tt.url2)

			// Check that hashes are non-empty
			if hash1 == "" {
				t.Errorf("hashURL(%q) returned empty string", tt.url1)
			}
			if hash2 == "" {
				t.Errorf("hashURL(%q) returned empty string", tt.url2)
			}

			// Check hash length (SHA256 produces 64 hex characters)
			if len(hash1) != 64 {
				t.Errorf("hashURL(%q) hash length = %d, want 64", tt.url1, len(hash1))
			}

			// Check same/different expectation
			if tt.same && hash1 != hash2 {
				t.Errorf("hashURL() same URLs produced different hashes: %q vs %q", hash1, hash2)
			}
			if !tt.same && hash1 == hash2 {
				t.Errorf("hashURL() different URLs produced same hash: %q", hash1)
			}
		})
	}
}

func TestCacheOperations(t *testing.T) {
	// Note: These tests require a filesystem and are more integration-style tests
	// They test the cache get/set operations together

	testURL := "ghcr.io/test/image:v1/test.yaml"
	testContent := []byte("test: value\nfoo: bar\n")

	// Test cache miss
	t.Run("cache miss", func(t *testing.T) {
		content, found := getCachedRemoteInclude("nonexistent-url-" + t.Name())
		if found {
			t.Errorf("getCachedRemoteInclude() found = true, want false for nonexistent URL")
		}
		if content != nil {
			t.Errorf("getCachedRemoteInclude() content = %v, want nil", content)
		}
	})

	// Test cache set and get
	t.Run("cache set and get", func(t *testing.T) {
		uniqueURL := testURL + "-" + t.Name()

		// Set cache
		err := setCachedRemoteInclude(uniqueURL, testContent)
		if err != nil {
			t.Errorf("setCachedRemoteInclude() error = %v, want nil", err)
			return
		}

		// Get from cache
		content, found := getCachedRemoteInclude(uniqueURL)
		if !found {
			t.Errorf("getCachedRemoteInclude() found = false, want true")
			return
		}

		if string(content) != string(testContent) {
			t.Errorf("getCachedRemoteInclude() content = %q, want %q", string(content), string(testContent))
		}
	})
}
