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

func (r *RegistryInfo) PullConfig(destPath string, imageName string) error {
	imageUrl := r.Registry + imageName

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
	fmt.Printf("\rGot configuration                                                                                 ")
	fmt.Println()

	// Run container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageUrl,
		Cmd:   []string{"sleep 60"},
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
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
	if err := os.RemoveAll(filepath.Join(destPath, "latest")); err != nil {
		return err
	}
	// Copy files from container
	for _, srcPath := range srcPaths {
		if err := copyFromContainer(cli, ctx, resp.ID, srcPath, destPath); err != nil {
			return err
		}
	}
	// Move files to correct locations
	originPath := filepath.Join(destPath, "data/")
	latestPath := filepath.Join(destPath, "latest/")
	if err := os.Rename(originPath, latestPath); err != nil {
		return err
	}
	// Copy templates latest to templates version
	version, err := os.ReadFile(filepath.Join(latestPath, "version"))
	if err != nil {
		return err
	}
	err = cp.Copy(latestPath, filepath.Join(destPath, string(version)))
	if err != nil {
		return err
	}
	fmt.Println("Configuration extracted")

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
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
