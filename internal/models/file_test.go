package models

import (
	"fmt"
	"testing"
)

func isEqualFiles(file1, file2 *File) bool {
	if file1 == nil && file2 == nil {
		return true
	}
	if file1.Path != file2.Path {
		return false
	}
	if file1.Name != file2.Name {
		return false
	}
	if file1.Type != file2.Type {
		return false
	}
	return true
}

func TestIsValidPath(t *testing.T) {
	tests := []struct {
		file     *File
		expected error
	}{
		{&File{Path: ".", Name: "testfile", Type: "txt"}, nil},
		{&File{Path: "", Name: "testfile", Type: "txt"}, nil},
		{&File{Path: ".", Name: "", Type: "txt"}, fmt.Errorf("type txt is not valid")},
		{&File{Path: ".", Name: "testfile", Type: ""}, fmt.Errorf("type  is not valid")},
	}

	for _, test := range tests {
		err := test.file.isValidPath()
		if err != nil && err.Error() != test.expected.Error() {
			t.Errorf("Expected error %v, but got %v", test.expected, err)
		}
		if err == nil && test.expected != nil {
			t.Errorf("Expected error %v, but got nil", test.expected)
		}
	}
}

func TestCompletePath(t *testing.T) {
	file := File{Path: ".", Name: "testfile", Type: "txt"}
	expected := "./testfile.txt"
	result := file.completePath()
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestCreateFileFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected *File
		hasError bool
	}{
		{"/home/user/testfile.txt", &File{Path: "/home/user", Name: "testfile", Type: "txt"}, false},
		{"testfile.txt", &File{Path: ".", Name: "testfile", Type: "txt"}, false},
		{"invalidfile", nil, true},
	}

	for _, test := range tests {
		result, err := CreateFileFromPath(test.path)
		if (err != nil) != test.hasError {
			t.Errorf("Expected error: %v, but got: %v", test.hasError, err)
		}
		if !isEqualFiles(result, test.expected) {
			t.Errorf("Expected %v, but got %v", test.expected, result)
		}
	}
}

func TestCreateFile(t *testing.T) {
	tests := []struct {
		path     string
		fileName string
		expected *File
		hasError bool
	}{
		{"/home/user", "testfile.txt", &File{Path: "/home/user", Name: "testfile", Type: "txt"}, false},
		{".", "testfile.txt", &File{Path: ".", Name: "testfile", Type: "txt"}, false},
		{"/home/user", "invalidfile", nil, true},
	}

	for _, test := range tests {
		result, err := CreateFile(test.path, test.fileName)
		if (err != nil) != test.hasError {
			t.Errorf("Expected error: %v, but got: %v", test.hasError, err)
		}
		if !isEqualFiles(result, test.expected) {
			t.Errorf("Expected %v, but got %v", test.expected, result)
		}
	}
}
