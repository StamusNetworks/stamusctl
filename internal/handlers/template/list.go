package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/stamus"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/jedib0t/go-pretty/v6/table"
)


func ListHandler() error {
	// Get embedded templates
	embeddedTemplates := getEmbeddedTemplates()

	// Get local templates
	localTemplates, err := getLocalTemplates()
	if err != nil {
		return err
	}

	// Get remote templates
	remoteTemplates, err := getRemoteTemplates()
	if err != nil {
		// Remote template discovery failed, but don't break the command
		remoteTemplates = []TemplateInfo{} // Ensure it's empty slice, not nil
	}

	// Prepare data
	rows := []table.Row{}

	// Add embedded templates
	for _, template := range embeddedTemplates {
		rows = append(rows, table.Row{template.Name, template.Version, template.Source, template.Status})
	}

	// Add local templates
	for _, template := range localTemplates {
		rows = append(rows, table.Row{template.Name, template.Version, template.Source, template.Status})
	}

	// Add remote templates (and mark cached ones)
	for _, template := range remoteTemplates {
		// Check if this remote template is also available locally (filesystem or Docker image)
		if isTemplateAvailableLocally(template.Name, template.Version, localTemplates) {
			template.Status = "cached (ready)"
		}
		rows = append(rows, table.Row{template.Name, template.Version, template.Source, template.Status})
	}

	// Merge duplicate entries (e.g., "1.1.4" and "1.1.4 (latest)")
	rows = mergeDuplicateVersions(rows)

	// Sort by name, then by version
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][0].(string) == rows[j][0].(string) {
			return rows[i][1].(string) < rows[j][1].(string)
		}
		return rows[i][0].(string) < rows[j][0].(string)
	})

	// Print
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	header := table.Row{"Template", "Version", "Source", "Status"}
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.AppendFooter(header)
	t.Render()
	return nil
}

type TemplateInfo struct {
	Name    string
	Version string
	Source  string
	Status  string
}

func getEmbeddedTemplates() []TemplateInfo {
	// For now, we know there's at least the embedded clearndr template
	templates := []TemplateInfo{
		{
			Name:    "clearndr",
			Version: "embedded",
			Source:  "built-in binary",
			Status:  "ready",
		},
	}
	return templates
}

func getLocalTemplates() ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// Check if templates folder exists
	if _, err := os.Stat(app.TemplatesFolder); os.IsNotExist(err) {
		return templates, nil
	}

	// Read templates directory
	entries, err := os.ReadDir(app.TemplatesFolder)
	if err != nil {
		return templates, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		templateName := entry.Name()
		templatePath := filepath.Join(app.TemplatesFolder, templateName)

		// Check for different versions in the template directory
		versions, err := getTemplateVersions(templatePath)
		if err != nil {
			continue
		}

		if len(versions) == 0 {
			// No versioned subdirectories, check if there's a config.yaml directly
			configPath := filepath.Join(templatePath, "config.yaml")
			if _, err := os.Stat(configPath); err == nil {
				// Try to resolve "latest" to actual version
				actualVersion := "latest"
				if resolvedVersion, err := resolveLatestVersion(templateName); err == nil && resolvedVersion != "latest" {
					actualVersion = resolvedVersion + " (latest)"
				}

				templates = append(templates, TemplateInfo{
					Name:    templateName,
					Version: actualVersion,
					Source:  "local filesystem",
					Status:  "ready",
				})
			}
		} else {
			// Add each version found
			for _, version := range versions {
				// Check if this is a "latest" directory and try to resolve it
				actualVersion := version
				if version == "latest" {
					if resolvedVersion, err := resolveLatestVersion(templateName); err == nil && resolvedVersion != "latest" {
						actualVersion = resolvedVersion + " (latest)"
					}
				}

				templates = append(templates, TemplateInfo{
					Name:    templateName,
					Version: actualVersion,
					Source:  "local filesystem",
					Status:  "ready",
				})
			}
		}
	}

	return templates, nil
}

func getTemplateVersions(templatePath string) ([]string, error) {
	var versions []string

	entries, err := os.ReadDir(templatePath)
	if err != nil {
		return versions, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		versionPath := filepath.Join(templatePath, entry.Name())
		configPath := filepath.Join(versionPath, "config.yaml")

		// Check if this version directory contains a config.yaml
		if _, err := os.Stat(configPath); err == nil {
			versions = append(versions, entry.Name())
		}
	}

	return versions, nil
}

type RegistryTagsResponse struct {
	Tags []string `json:"tags"`
}

