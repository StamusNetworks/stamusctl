package compose

import (
	"strings"
	"unicode"

	"stamus-ctl/internal/docker"
	"stamus-ctl/internal/logging"
)

func RetrieveValideInterfacesFromDockerContainer() ([]string, error) {
	alreadyHasBusybox, _ := docker.PullImageIfNotExisted("docker.io/library/", "busybox:latest")

	output, _ := docker.RunContainer("busybox", []string{
		"ls",
		"/sys/class/net",
	}, nil, "host")

	if !alreadyHasBusybox {
		logging.Sugar.Debug("busybox image was not previously installed, deleting.")
		docker.DeleteDockerImageByName("", "busybox:latest")
	}

	interfaces := strings.Split(output, "\n")
	interfaces = interfaces[:len(interfaces)-1]
	for i, in := range interfaces {
		in = strings.TrimFunc(in, unicode.IsControl)
		interfaces[i] = in
	}
	logging.Sugar.Debugw("detected interfaces.", "interfaces", interfaces)

	return interfaces, nil
}
