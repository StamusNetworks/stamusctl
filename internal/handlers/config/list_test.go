package config

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"stamus-ctl/internal/stamus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatStatus_Up(t *testing.T) {
	tests := []struct {
		name       string
		infos      stamus.Infos
		wantSubstr []string
	}{
		{
			name: "up with containers",
			infos: stamus.Infos{
				Status: stamus.StatusUp,
				Containers: stamus.ContainerStatus{
					Running: 3,
					Total:   3,
				},
			},
			wantSubstr: []string{"up", "(3/3)"},
		},
		{
			name: "up with single container",
			infos: stamus.Infos{
				Status: stamus.StatusUp,
				Containers: stamus.ContainerStatus{
					Running: 1,
					Total:   1,
				},
			},
			wantSubstr: []string{"up", "(1/1)"},
		},
		{
			name: "up with zero total containers",
			infos: stamus.Infos{
				Status:     stamus.StatusUp,
				Containers: stamus.ContainerStatus{Running: 0, Total: 0},
			},
			// No count string when Total == 0
			wantSubstr: []string{"up"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatus(tt.infos)
			for _, sub := range tt.wantSubstr {
				assert.Contains(t, result, sub)
			}
			assert.Contains(t, result, Green)
			assert.Contains(t, result, Reset)
		})
	}
}

