package docker

import (
	"bytes"
	"fmt"

	"stamus-ctl/internal/logging"

	"github.com/docker/docker/api/types/image"
)

func PullImageIfNotExisted(registry string, name string) (bool, error) {
	// name = name + ":main"
	logger := logging.Sugar.With("registry", registry, "name", name)

	logger.Debug("pulling image")

	alreadyHere, err := IsImageAlreadyInstalled(registry, name)
	logger.Debug("alreadyHere: ", alreadyHere)
	if err != nil {
		logger.Debugw("image failed to test", "error", err)
		return true, err
	}
	if alreadyHere {
		logger.Debugw("image found")
		return true, nil
	}

	logger.Debugw("image not found")

	s := logging.NewSpinner(
		fmt.Sprintf("Pulling %s. Please wait", name),
		fmt.Sprintf("Pulling %s done\n", name),
	)

	reader, err := cli.ImagePull(ctx, registry+name, image.PullOptions{})

	if err != nil {
		logging.SpinnerStop(s)
		logger.Debugw("image failed to pull", "error", err)
		return false, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	logging.SpinnerStop(s)
	logger.Debugw("image dl", "error", err)
	return false, nil
}