func getRemoteTemplates() ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// Get the default registry
	registryURL := app.DefaultRegistry
	if registryURL == "" {
		return templates, fmt.Errorf("no default registry configured")
	}

	// Get authentication info if available
	config, err := stamus.GetStamusConfig()
	if err != nil {
		return templates, fmt.Errorf("error getting config: %v", err)
	}

	// Known template names to check
	templateNames := []string{"clearndr"}

	// Query registry API for all available versions
	return queryRegistryForTemplates(registryURL, templateNames, config)
}

func queryRegistryForTemplates(registryURL string, templateNames []string, config *stamus.Config) ([]TemplateInfo, error) {
	// For GHCR, we need to use GitHub Packages API
	if strings.Contains(registryURL, "ghcr.io") {
		return queryGHCRForTemplates(registryURL, templateNames, config)
	}

	// For other registries, use Docker Registry V2 API
	return queryDockerRegistryForTemplates(registryURL, templateNames, config)
}

func queryGHCRForTemplates(registryURL string, templateNames []string, config *stamus.Config) ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// Parse GHCR URL: ghcr.io/stamusnetworks/stamusctl-templates
	parts := strings.Split(registryURL, "/")
	if len(parts) < 3 {
		return templates, fmt.Errorf("invalid GHCR URL format")
	}

	owner := parts[1]       // stamusnetworks
	repoName := parts[2]    // stamusctl-templates

	// For open source GHCR packages, use GitHub Releases API instead of registry API
	// This is public and doesn't require authentication
	return fetchGitHubReleases(owner, repoName, templateNames)
}

func queryDockerRegistryForTemplates(registryURL string, templateNames []string, config *stamus.Config) ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// Standard Docker Registry V2 API
	baseURL := "https://" + registryURL + "/v2"

	for _, templateName := range templateNames {
		url := fmt.Sprintf("%s/%s/tags/list", baseURL, templateName)

		tags, err := fetchDockerRegistryTags(url, config, registryURL)
		if err != nil {
			continue // Skip if this template doesn't exist
		}

		for _, tag := range tags {
			templates = append(templates, TemplateInfo{
				Name:    templateName,
				Version: tag,
				Source:  registryURL,
				Status:  "download needed",
			})
		}
	}

	return templates, nil
}

