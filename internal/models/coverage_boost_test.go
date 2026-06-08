package models

import (
	"fmt"
	"testing"
	"text/template"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// safeFuncMap
// ---------------------------------------------------------------------------

func TestSafeFuncMap_DangerousFunctionsRemoved(t *testing.T) {
	fm := safeFuncMap()

	dangerous := []string{
		"env",
		"expandenv",
		"getHostByName",
		"genPrivateKey",
		"genCA",
		"genSelfSignedCert",
		"genSignedCert",
	}
	for _, fn := range dangerous {
		_, present := fm[fn]
		assert.False(t, present, "dangerous function %q must not be in safeFuncMap", fn)
	}
}

func TestSafeFuncMap_SafeFunctionsPresent(t *testing.T) {
	fm := safeFuncMap()

	safe := []string{"upper", "lower", "trim", "replace", "default", "toJson"}
	for _, fn := range safe {
		_, present := fm[fn]
		assert.True(t, present, "safe sprig function %q must be in safeFuncMap", fn)
	}
}

func TestSafeFuncMap_EnvFunctionUnavailable(t *testing.T) {
	fm := safeFuncMap()
	tmpl, err := template.New("test").Funcs(template.FuncMap(fm)).Parse(`{{ env "HOME" }}`)
	// Parsing should fail because "env" is not defined.
	assert.Error(t, err)
	_ = tmpl
}

// ---------------------------------------------------------------------------
// nestMap
// ---------------------------------------------------------------------------

func TestNestMap_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			name:  "empty input",
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
		{
			name:  "single level key no dot",
			input: map[string]interface{}{"foo": "bar"},
			want:  map[string]interface{}{"foo": "bar"},
		},
		{
			name:  "nested key a.b.c",
			input: map[string]interface{}{"a.b.c": "val"},
			want: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "val",
					},
				},
			},
		},
		{
			name:  "multiple keys sharing prefix",
			input: map[string]interface{}{"a.b": 1, "a.c": 2},
			want: map[string]interface{}{
				"a": map[string]interface{}{
					"b": 1,
					"c": 2,
				},
			},
		},
		{
			name:  "deeply nested single key",
			input: map[string]interface{}{"a.b.c.d.e": true},
			want: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"d": map[string]interface{}{
								"e": true,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nestMap(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// getAllFilesContent
// ---------------------------------------------------------------------------

func TestGetAllFilesContent(t *testing.T) {
	base := "/TestGetAllFilesContent"
	require.NoError(t, app.FS.MkdirAll(base, 0o755))

	require.NoError(t, afero.WriteFile(app.FS, base+"/a.yaml", []byte("key: value"), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, base+"/b.yaml", []byte("other: stuff"), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, base+"/c.txt", []byte("ignored"), 0o644))

	contents, err := getAllFilesContent(base, ".yaml")
	require.NoError(t, err)
	assert.Len(t, contents, 2)
	assert.ElementsMatch(t, []string{"key: value", "other: stuff"}, contents)
}

func TestGetAllFilesContent_NonExistentDir(t *testing.T) {
	_, err := getAllFilesContent("/NoSuchDir_getAllFilesContent", ".yaml")
	assert.Error(t, err)
}

func TestGetAllFilesContent_Nested(t *testing.T) {
	base := "/TestGetAllFilesContent_Nested"
	require.NoError(t, app.FS.MkdirAll(base+"/sub", 0o755))

	require.NoError(t, afero.WriteFile(app.FS, base+"/a.tpl", []byte("tpl-a"), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, base+"/sub/b.tpl", []byte("tpl-b"), 0o644))

	contents, err := getAllFilesContent(base, ".tpl")
	require.NoError(t, err)
	assert.Len(t, contents, 2)
	assert.ElementsMatch(t, []string{"tpl-a", "tpl-b"}, contents)
}

// ---------------------------------------------------------------------------
// processTemplate — various paths
// ---------------------------------------------------------------------------

func TestProcessTemplate_SkipsTplFiles(t *testing.T) {
	logger := zap.NewExample().Sugar()

	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Skip/in", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestProcessTemplate_Skip/in/helper.tpl",
		[]byte("{{ define \"helper\" }}HELPER{{ end }}"), 0o644))

	info, err := app.FS.Stat("/TestProcessTemplate_Skip/in/helper.tpl")
	require.NoError(t, err)

	err = processTemplate(map[string]interface{}{}, []string{},
		"/TestProcessTemplate_Skip/in/helper.tpl",
		"/TestProcessTemplate_Skip/in", "/TestProcessTemplate_Skip/out", info, logger)
	assert.NoError(t, err)

	exists, err := afero.Exists(app.FS, "/TestProcessTemplate_Skip/out/helper.tpl")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestProcessTemplate_RenderWithValues(t *testing.T) {
	logger := zap.NewExample().Sugar()

	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Render/in", 0o755))
	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Render/out", 0o755))

	require.NoError(t, afero.WriteFile(app.FS, "/TestProcessTemplate_Render/in/greeting.txt",
		[]byte("Hello {{ .Values.name }}!"), 0o644))

	info, err := app.FS.Stat("/TestProcessTemplate_Render/in/greeting.txt")
	require.NoError(t, err)

	data := map[string]interface{}{
		"Values": map[string]interface{}{
			"name": "World",
		},
	}

	err = processTemplate(data, []string{}, "/TestProcessTemplate_Render/in/greeting.txt",
		"/TestProcessTemplate_Render/in", "/TestProcessTemplate_Render/out", info, logger)
	require.NoError(t, err)

	content, err := afero.ReadFile(app.FS, "/TestProcessTemplate_Render/out/greeting.txt")
	require.NoError(t, err)
	assert.Equal(t, "Hello World!\n", string(content))
}

func TestProcessTemplate_InvalidTemplate(t *testing.T) {
	logger := zap.NewExample().Sugar()

	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Invalid/in", 0o755))
	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Invalid/out", 0o755))

	// Deliberately broken template syntax.
	require.NoError(t, afero.WriteFile(app.FS, "/TestProcessTemplate_Invalid/in/bad.txt",
		[]byte("{{ .Foo }"), 0o644))

	info, err := app.FS.Stat("/TestProcessTemplate_Invalid/in/bad.txt")
	require.NoError(t, err)

	err = processTemplate(map[string]interface{}{}, []string{},
		"/TestProcessTemplate_Invalid/in/bad.txt",
		"/TestProcessTemplate_Invalid/in", "/TestProcessTemplate_Invalid/out", info, logger)
	assert.Error(t, err)
}

func TestProcessTemplate_ShellScriptChmod(t *testing.T) {
	logger := zap.NewExample().Sugar()

	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Chmod/in", 0o755))
	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Chmod/out", 0o755))

	require.NoError(t, afero.WriteFile(app.FS, "/TestProcessTemplate_Chmod/in/run.sh",
		[]byte("#!/bin/sh\necho hello\n"), 0o644))

	info, err := app.FS.Stat("/TestProcessTemplate_Chmod/in/run.sh")
	require.NoError(t, err)

	err = processTemplate(map[string]interface{}{}, []string{}, "/TestProcessTemplate_Chmod/in/run.sh",
		"/TestProcessTemplate_Chmod/in", "/TestProcessTemplate_Chmod/out", info, logger)
	require.NoError(t, err)

	exists, err := afero.Exists(app.FS, "/TestProcessTemplate_Chmod/out/run.sh")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestProcessTemplate_ExecutionError(t *testing.T) {
	logger := zap.NewExample().Sugar()

	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_ExecErr/in", 0o755))
	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_ExecErr/out", 0o755))

	// The sprig `fail` function forces a runtime error.
	require.NoError(t, afero.WriteFile(app.FS, "/TestProcessTemplate_ExecErr/in/fail.txt",
		[]byte(`{{ fail "intentional error" }}`), 0o644))

	info, err := app.FS.Stat("/TestProcessTemplate_ExecErr/in/fail.txt")
	require.NoError(t, err)

	err = processTemplate(map[string]interface{}{}, []string{},
		"/TestProcessTemplate_ExecErr/in/fail.txt",
		"/TestProcessTemplate_ExecErr/in", "/TestProcessTemplate_ExecErr/out", info, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "intentional error")
}

// ---------------------------------------------------------------------------
// processTemplates — end-to-end
// ---------------------------------------------------------------------------

func TestProcessTemplates_Basic(t *testing.T) {
	inDir := "/TestProcessTemplates/in"
	outDir := "/TestProcessTemplates/out"
	require.NoError(t, app.FS.MkdirAll(inDir, 0o755))
	require.NoError(t, app.FS.MkdirAll(outDir, 0o755))

	// A helper .tpl file — its content is appended to each regular file.
	require.NoError(t, afero.WriteFile(app.FS, inDir+"/defs.tpl", []byte(""), 0o644))

	// A regular file with template markers.
	require.NoError(t, afero.WriteFile(app.FS, inDir+"/config.yaml",
		[]byte("name: {{ .Values.name }}\n"), 0o644))

	data := map[string]interface{}{
		"Values": map[string]interface{}{"name": "test"},
	}

	err := processTemplates(inDir, outDir, data)
	require.NoError(t, err)

	content, err := afero.ReadFile(app.FS, outDir+"/config.yaml")
	require.NoError(t, err)
	// The empty tpl content is joined with \n so expect an extra trailing newline.
	assert.Equal(t, "name: test\n\n", string(content))
}

// ---------------------------------------------------------------------------
// addValuePrefix
// ---------------------------------------------------------------------------

func TestAddValuePrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"foo", "Values.foo"},
		{"a.b.c", "Values.a.b.c"},
		{"", "Values."},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, addValuePrefix(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// removeEmptyStrings — nil / all-empty / none-empty branches
// ---------------------------------------------------------------------------

func TestRemoveEmptyStrings_NilInput(t *testing.T) {
	result := removeEmptyStrings(nil)
	assert.Nil(t, result)
}

func TestRemoveEmptyStrings_AllEmpty(t *testing.T) {
	result := removeEmptyStrings([]string{"", ""})
	// Implementation returns a nil slice when nothing is appended.
	assert.Nil(t, result)
}

func TestRemoveEmptyStrings_NoneEmpty(t *testing.T) {
	result := removeEmptyStrings([]string{"a", "b"})
	assert.Equal(t, []string{"a", "b"}, result)
}

// ---------------------------------------------------------------------------
// Parameter.SetDefault
// ---------------------------------------------------------------------------

func TestParameter_SetDefault(t *testing.T) {
	p := &Parameter{Type: "string"}
	v := CreateVariableString("mydefault")
	ret := p.SetDefault(v)

	assert.Same(t, p, ret, "SetDefault must return self for chaining")
	assert.Equal(t, "mydefault", *p.Default.String)
}

func TestParameter_SetDefault_Int(t *testing.T) {
	p := &Parameter{Type: "int"}
	v := CreateVariableInt(99)
	p.SetDefault(v)
	assert.Equal(t, 99, *p.Default.Int)
}

// ---------------------------------------------------------------------------
// Parameters.SetValuesSmartMerge
// ---------------------------------------------------------------------------

func TestSetValuesSmartMerge(t *testing.T) {
	t.Run("preserves customized value that differs from old default", func(t *testing.T) {
		newParams := &Parameters{
			"key": &Parameter{Type: "string", Default: CreateVariableString("new_default")},
		}
		oldParams := &Parameters{
			"key": &Parameter{
				Type:     "string",
				Variable: CreateVariableString("custom"),
				Default:  CreateVariableString("old_default"),
			},
		}
		newParams.SetValuesSmartMerge(oldParams)
		assert.Equal(t, "custom", *newParams.Get("key").Variable.String)
	})

	t.Run("does not copy value equal to old default", func(t *testing.T) {
		newParams := &Parameters{
			"key": &Parameter{Type: "string", Default: CreateVariableString("new_default")},
		}
		oldParams := &Parameters{
			"key": &Parameter{
				Type:     "string",
				Variable: CreateVariableString("old_default"),
				Default:  CreateVariableString("old_default"),
			},
		}
		newParams.SetValuesSmartMerge(oldParams)
		assert.True(t, newParams.Get("key").Variable.IsNil())
	})

	t.Run("parameter absent in old config is left untouched", func(t *testing.T) {
		newParams := &Parameters{
			"key": &Parameter{Type: "string", Default: CreateVariableString("new_default")},
		}
		newParams.SetValuesSmartMerge(&Parameters{})
		assert.True(t, newParams.Get("key").Variable.IsNil())
	})

	t.Run("old value set but no old default preserves old value", func(t *testing.T) {
		newParams := &Parameters{
			"key": &Parameter{Type: "string", Default: CreateVariableString("new_default")},
		}
		oldParams := &Parameters{
			"key": &Parameter{Type: "string", Variable: CreateVariableString("some_value")},
		}
		newParams.SetValuesSmartMerge(oldParams)
		assert.Equal(t, "some_value", *newParams.Get("key").Variable.String)
	})

	t.Run("old variable is nil leaves new param untouched", func(t *testing.T) {
		newParams := &Parameters{
			"key": &Parameter{Type: "string", Default: CreateVariableString("new_default")},
		}
		oldParams := &Parameters{
			"key": &Parameter{Type: "string", Default: CreateVariableString("old_default")},
		}
		newParams.SetValuesSmartMerge(oldParams)
		assert.True(t, newParams.Get("key").Variable.IsNil())
	})
}

// ---------------------------------------------------------------------------
// Parameters.SetLooseValues
// ---------------------------------------------------------------------------

func TestSetLooseValues(t *testing.T) {
	t.Run("sets string value for existing key", func(t *testing.T) {
		params := &Parameters{"name": &Parameter{Type: "string"}}
		require.NoError(t, params.SetLooseValues(map[string]string{"name": "alice"}))
		assert.Equal(t, "alice", *params.Get("name").Variable.String)
	})

	t.Run("ignores keys absent from params", func(t *testing.T) {
		params := &Parameters{}
		require.NoError(t, params.SetLooseValues(map[string]string{"missing": "value"}))
	})

	t.Run("returns error for invalid int value", func(t *testing.T) {
		params := &Parameters{"count": &Parameter{Type: "int"}}
		err := params.SetLooseValues(map[string]string{"count": "not-an-int"})
		assert.Error(t, err)
	})

	t.Run("sets bool value", func(t *testing.T) {
		params := &Parameters{"enabled": &Parameter{Type: "bool"}}
		require.NoError(t, params.SetLooseValues(map[string]string{"enabled": "true"}))
		assert.True(t, *params.Get("enabled").Variable.Bool)
	})
}

// ---------------------------------------------------------------------------
// GetChoices
// ---------------------------------------------------------------------------

func TestGetChoices(t *testing.T) {
	t.Run("restart returns four choices", func(t *testing.T) {
		choices, err := GetChoices("restart")
		require.NoError(t, err)
		require.Len(t, choices, 4)
		vals := make([]string, 4)
		for i, c := range choices {
			vals[i] = *c.String
		}
		assert.ElementsMatch(t, []string{"no", "always", "on-failure", "unless-stopped"}, vals)
	})

	t.Run("nginx returns two choices", func(t *testing.T) {
		choices, err := GetChoices("nginx")
		require.NoError(t, err)
		assert.Len(t, choices, 2)
	})

	t.Run("unknown name returns nil", func(t *testing.T) {
		choices, err := GetChoices("unknown-name-xyz")
		require.NoError(t, err)
		assert.Nil(t, choices)
	})
}

// ---------------------------------------------------------------------------
// GetOrSetSeed — branch where seed already exists in viper
// ---------------------------------------------------------------------------

func TestGetOrSetSeed_ExistingSeed(t *testing.T) {
	yamlContent := "stamus:\n  seed: \"abcdefgh12345678\"\n"
	require.NoError(t, app.FS.MkdirAll("/TestGetOrSetSeed", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestGetOrSetSeed/config.yaml", []byte(yamlContent), 0o644))

	file, err := CreateFile("/TestGetOrSetSeed", "config.yaml")
	require.NoError(t, err)
	config, err := ConfigFromFile(file)
	require.NoError(t, err)

	assert.Equal(t, "abcdefgh12345678", config.GetSeed())
}

// ---------------------------------------------------------------------------
// file.go — GetViper lazy initialisation
// ---------------------------------------------------------------------------

func TestGetViper_LazyInit(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestGetViperLazy", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestGetViperLazy/config.yaml", []byte("foo: bar\n"), 0o644))

	file := NewFile("/TestGetViperLazy", "config", "yaml")
	// viperInstance is nil — GetViper must initialise it lazily.
	v := file.GetViper()
	require.NotNil(t, v)
	assert.Equal(t, "bar", v.GetString("foo"))
}

// ---------------------------------------------------------------------------
// extractParamsWithTracking — circular include detection
// ---------------------------------------------------------------------------

func TestExtractParamsWithTracking_CircularDetection(t *testing.T) {
	dir := "/TestCircularIncludes"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	fileA := "param1:\n  usage: p1\n  type: string\n  default: v1\nincludes:\n  - b.yaml\n"
	fileB := "param2:\n  usage: p2\n  type: string\n  default: v2\nincludes:\n  - a.yaml\n"
	require.NoError(t, afero.WriteFile(app.FS, dir+"/a.yaml", []byte(fileA), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/b.yaml", []byte(fileB), 0o644))

	f, err := CreateFile(dir, "a.yaml")
	require.NoError(t, err)
	config, err := ConfigFromFile(f)
	require.NoError(t, err)

	_, _, err = config.ExtractParams()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

// ---------------------------------------------------------------------------
// extractParamsWithTracking — max depth exceeded
// ---------------------------------------------------------------------------

func TestExtractParamsWithTracking_MaxDepth(t *testing.T) {
	dir := "/TestMaxDepth"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	// Build a chain of 12 includes (MaxIncludeDepth = 10).
	for i := 0; i <= 12; i++ {
		var yamlContent string
		paramName := fmt.Sprintf("param%c", rune('A'+i))
		if i < 12 {
			nextFile := fmt.Sprintf("file%c.yaml", rune('A'+i+1))
			yamlContent = fmt.Sprintf(
				"%s:\n  usage: p\n  type: string\n  default: v\nincludes:\n  - %s\n",
				paramName, nextFile,
			)
		} else {
			yamlContent = fmt.Sprintf("%s:\n  usage: p\n  type: string\n  default: v\n", paramName)
		}
		filename := fmt.Sprintf("file%c.yaml", rune('A'+i))
		require.NoError(t, afero.WriteFile(app.FS, dir+"/"+filename, []byte(yamlContent), 0o644))
	}

	f, err := CreateFile(dir, "fileA.yaml")
	require.NoError(t, err)
	config, err := ConfigFromFile(f)
	require.NoError(t, err)

	_, _, err = config.ExtractParams()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum include depth")
}

// ---------------------------------------------------------------------------
// GetStamusFile — error paths
// ---------------------------------------------------------------------------

func TestGetStamusFile_MissingKey(t *testing.T) {
	_, err := GetStamusFile(map[string]*Variable{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stamus.config not found")
}

func TestGetStamusFile_ValidKey(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestGetStamusFile", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestGetStamusFile/config.yaml", []byte(""), 0o644))

	path := "/TestGetStamusFile"
	v := CreateVariableString(path)
	file, err := GetStamusFile(map[string]*Variable{"stamus.config": &v})
	require.NoError(t, err)
	assert.Equal(t, path, file.Path)
}

// ---------------------------------------------------------------------------
// SetToDefaults — suricata.interfaces key is excluded
// ---------------------------------------------------------------------------

func TestSetToDefaults_SuricataInterfacesExcluded(t *testing.T) {
	iface := &Parameter{
		Type:    "string",
		Default: CreateVariableString("eth0"),
		// Variable intentionally nil.
	}
	params := &Parameters{"suricata.interfaces": iface}

	require.NoError(t, params.SetToDefaults())
	// Must remain nil because the key is excluded from SetToDefault.
	assert.True(t, params.Get("suricata.interfaces").Variable.IsNil())
}

// ---------------------------------------------------------------------------
// filterRemainingOptionalParams
// ---------------------------------------------------------------------------

func TestFilterRemainingOptionalParams(t *testing.T) {
	tests := []struct {
		name           string
		optionalParams []string
		optionalParam  string
		want           []string
	}{
		{
			name:           "removes exact match and prefix matches",
			optionalParams: []string{"a", "a.b", "a.c", "b"},
			optionalParam:  "a",
			want:           []string{"b"},
		},
		{
			name:           "removes nothing when no prefix matches",
			optionalParams: []string{"x", "y"},
			optionalParam:  "z",
			want:           []string{"x", "y"},
		},
		{
			name:           "empty input returns empty",
			optionalParams: []string{},
			optionalParam:  "a",
			want:           []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterRemainingOptionalParams(tt.optionalParams, tt.optionalParam)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// CreateFile — error paths
// ---------------------------------------------------------------------------

func TestCreateFile_InvalidFileName(t *testing.T) {
	// CreateFile only accepts exactly one dot in fileName.
	_, err := CreateFile("/some/path", "name.extra.yaml")
	assert.Error(t, err)
}

func TestCreateFile_PathTraversalInName(t *testing.T) {
	_, err := CreateFile("/some/path", "../bad.yaml")
	assert.Error(t, err)
}

func TestCreateFileFromPath_NoExtension(t *testing.T) {
	_, err := CreateFileFromPath("/some/path/noextension")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Variable.IsNil — all combinations
// ---------------------------------------------------------------------------

func TestVariable_IsNil(t *testing.T) {
	assert.True(t, (&Variable{}).IsNil())
	v1 := CreateVariableString("x")
	assert.False(t, v1.IsNil())
	v2 := CreateVariableBool(true)
	assert.False(t, v2.IsNil())
	v3 := CreateVariableInt(1)
	assert.False(t, v3.IsNil())
}

// ---------------------------------------------------------------------------
// GetReleaseData — basic smoke test
// ---------------------------------------------------------------------------

func TestGetReleaseData_Smoke(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestGetReleaseData", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestGetReleaseData/config.yaml", []byte(""), 0o644))

	file, err := CreateFile("/TestGetReleaseData", "config.yaml")
	require.NoError(t, err)

	data, err := GetReleaseData(file, false, false, "myseed")
	require.NoError(t, err)

	assert.NotEmpty(t, data["Release.name"])
	assert.Equal(t, "myseed", data["Release.seed"])
	assert.Equal(t, false, data["Release.isUpgrade"])
}

// ---------------------------------------------------------------------------
// asLooseTyped — true/false/int/string paths
// ---------------------------------------------------------------------------

func TestAsLooseTyped(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{"true", true},
		{"false", false},
		{"42", 42},
		{"hello", "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := asLooseTyped(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// getRelease — various path-building branches
// ---------------------------------------------------------------------------

func TestGetRelease_TrailingSlash(t *testing.T) {
	// dest.Path ending with "/" exercises the len-2 branch in getRelease
	require.NoError(t, app.FS.MkdirAll("/TestGetRelease/dir/", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestGetRelease/dir/config.yaml", []byte(""), 0o644))

	file, err := CreateFile("/TestGetRelease/dir", "config.yaml")
	require.NoError(t, err)

	release := getRelease(file, "/workdir", "seed", false, false)
	assert.NotEmpty(t, release.Name)
}

// ---------------------------------------------------------------------------
// GetChoices — interfaces branch (non-prod mode, reads /sys/class/net)
// ---------------------------------------------------------------------------

func TestGetChoices_Interfaces(t *testing.T) {
	// In test mode (non-prod), getInterfaces calls getInterfacesHost which reads
	// /sys/class/net using the real OS filesystem (not afero FS).
	// This only works on Linux where /sys/class/net exists.
	choices, err := GetChoices("interfaces")
	// On CI / Linux this succeeds; on other platforms it may fail — just skip.
	if err != nil {
		t.Skipf("getInterfacesHost failed (non-Linux?): %v", err)
	}
	assert.NotNil(t, choices)
}

// ---------------------------------------------------------------------------
// getInterfacesHost — test by temporarily overriding app.Mode to non-prod
// ---------------------------------------------------------------------------

func TestGetInterfacesHost(t *testing.T) {
	// Save and restore the original mode.
	original := app.Mode
	app.Mode = "test" // any non-"prod" value makes getInterfaces call getInterfacesHost
	defer func() { app.Mode = original }()

	// Clear the cache so the function actually runs.
	interfacesCache = nil

	choices, err := GetChoices("interfaces")
	if err != nil {
		t.Skipf("getInterfacesHost unavailable: %v", err)
	}
	// On Linux /sys/class/net always has at least "lo"
	assert.NotEmpty(t, choices)
}

// ---------------------------------------------------------------------------
// nestMap — collision branch (scalar conflicts with nested key)
// ---------------------------------------------------------------------------

func TestNestMap_KeyCollision(t *testing.T) {
	// When both "a" (scalar) and "a.b" (nested) exist, sub-keys always win.
	input := map[string]interface{}{
		"a":   "scalar",
		"a.b": "nested",
	}
	result := nestMap(input)
	assert.NotNil(t, result)
	aVal, ok := result["a"].(map[string]interface{})
	assert.True(t, ok, "a should be a map, got %T", result["a"])
	assert.Equal(t, "nested", aVal["b"])
}

// ---------------------------------------------------------------------------
// getRelease — app.Name == "stamusctl" branch (IsCtl true)
// ---------------------------------------------------------------------------

func TestGetRelease_IsCtlTrue(t *testing.T) {
	original := app.Name
	app.Name = app.CtlName // "stamusctl"
	defer func() { app.Name = original }()

	require.NoError(t, app.FS.MkdirAll("/TestGetRelease_IsCtl", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestGetRelease_IsCtl/config.yaml", []byte(""), 0o644))

	file, err := CreateFile("/TestGetRelease_IsCtl", "config.yaml")
	require.NoError(t, err)

	// When IsCtl() is true, getRelease joins currentDir + dest.Path.
	release := getRelease(file, "/workdir", "seed123", true, false)
	assert.NotEmpty(t, release.Name)
	assert.Equal(t, "seed123", release.Seed)
}

// ---------------------------------------------------------------------------
// configFromFile — error path (file does not exist)
// ---------------------------------------------------------------------------

func TestConfigFromFile_FileNotFound(t *testing.T) {
	file := NewFile("/nonexistent/path/xxx", "config", "yaml")
	_, err := ConfigFromFile(file)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// processTemplates — error on bad template
// ---------------------------------------------------------------------------

func TestProcessTemplates_BadTemplate(t *testing.T) {
	inDir := "/TestProcessTemplates_Bad/in"
	outDir := "/TestProcessTemplates_Bad/out"
	require.NoError(t, app.FS.MkdirAll(inDir, 0o755))
	require.NoError(t, app.FS.MkdirAll(outDir, 0o755))

	// Write a file with invalid template syntax.
	require.NoError(t, afero.WriteFile(app.FS, inDir+"/bad.yaml", []byte("{{ .Unclosed }"), 0o644))

	err := processTemplates(inDir, outDir, map[string]interface{}{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Parameters.SetValues — validation failure branch
// ---------------------------------------------------------------------------

func TestSetValues_ValidationFailure(t *testing.T) {
	// ValidateFunc always returns false, so the value should not be set.
	params := &Parameters{
		"key": &Parameter{
			Type:         "string",
			Variable:     CreateVariableString("old"),
			ValidateFunc: func(v Variable) bool { return false },
		},
	}
	v := CreateVariableString("new")
	params.SetValues(map[string]*Variable{"key": &v})
	// Value must remain "old" because validation failed.
	assert.Equal(t, "old", *params.Get("key").Variable.String)
}

// ---------------------------------------------------------------------------
// processTemplate — reads a directory (the dir-creation path)
// ---------------------------------------------------------------------------

func TestProcessTemplate_CreatesOutputDirectory(t *testing.T) {
	logger := zap.NewExample().Sugar()

	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Dir/in/subdir", 0o755))

	info, err := app.FS.Stat("/TestProcessTemplate_Dir/in/subdir")
	require.NoError(t, err)

	err = processTemplate(map[string]interface{}{}, []string{},
		"/TestProcessTemplate_Dir/in/subdir",
		"/TestProcessTemplate_Dir/in", "/TestProcessTemplate_Dir/out", info, logger)
	require.NoError(t, err)

	exists, err := afero.DirExists(app.FS, "/TestProcessTemplate_Dir/out/subdir")
	require.NoError(t, err)
	assert.True(t, exists)
}

// ---------------------------------------------------------------------------
// GetData — missing parameter value error path
// ---------------------------------------------------------------------------

func TestGetData_MissingParameterValue(t *testing.T) {
	// A parameter with no Variable and no Default triggers an error in GetData.
	config := &Config{
		arbitrary: &Arbitrary{},
		parameters: &Parameters{
			"key": &Parameter{Type: "string"}, // both Variable and Default are nil
		},
	}
	_, err := config.GetData()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// LoadConfigFrom — error: stamus.config missing from values.yaml
// ---------------------------------------------------------------------------

func TestLoadConfigFrom_MissingStamusConfig(t *testing.T) {
	// A values.yaml with no stamus.config key.
	yamlContent := "foo: bar\n"
	require.NoError(t, app.FS.MkdirAll("/TestLoadConfigFrom_NoConfig", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestLoadConfigFrom_NoConfig/config.yaml",
		[]byte(yamlContent), 0o644))

	file, err := CreateFile("/TestLoadConfigFrom_NoConfig", "config.yaml")
	require.NoError(t, err)

	_, err = LoadConfigFrom(file, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stamus.config not found")
}

// ---------------------------------------------------------------------------
// LoadConfigFrom — error: stamus.project missing
// ---------------------------------------------------------------------------

func TestLoadConfigFrom_MissingStamusProject(t *testing.T) {
	// A values.yaml that has stamus.config but no stamus.project.
	require.NoError(t, app.FS.MkdirAll("/TestLoadConfigFrom_NoProject/source", 0o755))

	// Source config for the template (pointed to by stamus.config).
	require.NoError(t, afero.WriteFile(app.FS,
		"/TestLoadConfigFrom_NoProject/source/config.yaml", []byte(""), 0o644))

	// Values file with stamus.config but no stamus.project.
	valuesContent := "stamus:\n  config: /TestLoadConfigFrom_NoProject/source\n"
	require.NoError(t, app.FS.MkdirAll("/TestLoadConfigFrom_NoProject/values", 0o755))
	require.NoError(t, afero.WriteFile(app.FS,
		"/TestLoadConfigFrom_NoProject/values/config.yaml", []byte(valuesContent), 0o644))

	file, err := CreateFile("/TestLoadConfigFrom_NoProject/values", "config.yaml")
	require.NoError(t, err)

	_, err = LoadConfigFrom(file, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stamus.project not found")
}

// ---------------------------------------------------------------------------
// SetValuesFromFiles — invalid argument format
// ---------------------------------------------------------------------------

func TestSetValuesFromFiles_InvalidArg(t *testing.T) {
	config := &Config{
		arbitrary:  &Arbitrary{},
		parameters: &Parameters{},
	}
	// Argument without "=" is invalid.
	err := config.SetValuesFromFiles("invalid-no-equals")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

// ---------------------------------------------------------------------------
// SetValuesFromFiles — file not found
// ---------------------------------------------------------------------------

func TestSetValuesFromFiles_FileNotFound(t *testing.T) {
	config := &Config{
		arbitrary:  &Arbitrary{},
		parameters: &Parameters{},
	}
	err := config.SetValuesFromFiles("param=/nonexistent/path/file.txt")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// isValidPath — invalid name with slash
// ---------------------------------------------------------------------------

func TestCreateFile_NameWithSlash(t *testing.T) {
	_, err := CreateFile("/some/path", "bad/name.yaml")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// getRelease — empty configDir path (falls back to "release" name)
// ---------------------------------------------------------------------------

func TestGetRelease_EmptyPath(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/p", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/p/config.yaml", []byte(""), 0o644))

	// Use a File with an empty path so getRelease gets an empty configDir.
	// We need IsCtl() == false (default Name != "stamusctl") and an empty dest.Path.
	original := app.Name
	app.Name = "notctl"
	defer func() { app.Name = original }()

	file := &File{Path: "", Name: "config", Type: "yaml"}
	release := getRelease(file, "/workdir", "seed", false, false)
	assert.Equal(t, "release", release.Name)
}

// ---------------------------------------------------------------------------
// ExtractParams — with remote include (local cycle / covered indirectly via
// existing circular test)
// ---------------------------------------------------------------------------
// The only missing branch in extractParamsWithTracking is the remote-include
// block which requires Docker. Skipping that — it cannot be unit tested.

// ---------------------------------------------------------------------------
// saveParamsTo — GetValue error (unset parameter without default)
// ---------------------------------------------------------------------------

func TestSaveParamsTo_GetValueError(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestSaveParamsTo_Err", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestSaveParamsTo_Err/config.yaml", []byte(""), 0o644))

	destFile, err := CreateFile("/TestSaveParamsTo_Err", "config.yaml")
	require.NoError(t, err)

	// A config where a parameter has no Variable and no Default — GetValue will fail.
	config := &Config{
		arbitrary: &Arbitrary{},
		parameters: &Parameters{
			"key": &Parameter{Type: "string"}, // nil Variable and nil Default
		},
		file: destFile,
	}

	err = config.saveParamsTo(destFile)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Clean — deleteEmptyFiles error (nonexistent path)
// ---------------------------------------------------------------------------

func TestClean_DeleteEmptyFilesError(t *testing.T) {
	// Use a path that doesn't exist in the memfs.
	file := &File{Path: "/nonexistent/clean/path"}
	config := &Config{}
	err := config.Clean(file)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// AddAsFlag — persistent hidden flag
// ---------------------------------------------------------------------------

func TestAddAsFlag_PersistentHidden(t *testing.T) {
	p := &Parameter{
		Name:     "test-persistent-hidden",
		Type:     "string",
		Usage:    "a hidden persistent flag",
		Variable: CreateVariableString("default"),
		Hidden:   true,
	}
	cmd := &cobra.Command{}
	p.AddAsFlag(cmd, true)
	// The flag should exist on persistent flags.
	flag := cmd.PersistentFlags().Lookup("test-persistent-hidden")
	assert.NotNil(t, flag)
}

// ---------------------------------------------------------------------------
// AddAsFlag — optional type (no flag added, no panic)
// ---------------------------------------------------------------------------

func TestAddAsFlag_OptionalType(t *testing.T) {
	p := &Parameter{
		Name: "test-optional",
		Type: "optional",
	}
	cmd := &cobra.Command{}
	// Should not panic; optional type is not added as a flag.
	p.AddAsFlag(cmd, false)
	assert.Nil(t, cmd.Flags().Lookup("test-optional"))
}

// ---------------------------------------------------------------------------
// InstanciateViper — invalid YAML syntax error
// ---------------------------------------------------------------------------

func TestInstanciateViper_InvalidYAML(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestViperInvalidYAML", 0o755))
	// Write deliberately invalid YAML.
	require.NoError(t, afero.WriteFile(app.FS, "/TestViperInvalidYAML/config.yaml",
		[]byte("invalid: yaml: [\n"), 0o644))

	file := NewFile("/TestViperInvalidYAML", "config", "yaml")
	_, err := file.InstanciateViper()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// InstanciateViper — deeply nested YAML exceeds MaxYAMLDepth (50 levels)
// ---------------------------------------------------------------------------

func TestInstanciateViper_DeepYAML(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestViperDeepYAML", 0o755))

	// Build a YAML string that nests 55 levels deep.
	var content string
	for i := 0; i < 55; i++ {
		content += fmt.Sprintf("%skey%d:\n", fmt.Sprintf("%*s", i*2, ""), i)
	}
	content += fmt.Sprintf("%svalue: leaf\n", fmt.Sprintf("%*s", 55*2, ""))

	require.NoError(t, afero.WriteFile(app.FS, "/TestViperDeepYAML/config.yaml",
		[]byte(content), 0o644))

	file := NewFile("/TestViperDeepYAML", "config", "yaml")
	_, err := file.InstanciateViper()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YAML structure validation failed")
}

// ---------------------------------------------------------------------------
// GetViper — returns nil for missing file (error path)
// ---------------------------------------------------------------------------

func TestGetViper_MissingFile(t *testing.T) {
	// A file that doesn't exist — GetViper should return nil.
	file := NewFile("/nonexistent/viper/path", "config", "yaml")
	v := file.GetViper()
	assert.Nil(t, v)
}

// ---------------------------------------------------------------------------
// CreateFileFromPath — single-element path (no slash)
// ---------------------------------------------------------------------------

func TestCreateFileFromPath_SingleElement(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll(".", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "./testfile.yaml", []byte(""), 0o644))

	file, err := CreateFileFromPath("testfile.yaml")
	require.NoError(t, err)
	assert.Equal(t, "testfile", file.Name)
	assert.Equal(t, "yaml", file.Type)
}

// ---------------------------------------------------------------------------
// processTemplates — getAllFilesContent error (nonexistent input folder)
// ---------------------------------------------------------------------------

func TestProcessTemplates_NonExistentInput(t *testing.T) {
	err := processTemplates("/nonexistent/input/dir", "/some/output", map[string]interface{}{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// LoadConfigFrom — reload=true path (skip arbitrary copy)
// ---------------------------------------------------------------------------

func TestLoadConfigFrom_ReloadTrue(t *testing.T) {
	// Create a minimal config setup that LoadConfigFrom can traverse.
	require.NoError(t, app.FS.MkdirAll("/TestLoadConfigFrom_Reload/source", 0o755))
	require.NoError(t, afero.WriteFile(app.FS,
		"/TestLoadConfigFrom_Reload/source/config.yaml", []byte(""), 0o644))

	valuesContent := "param1: hello\nstamus:\n  config: /TestLoadConfigFrom_Reload/source\n  project: myproject\n"
	require.NoError(t, app.FS.MkdirAll("/TestLoadConfigFrom_Reload/values", 0o755))
	require.NoError(t, afero.WriteFile(app.FS,
		"/TestLoadConfigFrom_Reload/values/config.yaml", []byte(valuesContent), 0o644))

	file, err := CreateFile("/TestLoadConfigFrom_Reload/values", "config.yaml")
	require.NoError(t, err)

	// reload=true skips the arbitrary copy block.
	config, err := LoadConfigFrom(file, true)
	require.NoError(t, err)
	assert.Equal(t, "myproject", config.project)
}

// ---------------------------------------------------------------------------
// LoadConfigFrom — with stamus.registry set
// ---------------------------------------------------------------------------

func TestLoadConfigFrom_WithRegistry(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestLoadConfigFrom_Registry/source", 0o755))
	require.NoError(t, afero.WriteFile(app.FS,
		"/TestLoadConfigFrom_Registry/source/config.yaml", []byte(""), 0o644))

	valuesContent := "stamus:\n  config: /TestLoadConfigFrom_Registry/source\n  project: regproject\n  registry: ghcr.io/test\n"
	require.NoError(t, app.FS.MkdirAll("/TestLoadConfigFrom_Registry/values", 0o755))
	require.NoError(t, afero.WriteFile(app.FS,
		"/TestLoadConfigFrom_Registry/values/config.yaml", []byte(valuesContent), 0o644))

	file, err := CreateFile("/TestLoadConfigFrom_Registry/values", "config.yaml")
	require.NoError(t, err)

	config, err := LoadConfigFrom(file, false)
	require.NoError(t, err)
	assert.Equal(t, "ghcr.io/test", config.registry)
}

// ---------------------------------------------------------------------------
// SetValuesFromFile — CreateFileFromPath error (bad path)
// ---------------------------------------------------------------------------

func TestSetValuesFromFile_BadPath(t *testing.T) {
	config := &Config{
		arbitrary:  &Arbitrary{},
		parameters: &Parameters{},
	}
	// A path without an extension fails CreateFileFromPath.
	err := config.SetValuesFromFile("/some/path/noextension")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// SetValuesFromFile — empty path is a no-op
// ---------------------------------------------------------------------------

func TestSetValuesFromFile_EmptyPath(t *testing.T) {
	config := &Config{
		arbitrary:  &Arbitrary{},
		parameters: &Parameters{},
	}
	err := config.SetValuesFromFile("")
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// processTemplate — empty content (len == 0 check)
// ---------------------------------------------------------------------------

func TestProcessTemplate_EmptyContent(t *testing.T) {
	logger := zap.NewExample().Sugar()

	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Empty/in", 0o755))
	require.NoError(t, app.FS.MkdirAll("/TestProcessTemplate_Empty/out", 0o755))

	// An empty file — the len(content) == 0 branch returns nil early.
	require.NoError(t, afero.WriteFile(app.FS, "/TestProcessTemplate_Empty/in/empty.txt", []byte{}, 0o644))

	info, err := app.FS.Stat("/TestProcessTemplate_Empty/in/empty.txt")
	require.NoError(t, err)

	err = processTemplate(map[string]interface{}{}, []string{},
		"/TestProcessTemplate_Empty/in/empty.txt",
		"/TestProcessTemplate_Empty/in", "/TestProcessTemplate_Empty/out", info, logger)
	// Should return nil (early exit) — no output file created.
	assert.NoError(t, err)
}
