package models

import (
	"archive/tar"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	cp "github.com/otiai10/copy"
	"github.com/spf13/afero"
)

type RegistryInfo struct {
	Registry string `json:"registry"`
	Username string `json:"username"`
	Password string `json:"password"`
	Verif    bool   `json:"verif"`
}

func (r *RegistryInfo) ValidateRegistry() error {
	if r.Registry == "" {
		return fmt.Errorf("missing registry")
	}
	return nil
}

func (r *RegistryInfo) ValidateAllRegistry() error {
	if r.Registry == "" {
		return fmt.Errorf("missing registry")
	}
	if r.Username == "" {
		return fmt.Errorf("missing username")
	}
	if r.Password == "" {
		return fmt.Errorf("missing password")
	}
	return nil
}

var (
	ErrMarshalingAuthConfig = errors.New("Error marshaling auth config")
	ErrPullingImage         = errors.New("Error pulling image")
)

func (r *RegistryInfo) TryPullConfig(ctx context.Context, cli *client.Client, imageName, imageURL string) error {
	logger := logging.Sugar.With("imageURL", imageURL, "imageName", imageName)

	logger.Debug("Try pulling")

	// Create docker client

	// Create auth config
	pullOptions := image.PullOptions{}
	if r.Username != "" && r.Password != "" {
		authConfig := registry.AuthConfig{
			Username: r.Username,
			Password: r.Password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMarshalingAuthConfig, err)
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		pullOptions = image.PullOptions{
			RegistryAuth: authStr,
		}
	}

	// Pull image
	out, err := cli.ImagePull(ctx, imageURL, pullOptions)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPullingImage, err)
	}
	defer out.Close()

	// Parse progress details
	type ImagePullResponse struct {
		Progress string `json:"progress"`
		Status   string `json:"status"`
	}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		var pullResp ImagePullResponse
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &pullResp); err != nil {
			fmt.Fprintf(os.Stderr, "\rError unmarshalling progress detail: %v", err)

			continue // Skip lines that can't be unmarshalled
		}

		if pullResp.Progress != "" {
			fmt.Printf("\r%s %s", pullResp.Status, pullResp.Progress)
		}
	}
	logger.Info("Got configuration")

	return nil
}

func (r *RegistryInfo) PullConfigAndUnwrap(destPath string, project, version string) error {
	ctx := context.Background()

	imageName := "/" + project + ":" + version
	imageURL := r.Registry + imageName

	logger := logging.Sugar.With("imageURL", imageURL, "imageName", imageName)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()
	err = r.TryPullConfig(ctx, cli, imageName, imageURL)
	if err != nil {
		if errors.Is(err, ErrMarshalingAuthConfig) {
			logger.Info("Error marshaling auth config")
			return err
		}
		if errors.Is(err, ErrPullingImage) {
			logger.Info("Error pulling image")
			return err
		}
	}

	// Run container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageURL,
		Cmd:   []string{"sleep", "60"},
	}, nil, nil, nil, "")
	if err != nil {
		logger.Debug("Container creation failed")
		return err
	}
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		logger.Debug("Container start failed")
		return err
	}

	// Kill container
	defer func() {
		if err := cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); err != nil {
			fmt.Printf("Failed to remove container: %v\n", err)
		}
	}()

	// Extract conf from container
	srcPaths := []string{"/data", "/sbin"} // Source path inside the container
	// Remove existing configuration
	if err := app.FS.RemoveAll(filepath.Join(destPath, version)); err != nil {
		return err
	}
	// Copy files from container
	for _, srcPath := range srcPaths {
		if err := copyFromContainer(cli, ctx, resp.ID, srcPath, destPath); err != nil {
			logger.Debug("Container copy from failed")
			return err
		}
	}
	// Move files to correct locations
	originPath := filepath.Join(destPath, "data/")
	versionPath := filepath.Join(destPath, version+"/")
	if err := app.FS.Rename(originPath, versionPath); err != nil {
		return err
	}
	// Copy templates latest to templates version
	versionFromTemplate, err := afero.ReadFile(app.FS, versionPath+"/version")
	if err != nil {
		return err
	}

	if versionPath != filepath.Join(destPath, string(versionFromTemplate)) {
		err = cp.Copy(versionPath, filepath.Join(destPath, string(versionFromTemplate)))
		if err != nil {
			return err
		}
	}
	logger.Info("Configuration extracted")

	logger.Debug("Pull success")
	return nil
}

func copyFromContainer(cli *client.Client, ctx context.Context, containerID, srcPath, destPath string) error {
	reader, _, err := cli.CopyFromContainer(ctx, containerID, srcPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(destPath, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := app.FS.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			logger := logging.Sugar.With("target", target, "srcPath", srcPath, "containerID", containerID)
			if err := app.FS.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			outFile, err := app.FS.Create(target)
			if err != nil {
				logger.Debug("creating failed")
				return err
			}
			written, err := io.Copy(outFile, tr)
			if err != nil {
				outFile.Close()
				logger.Debug("copying failed")
				return err
			}
			logger.Debug("copied ", written, " bytes")
			outFile.Close()
		}
	}

	return nil
}