func TestFormatStatus_Partial(t *testing.T) {
	tests := []struct {
		name  string
		infos stamus.Infos
	}{
		{
			name: "partial with some running",
			infos: stamus.Infos{
				Status: stamus.StatusPartial,
				Containers: stamus.ContainerStatus{
					Running: 2,
					Total:   4,
				},
			},
		},
		{
			name: "partial with one running",
			infos: stamus.Infos{
				Status: stamus.StatusPartial,
				Containers: stamus.ContainerStatus{
					Running: 1,
					Total:   5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatus(tt.infos)
			assert.Contains(t, result, "partial")
			assert.Contains(t, result, Yellow)
			assert.Contains(t, result, Reset)
			cs := tt.infos.Containers
			if cs.Total > 0 {
				assert.Contains(t, result, "(")
				assert.Contains(t, result, "/")
			}
		})
	}
}

func TestFormatStatus_Unhealthy(t *testing.T) {
	tests := []struct {
		name  string
		infos stamus.Infos
	}{
		{
			name: "unhealthy with containers",
			infos: stamus.Infos{
				Status: stamus.StatusUnhealthy,
				Containers: stamus.ContainerStatus{
					Running:   3,
					Total:     3,
					Unhealthy: 1,
				},
			},
		},
		{
			name: "unhealthy all containers bad",
			infos: stamus.Infos{
				Status: stamus.StatusUnhealthy,
				Containers: stamus.ContainerStatus{
					Running:   2,
					Total:     2,
					Unhealthy: 2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatus(tt.infos)
			assert.Contains(t, result, "unhealthy")
			assert.Contains(t, result, Red)
			assert.Contains(t, result, Reset)
		})
	}
}

func TestFormatStatus_Down(t *testing.T) {
	tests := []struct {
		name  string
		infos stamus.Infos
	}{
		{
			name: "down with no containers",
			infos: stamus.Infos{
				Status:     stamus.StatusDown,
				Containers: stamus.ContainerStatus{Running: 0, Total: 0},
			},
		},
		{
			name: "unknown status treated as down",
			infos: stamus.Infos{
				Status:     stamus.Status("unknown"),
				Containers: stamus.ContainerStatus{Running: 0, Total: 5},
			},
		},
		{
			name: "down status with total > 0 shows count",
			infos: stamus.Infos{
				Status:     stamus.StatusDown,
				Containers: stamus.ContainerStatus{Running: 0, Total: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatus(tt.infos)
			assert.Contains(t, result, "down")
			// Down does not have ANSI color codes
			assert.NotContains(t, result, Green)
			assert.NotContains(t, result, Yellow)
			assert.NotContains(t, result, Red)
		})
	}
}

func TestFormatStatus_NoBracketsWhenZeroTotal(t *testing.T) {
	infos := stamus.Infos{
		Status:     stamus.StatusUp,
		Containers: stamus.ContainerStatus{Running: 0, Total: 0},
	}
	result := formatStatus(infos)
	assert.NotContains(t, result, "(")
	assert.NotContains(t, result, "/")
}

func TestListHandler_Success(t *testing.T) {
	// Save and replace package-level vars
	origGetInstances := getInstances
	origWriter := outputWriter
	defer func() {
		getInstances = origGetInstances
		outputWriter = origWriter
	}()

	// Mock getInstances to return controlled data
	mockInstances := stamus.Instances{
		stamus.Folder("/configs/default"): stamus.Infos{
			Project:    "clearndr",
			Version:    "1.2.3",
			Status:     stamus.StatusUp,
			Containers: stamus.ContainerStatus{Running: 2, Total: 2},
		},
		stamus.Folder("/configs/second"): stamus.Infos{
			Project:    "other-project",
			Version:    "0.9.0",
			Status:     stamus.StatusDown,
			Containers: stamus.ContainerStatus{Running: 0, Total: 0},
		},
	}
	getInstances = func() (stamus.Instances, error) {
		return mockInstances, nil
	}

	// Capture output
	var buf bytes.Buffer
	outputWriter = &buf

	err := ListHandler()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "clearndr")
	assert.Contains(t, output, "1.2.3")
	assert.Contains(t, output, "other-project")
	assert.Contains(t, output, "0.9.0")
}

func TestListHandler_EmptyInstances(t *testing.T) {
	origGetInstances := getInstances
	origWriter := outputWriter
	defer func() {
		getInstances = origGetInstances
		outputWriter = origWriter
	}()

	getInstances = func() (stamus.Instances, error) {
		return stamus.Instances{}, nil
	}

	var buf bytes.Buffer
	outputWriter = &buf

	err := ListHandler()
	require.NoError(t, err)

	// Table headers should still be present
	output := buf.String()
	assert.Contains(t, output, "LOCATION")
	assert.Contains(t, output, "PROJECT")
	assert.Contains(t, output, "VERSION")
	assert.Contains(t, output, "STATUS")
}

func TestListHandler_GetInstancesError(t *testing.T) {
	origGetInstances := getInstances
	origWriter := outputWriter
	defer func() {
		getInstances = origGetInstances
		outputWriter = origWriter
	}()

	sentinel := errors.New("docker daemon unavailable")
	getInstances = func() (stamus.Instances, error) {
		return nil, sentinel
	}

	var buf bytes.Buffer
	outputWriter = &buf

	err := ListHandler()
	assert.ErrorIs(t, err, sentinel)
	// Nothing should have been written
	assert.Empty(t, buf.String())
}

func TestListHandler_TableColumns(t *testing.T) {
	origGetInstances := getInstances
	origWriter := outputWriter
	defer func() {
		getInstances = origGetInstances
		outputWriter = origWriter
	}()

	getInstances = func() (stamus.Instances, error) {
		return stamus.Instances{
			stamus.Folder("/data/myconfig"): stamus.Infos{
				Project:    "my-project",
				Version:    "v2.0.0",
				Status:     stamus.StatusPartial,
				Containers: stamus.ContainerStatus{Running: 1, Total: 3},
			},
		}, nil
	}

	var buf bytes.Buffer
	outputWriter = &buf

	err := ListHandler()
	require.NoError(t, err)

	output := buf.String()
	// Verify all four columns appear
	for _, col := range []string{"LOCATION", "PROJECT", "VERSION", "STATUS"} {
		assert.True(t, strings.Contains(output, col), "expected column %q in output", col)
	}
	assert.Contains(t, output, "/data/myconfig")
	assert.Contains(t, output, "my-project")
	assert.Contains(t, output, "v2.0.0")
}
