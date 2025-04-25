package utils

import (
	"strings"

	"stamus-ctl/internal/logging"
)

func ExtractArgs(args []string) map[string]string {
	paramsArgs := make(map[string]string)
	for _, arg := range args {
		splited := strings.Split(arg, "=")
		if len(splited) != 2 {
			logging.Sugar.Info("Error: invalid argument", arg)
		} else {
			paramsArgs[splited[0]] = splited[1]
		}
	}
	return paramsArgs
}
