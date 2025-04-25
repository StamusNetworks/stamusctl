package docker

import (
	"errors"
	"slices"

	"stamus-ctl/internal/logging"

	"github.com/docker/docker/api/types/image"
)

func ImageName(image image.Summary) string {
	logger := logging.Sugar.With("RepoTags", image.RepoTags)
	if len(image.RepoTags) == 0 {
		logger.Debug("no tag found")
		return "none"
	}

	return image.RepoTags[0]
}

func GetImagesName(images []image.Summary) []string {
	var names []string
	for _, image := range images {
		name := ImageName(image)
		names = append(names, name)
	}

	return names
}

func GetInstalledImagesName() ([]string, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	names := GetImagesName(images)

	return names, nil
}

func IsImageAlreadyInstalled(registry, name string) (bool, error) {
	logger := logging.Sugar.With("registry", registry, "name", name, "location", "IsImageAlreadyInstalled")

	logger.Debug("searching for image")
	images, err := GetInstalledImagesName()
	if err != nil {
		return false, err
	}

	logger.Debug(images)

	if registry == "docker.io/library/" {
		logger.Debug("skipping registry as it is docker")
		return slices.Contains(images, name), err
	}

	return slices.Contains(images, registry+name), err
}

func GetImageIdFromName(registry, name string) (string, error) {
	logger := logging.Sugar.With("registry", registry, "name", name, "location", "GetImageIdFromName")

	logger.Debug("searching for imageID")
	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	for _, image := range images {
		shortName := ImageName(image)

		if registry == "docker.io/library/" && shortName == name {
			logging.Sugar.Debugw("image name found", "image.ID", image.ID, "shortName", shortName, "name", name)
			return image.ID, nil
		}

		if shortName == registry+name {
			logging.Sugar.Debugw("image name found", "image.ID", image.ID, "shortName", shortName, "name", name)
			return image.ID, nil
		}
	}

	logging.Sugar.Debugw("image not found", " name ", name)
	return "", errors.New("image not found")
}