// getRegistryCredentials reads registry credentials from the config file
// This avoids importing the stamus package which would create an import cycle
// registryHost should be extracted from the image reference (e.g., "ghcr.io" from "ghcr.io/org/image:tag")
func getRegistryCredentials(registryHost string) (*RegistryInfo, error) {
	// Read config file
	configPath := filepath.Join(app.ConfigFolder, "config.json")
	bytes, err := afero.ReadFile(app.FS, configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create config directory and empty config file
			if mkdirErr := app.FS.MkdirAll(app.ConfigFolder, 0o755); mkdirErr != nil {
				return nil, fmt.Errorf("failed to create config directory: %w", mkdirErr)
			}
			emptyConfig := []byte(`{"registries":{}}`)
			if writeErr := afero.WriteFile(app.FS, configPath, emptyConfig, 0o644); writeErr != nil {
				return nil, fmt.Errorf("failed to create config file: %w", writeErr)
			}
			// Return empty credentials for anonymous access
			return &RegistryInfo{
				Registry: registryHost,
				Username: "",
				Password: "",
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse just the registries part
	var config struct {
		Registries map[string]map[string]string `json:"registries"`
	}
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract registry key (hostname without port)
	registryKey := strings.Split(registryHost, ":")[0]

	// Try to find matching registry credentials
	// Try exact match first, then without port
	registriesToTry := []string{registryHost, registryKey}

	for _, tryRegistry := range registriesToTry {
		if logins, ok := config.Registries[tryRegistry]; ok {
			for username, password := range logins {
				return &RegistryInfo{
					Registry: tryRegistry,
					Username: username,
					Password: password,
				}, nil
			}
		}
	}

	// If no credentials found, return empty credentials for anonymous access
	// This allows pulling from public registries
	return &RegistryInfo{
		Registry: registryHost,
		Username: "",
		Password: "",
	}, nil
}

// isRemoteInclude checks if path looks like a registry URL
func isRemoteInclude(path string) bool {
	// Local paths start with ., /, or ~
	if strings.HasPrefix(path, ".") || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "~") {
		return false
	}
	// Remote must have : or @ for tag/digest
	return strings.Contains(path, ":") || strings.Contains(path, "@")
}

// parseRemoteInclude splits URL into image ref and file path
// Input: "registry.com/namespace/image:tag/path/to/file.yaml"
// Returns: imageRef="registry.com/namespace/image:tag", filePath="path/to/file.yaml"
func parseRemoteInclude(include string) (imageRef, filePath string, err error) {
	// Find the position of : or @ for tag/digest separator
	tagIdx := strings.LastIndex(include, ":")
	digestIdx := strings.LastIndex(include, "@")

	// Determine which separator to use
	sepIdx := tagIdx
	if digestIdx > tagIdx {
		sepIdx = digestIdx
	}

	if sepIdx == -1 {
		return "", "", fmt.Errorf("invalid remote include format: missing tag or digest")
	}

	// Find the next / after the separator
	filePathStart := strings.Index(include[sepIdx:], "/")
	if filePathStart == -1 {
		return "", "", fmt.Errorf("invalid remote include format: missing file path")
	}

	// Calculate absolute position
	filePathStart += sepIdx

	// Split into image ref and file path
	imageRef = include[:filePathStart]
	filePath = include[filePathStart+1:] // Skip the /

	// Validate file path for security (no .. or leading /)
	if strings.Contains(filePath, "..") {
		return "", "", fmt.Errorf("invalid file path: contains ..")
	}
	if strings.HasPrefix(filePath, "/") {
		return "", "", fmt.Errorf("invalid file path: starts with /")
	}

	return imageRef, filePath, nil
}

// hashURL generates a SHA256 hash of a URL for use as a cache key
func hashURL(url string) string {
	h := sha256.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))
}

// getCachedRemoteInclude checks cache for URL
func getCachedRemoteInclude(url string) ([]byte, bool) {
	cacheKey := hashURL(url)
	cachePath := filepath.Join(xdg.CacheHome, "stamus", "remote-includes", cacheKey)

	content, err := afero.ReadFile(app.FS, cachePath)
	if err != nil {
		return nil, false // Cache miss
	}
	return content, true
}

// setCachedRemoteInclude saves to cache
func setCachedRemoteInclude(url string, content []byte) error {
	cacheKey := hashURL(url)
	cachePath := filepath.Join(xdg.CacheHome, "stamus", "remote-includes", cacheKey)

	// Create directory if needed
	if err := app.FS.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}

	// Write file with restricted permissions (may contain registry credentials)
	return afero.WriteFile(app.FS, cachePath, content, 0o600)
}

// pullRemoteInclude fetches a single file from registry
func pullRemoteInclude(ctx context.Context, registryInfo *RegistryInfo, imageRef, filePath string) ([]byte, error) {
	logger := logging.Sugar.With("imageRef", imageRef, "filePath", filePath)

	// Add overall timeout for the operation to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// 1. Pull image (reuse TryPullConfig pattern)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	err = registryInfo.TryPullConfig(ctx, cli, "", imageRef)
	if err != nil {
		logger.Debug("Failed to pull image")
		return nil, err
	}

	// 2. Create temporary container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageRef,
		Cmd:   []string{"sleep", "60"},
	}, nil, nil, nil, "")
	if err != nil {
		logger.Debug("Container creation failed")
		return nil, err
	}
	defer func() {
		// Use background context to ensure cleanup happens even if parent context cancelled
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := cli.ContainerRemove(cleanupCtx, resp.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			logger.Error("Failed to remove container", "containerID", resp.ID, "error", err)
		}
	}()

	// 3. Extract file from container (reuse existing tar code from copyFromContainer)
	// Normalize the path to ensure it starts with /
	extractPath := "/" + strings.TrimPrefix(filePath, "/")
	reader, _, err := cli.CopyFromContainer(ctx, resp.ID, extractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract file %s from container: %w", filePath, err)
	}
	defer reader.Close()

	// 4. Read tar archive and extract file
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Match either full path or just the filename
		if header.Name == filePath || header.Name == filepath.Base(filePath) {
			return io.ReadAll(tarReader)
		}
	}

	return nil, fmt.Errorf("file not found in image: %s", filePath)
}
