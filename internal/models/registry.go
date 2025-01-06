package models

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"stamus-ctl/internal/logging"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	cp "github.com/otiai10/copy"
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

func (r *RegistryInfo) PullConfig(destPath string, project, version string) error {
	imageName := "/" + project + ":" + version
	imageUrl := r.Registry + imageName

	logger := logging.Sugar.With("imageUrl", imageUrl, "imageName", imageName)

	logger.Debug("Try pulling")

	// Create docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	// Create auth config
	pullOptions := image.PullOptions{}
	if r.Username != "" && r.Password != "" {
		authConfig := registry.AuthConfig{
			Username: r.Username,
			Password: r.Password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return err
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		pullOptions = image.PullOptions{
			RegistryAuth: authStr,
		}
	}

	// Pull image
	out, err := cli.ImagePull(ctx, imageUrl, pullOptions)
	if err != nil {
		logger.Debug("Pull failed")
		return err
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

	// Run container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageUrl,
		Cmd:   []string{"sleep 60"},
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
	if err := os.RemoveAll(filepath.Join(destPath, version)); err != nil {
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
	if err := os.Rename(originPath, versionPath); err != nil {
		return err
	}
	// Copy templates latest to templates version
	versionFromTemplate, err := os.ReadFile(versionPath + "/version")
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
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			logger := logging.Sugar.With("target", target, "srcPath", srcPath, "containerID", containerID)
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
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