func fetchGitHubPackageVersions(apiURL string, config *stamus.Config) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Set user agent (required by GitHub API)
	req.Header.Set("User-Agent", "stamusctl")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Add GitHub token if available (check for common environment variables)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	} else if token := os.Getenv("GH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	// Could also check config for stored GitHub credentials in the future

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// GitHub Packages API returns an array of version objects
	var versions []struct {
		Name string `json:"name"`
		Metadata struct {
			Container struct {
				Tags []string `json:"tags"`
			} `json:"container"`
		} `json:"metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}

	var tags []string
	for _, version := range versions {
		// Add all tags from this version
		tags = append(tags, version.Metadata.Container.Tags...)
	}

	return tags, nil
}

func fetchDockerRegistryTags(url string, config *stamus.Config, registryURL string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication if available
	if config != nil && config.Registries != nil {
		if logins, exists := config.Registries[stamus.Registry(registryURL)]; exists {
			// Use the first available login for this registry
			for user, token := range logins {
				req.SetBasicAuth(string(user), string(token))
				break
			}
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var tagsResp RegistryTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, err
	}

	return tagsResp.Tags, nil
}

func isTemplateCachedLocally(templateName, version string) bool {
	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}
	defer cli.Close()

	// Check if image exists locally
	registryURL := app.DefaultRegistry
	expectedImage := fmt.Sprintf("%s/%s:%s", registryURL, templateName, version)

	images, err := cli.ImageList(context.Background(), image.ListOptions{All: true})
	if err != nil {
		return false
	}

	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			if repoTag == expectedImage {
				return true
			}
		}
	}

	return false
}

func isTemplateAvailableLocally(templateName, version string, localTemplates []TemplateInfo) bool {
	// Check if this template version exists in local templates
	for _, local := range localTemplates {
		if local.Name == templateName {
			// Handle cases where local version might be "1.0.0 (latest)" and we're looking for "1.0.0"
			localVersion := local.Version
			if strings.Contains(localVersion, " (latest)") {
				localVersion = strings.Split(localVersion, " (latest)")[0]
			}

			if localVersion == version {
				return true
			}

			// Also check if remote version matches a "latest" tag
			if local.Version == "latest" && version != "latest" {
				if resolvedVersion, err := resolveLatestVersion(templateName); err == nil && resolvedVersion == version {
					return true
				}
			}
		}
	}

	// Also check if it's cached as a Docker image
	return isTemplateCachedLocally(templateName, version)
}

func resolveLatestVersion(templateName string) (string, error) {
	// Create Docker client to inspect the latest image
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "latest", err // fallback to "latest" if Docker unavailable
	}
	defer cli.Close()

	registryURL := app.DefaultRegistry
	latestImage := fmt.Sprintf("%s/%s:latest", registryURL, templateName)

	// Get image inspect info for latest tag
	latestInspect, _, err := cli.ImageInspectWithRaw(context.Background(), latestImage)
	if err != nil {
		return "latest", err // Image not found locally, fallback to "latest"
	}

	// Get all images for this template to compare
	images, err := cli.ImageList(context.Background(), image.ListOptions{All: true})
	if err != nil {
		return "latest", err
	}

	// Look for images with the same digest/ID as the latest image
	for _, img := range images {
		if img.ID == latestInspect.ID {
			for _, repoTag := range img.RepoTags {
				// Skip the latest tag itself, look for the actual version
				if strings.HasPrefix(repoTag, registryURL+"/"+templateName+":") &&
					!strings.HasSuffix(repoTag, ":latest") {
					// Extract version from tag
					version := strings.TrimPrefix(repoTag, registryURL+"/"+templateName+":")
					return version, nil
				}
			}
		}
	}

	// If we can't find a matching local image, try to determine from remote releases
	// This assumes "latest" corresponds to the most recent non-prerelease version
	return resolveLatestFromReleases(templateName)
}

func resolveLatestFromReleases(templateName string) (string, error) {
	// Get GitHub releases to determine what "latest" should be
	releases, err := fetchGitHubReleases("stamusnetworks", "stamusctl-templates", []string{templateName})
	if err != nil || len(releases) == 0 {
		return "latest", err
	}

	// Find the most recent non-prerelease version
	// GitHub API returns releases in reverse chronological order (newest first)
	for _, release := range releases {
		version := release.Version
		// Skip prereleases (versions containing "-")
		if !strings.Contains(version, "-") {
			return version, nil
		}
	}

	// If no stable release found, return the first one
	return releases[0].Version, nil
}

func mergeDuplicateVersions(rows []table.Row) []table.Row {
	type VersionKey struct {
		template string
		version  string
	}

	// Map to track merged entries
	merged := make(map[VersionKey]*table.Row)
	var result []table.Row

	for _, row := range rows {
		template := row[0].(string)
		version := row[1].(string)
		source := row[2].(string)
		status := row[3].(string)

		// Extract base version (remove " (latest)" if present)
		baseVersion := version
		isLatest := false
		if strings.Contains(version, " (latest)") {
			baseVersion = strings.Split(version, " (latest)")[0]
			isLatest = true
		}

		key := VersionKey{template: template, version: baseVersion}

		if existing, exists := merged[key]; exists {
			// Merge with existing entry
			existingVersion := (*existing)[1].(string)
			existingSource := (*existing)[2].(string)
			existingStatus := (*existing)[3].(string)

			// Determine the merged version name
			mergedVersion := baseVersion
			if isLatest || strings.Contains(existingVersion, " (latest)") {
				mergedVersion = baseVersion + " (latest)"
			}

			// Merge sources - prioritize local filesystem, then show remote info
			mergedSource := existingSource
			if source == "local filesystem" || existingSource == "local filesystem" {
				mergedSource = "local filesystem"
			} else {
				// Both are remote, keep existing
				mergedSource = existingSource
			}

			// Merge status - prioritize "ready" over "cached (ready)" over "download needed"
			mergedStatus := existingStatus
			if status == "ready" || existingStatus == "ready" {
				mergedStatus = "ready"
			} else if status == "cached (ready)" || existingStatus == "cached (ready)" {
				mergedStatus = "cached (ready)"
			}

			// Update the existing entry
			(*existing)[1] = mergedVersion
			(*existing)[2] = mergedSource
			(*existing)[3] = mergedStatus
		} else {
			// New entry
			newRow := table.Row{template, version, source, status}
			merged[key] = &newRow
			result = append(result, newRow)
		}
	}

	return result
}

func fetchGitHubReleases(owner, repoName string, templateNames []string) ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// Use GitHub Releases API - completely public, no auth required
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repoName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Set user agent (recommended by GitHub API)
	req.Header.Set("User-Agent", "stamusctl")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// GitHub Releases API returns an array of release objects
	var releases []struct {
		TagName string `json:"tag_name"`
		Draft   bool   `json:"draft"`
		Prerelease bool `json:"prerelease"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	// For each template name, add all release versions
	for _, templateName := range templateNames {
		for _, release := range releases {
			// Skip draft releases
			if release.Draft {
				continue
			}

			templates = append(templates, TemplateInfo{
				Name:    templateName,
				Version: release.TagName,
				Source:  "github.com/StamusNetworks",
				Status:  "download needed",
			})
		}
	}

	return templates, nil
}
