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

	var alreadyHere bool
	err := WithRetrySimple(func() error {
		var checkErr error
		alreadyHere, checkErr = IsImageAlreadyInstalled(registry, name)
		return checkErr
	}, fmt.Sprintf("check-image-exists-%s", name))

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

	err = WithRetrySimple(func() error {
		reader, pullErr := cli.ImagePull(ctx, registry+name, image.PullOptions{})
		if pullErr != nil {
			return pullErr
		}

		buf := new(bytes.Buffer)
		_, readErr := buf.ReadFrom(reader)
		if readErr != nil {
			return readErr
		}
		reader.Close()

		return nil
	}, fmt.Sprintf("pull-image-%s", name))

	logging.SpinnerStop(s)

	if err != nil {
		logger.Debugw("image failed to pull", "error", err)
		return false, err
	}

	logger.Debugw("image dl", "error", err)
	return false, nil
}
