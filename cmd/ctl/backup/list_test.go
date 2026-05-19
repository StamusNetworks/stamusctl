package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatSize_Zero(t *testing.T) {
	assert.Equal(t, "0 B", formatSize(0))
}

func TestFormatSize_SubKilo(t *testing.T) {
	assert.Equal(t, "1 B", formatSize(1))
	assert.Equal(t, "512 B", formatSize(512))
	assert.Equal(t, "1023 B", formatSize(1023))
}

func TestFormatSize_Kilobytes(t *testing.T) {
	assert.Equal(t, "1.0 KB", formatSize(1024))
	assert.Equal(t, "1.5 KB", formatSize(1536))
	assert.Equal(t, "1023.0 KB", formatSize(1023*1024))
}

func TestFormatSize_Megabytes(t *testing.T) {
	assert.Equal(t, "1.0 MB", formatSize(1024*1024))
	assert.Equal(t, "2.0 MB", formatSize(2*1024*1024))
}

func TestFormatSize_Gigabytes(t *testing.T) {
	assert.Equal(t, "1.0 GB", formatSize(1024*1024*1024))
}

func TestFormatSize_Terabytes(t *testing.T) {
	assert.Equal(t, "1.0 TB", formatSize(1024*1024*1024*1024))
}
