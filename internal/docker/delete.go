package docker

import (
	"stamus-ctl/internal/logging"

	"github.com/docker/docker/api/types/image"
)

func DeleteDockerImageByName(registry, name string) (bool, error) {
	id, err := GetImageIdFromName(registry, name)
	if err != nil {
		logging.Sugar.Warnw("image id not found", "error", err)
		return false, err
	}

	if _, err = cli.ImageRemove(ctx, id, image.RemoveOptions{}); err != nil {
		return false, err
	}

	return true, nil
}
