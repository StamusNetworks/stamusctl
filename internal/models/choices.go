package models

import (
	"io/ioutil"
	"log"
	"strings"
	"time"
	"unicode"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/docker"
	"stamus-ctl/internal/logging"
)

// Get the choices for a given variable
// Returns a list of choices given the variable name
func GetChoices(name string) ([]Variable, error) {
	switch name {
	case "restart":
		return []Variable{
			CreateVariableString("no"),
			CreateVariableString("always"),
			CreateVariableString("on-failure"),
			CreateVariableString("unless-stopped"),
		}, nil
	case "nginx":
		return []Variable{
			CreateVariableString("nginx"),
			CreateVariableString("nginx-exec"),
		}, nil
	case "interfaces":
		return getInterfaces()
	default:
		return nil, nil
	}
}

// DisableInterfaceDetection skips host/container network interface detection
// when resolving the "interfaces" choices. It is enabled by `nix init`, where
// the host running the command is typically not the host that will run the
// stack, so probing local interfaces is misleading.
var DisableInterfaceDetection = false

// Get the list of network interfaces
// Depending on the mode (prod or test), it will either use the host or a busybox container
func getInterfaces() ([]Variable, error) {
	// When interface detection is disabled (e.g. `nix init`), return no choices
	// so the interface is provided as free text rather than picked from this host.
	if DisableInterfaceDetection {
		return nil, nil
	}
	go func() {
		time.Sleep(1 * time.Minute)
		interfacesCache = []Variable{}
	}()
	if app.Mode.IsProd() {
		return getInterfacesBusybox()
	} else {
		return getInterfacesHost()
	}
}

// Get the list of network interfaces using the host
func getInterfacesHost() ([]Variable, error) {
	// Define the directory where network interfaces are listed
	if len(interfacesCache) != 0 {
		return interfacesCache, nil
	}

	// Read the directory contents
	netDir := "/sys/class/net"
	files, err := ioutil.ReadDir(netDir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", netDir, err)
	}

	// Loop through the files and print the names
	interfaces := []Variable{}
	for _, file := range files {
		interfaces = append(interfaces, CreateVariableString(file.Name()))
	}
	interfacesCache = interfaces

	return interfaces, nil
}

var interfacesCache []Variable

// Get the list of network interfaces using a busybox container
func getInterfacesBusybox() ([]Variable, error) {
	if len(interfacesCache) != 0 {
		return interfacesCache, nil
	}

	s := logging.NewSpinner(
		"Identifying interfaces",
		"",
	)

	_, err := docker.PullImageIfNotExisted("docker.io/library/", "busybox:latest")
	if err != nil {
		logging.SpinnerStop(s)
		return getInterfacesHost()
	}

	output, _ := docker.RunContainer("busybox:latest", []string{
		"ls",
		"/sys/class/net",
	}, nil, "host")

	interfaces := strings.Split(output, "\n")
	interfaces = interfaces[:len(interfaces)-1]
	for i, in := range interfaces {
		in = strings.TrimFunc(in, unicode.IsControl)
		interfaces[i] = in
	}
	logging.Sugar.Debugw("detected interfaces.", "interfaces", interfaces)

	interfacesVariables := []Variable{}
	for _, in := range interfaces {
		if in != "" {
			interfacesVariables = append(interfacesVariables, CreateVariableString(in))
		}
	}

	logging.SpinnerStop(s)
	interfacesCache = interfacesVariables

	return interfacesVariables, nil
}
