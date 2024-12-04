package utils

import (
	"fmt"
	"log"
	"os"

	cp "github.com/otiai10/copy"
)

func Copy(inputPath string, outputPath string) error {
	fmt.Println("Setting content from ", inputPath, " to ", outputPath)
	// Check input path exists
	info, err := os.Stat(inputPath)
	if err != nil {
		log.Println(info, err)
		return fmt.Errorf("input path does not exist: %s", inputPath)
	}

	err = cp.Copy(inputPath, outputPath)
	if err != nil {
		return err
	}

	return nil
}
