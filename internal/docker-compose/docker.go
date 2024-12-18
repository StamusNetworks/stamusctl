package compose

import (
	"os"
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

func GenerateSSLWithDocker(sslPath string) error {
	logging.Sugar.Debugw("Generating ssl cert.", "path", sslPath)

	err := os.MkdirAll(sslPath, os.ModePerm)
	if err != nil {
		logging.Sugar.Errorw("cannot create cert containing folder.", "error", err)
	}

	alreadyHasNginx, err := docker.PullImageIfNotExisted("docker.io/library/", "nginx:1.27")
	if err != nil {
		logging.Sugar.Warnw("nginx pull failed", "error", err)
		return err
	}

	_, err = docker.RunContainer("nginx", []string{
		"openssl",
		"req", "-new", "-nodes", "-x509",
		"-subj", "/C=FR/ST=IDF/L=Paris/O=Stamus/CN=ClearNDR",
		"-days", "3650",
		"-keyout", "/etc/nginx/ssl/scirius.key",
		"-out", "/etc/nginx/ssl/scirius.crt",
		"-extensions", "v3_ca",
	}, []string{
		sslPath + ":/etc/nginx/ssl",
	}, "")

	if err != nil {
		logging.Sugar.Infow("cannot generate cert.", "error", err)
		return err
	}

	if !alreadyHasNginx {
		logging.Sugar.Debug("nginx image was not previously installed, deleting.")
		docker.DeleteDockerImageByName("", "nginx")
	}

	logging.Sugar.Debug("cert created.", "path", sslPath)
	return nil
}
