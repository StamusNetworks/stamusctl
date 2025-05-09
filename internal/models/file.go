package models

import (
	// Common

	"fmt"
	"log"
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"

	"github.com/spf13/viper"
	// External
)

// Used to setup viper instances
type File struct {
	Path          string
	Name          string
	Type          string
	viperInstance *viper.Viper
}

func NewFile(path, name, fileType string) *File {
	return &File{
		Path: path,
		Name: name,
		Type: fileType,
	}
}

func (f *File) InstanciateViper() (*viper.Viper, error) {
	viperInstance, err := InstanciateViper(f)
	if err != nil {
		return nil, err
	}
	f.viperInstance = viperInstance
	return viperInstance, nil
}

// Used to get the file as properties from path
func CreateFileFromPath(path string) (*File, error) {
	// Extract the file properties
	pathSplited := strings.Split(path, "/")
	if len(pathSplited) < 2 {
		pathSplited = []string{".", pathSplited[0]}
	}
	nameSplited := strings.Split(pathSplited[len(pathSplited)-1], ".")
	// Validate name
	if len(nameSplited) < 2 {
		return nil, fmt.Errorf("path %s is not a valid file name", path)
	}
	// File
	file := NewFile(
		strings.Join(pathSplited[:len(pathSplited)-1], "/"),
		strings.Join(nameSplited[:len(nameSplited)-1], "."),
		nameSplited[len(nameSplited)-1],
	)
	// Validate all
	err := file.isValidPath()
	if err != nil {
		return nil, err
	}
	// Return file instance
	return file, nil
}

// Used create a file from path and name
func CreateFile(path string, fileName string) (*File, error) {
	// Extract the file properties
	nameSplited := strings.Split(fileName, ".")
	if len(nameSplited) != 2 {
		return nil, fmt.Errorf("path %s is not a valid file name", path)
	}
	// File
	file := NewFile(
		path,
		nameSplited[0],
		nameSplited[1],
	)

	// Validate all
	err := file.isValidPath()
	if err != nil {
		return nil, err
	}

	// Return file instance
	return file, nil
}

func (f *File) completePath() string {
	return f.Path + "/" + f.Name + "." + f.Type
}

func (f *File) GetViper() *viper.Viper {
	if f.viperInstance == nil {
		_, err := f.InstanciateViper()
		if err != nil {
			log.Println("Error instanciating viper", err)
			logging.Sugar.Error("Error instanciating viper", err)
			return nil
		}
	}
	return f.viperInstance
}

// Empirical function to check if a path is valid
func (f *File) isValidPath() error {
	// Check if file already exists
	if _, err := app.FS.Stat(f.completePath()); err == nil {
		return nil
	}

	// Check parts
	if f.Path == "" {
		f.Path = "."
	}
	if f.Name == "" || f.Type == "" {
		return fmt.Errorf("type %s is not valid", f.Type)
	}

	// Return error if not possible
	return nil
}
