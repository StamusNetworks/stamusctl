package models

// extra_coverage_test.go — additional tests to push internal/models coverage
// past 80%.  All test names are checked against existing tests to avoid
// redeclaration errors.

import (
	"embed"
	"fmt"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ExtractEmbedTo — real embed.FS (0 % → covered)
//
// getAllFilesEmbed("testdata", ...) produces "testdata/fixture.txt" which
// embed.FS.ReadFile can handle (unlike the "./" prefix for root paths).
// ---------------------------------------------------------------------------

//go:embed testdata
var testdataEmbed embed.FS

func TestExtractEmbedTo_WithTestdataEmbed(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	outputFolder := "/TestExtractEmbedTo_Testdata"

	err := ExtractEmbedTo("testdata", testdataEmbed, outputFolder)
	require.NoError(t, err)

	// The fixture file should have been extracted.
	exists, err := afero.Exists(app.FS, outputFolder+"/fixture.txt")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExtractEmbedTo_ContentVerification(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	outputFolder := "/TestExtractEmbedTo_Content"

	err := ExtractEmbedTo("testdata", testdataEmbed, outputFolder)
	require.NoError(t, err)

	content, err := afero.ReadFile(app.FS, outputFolder+"/fixture.txt")
	require.NoError(t, err)
	assert.Contains(t, string(content), "test fixture content")
}

// getAllFilesEmbed — traverses testdata directory and returns file paths.
func TestGetAllFilesEmbed_TestdataDir(t *testing.T) {
	files := getAllFilesEmbed("testdata", testdataEmbed)
	assert.NotEmpty(t, files)
	for _, f := range files {
		assert.NotEmpty(t, f, "file path must not be empty")
	}
}

// ---------------------------------------------------------------------------
// extractParam — full path with choices (83 % → more coverage)
// ---------------------------------------------------------------------------

func TestExtractParam_WithNginxChoices(t *testing.T) {
	dir := "/TestExtractParam_Nginx"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	content := `
proxy:
  usage: proxy mode
  type: string
  default: nginx
  choices: nginx
`
	require.NoError(t, afero.WriteFile(app.FS, dir+"/config.yaml", []byte(content), 0o644))

	f, err := CreateFile(dir, "config.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	param, err := cfg.extractParam("proxy")
	require.NoError(t, err)
	assert.NotNil(t, param)
	assert.Len(t, param.Choices, 2) // "nginx" choice set returns 2 items
}

func TestExtractParam_BoolType(t *testing.T) {
	dir := "/TestExtractParam_Bool"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	content := `
enabled:
  usage: enable flag
  type: bool
  default: false
`
	require.NoError(t, afero.WriteFile(app.FS, dir+"/config.yaml", []byte(content), 0o644))

	f, err := CreateFile(dir, "config.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	param, err := cfg.extractParam("enabled")
	require.NoError(t, err)
	assert.Equal(t, "bool", param.Type)
}

func TestExtractParam_IntType(t *testing.T) {
	dir := "/TestExtractParam_Int"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	content := `
count:
  usage: item count
  type: int
  default: 5
`
	require.NoError(t, afero.WriteFile(app.FS, dir+"/config.yaml", []byte(content), 0o644))

	f, err := CreateFile(dir, "config.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	param, err := cfg.extractParam("count")
	require.NoError(t, err)
	assert.Equal(t, "int", param.Type)
	assert.Equal(t, 5, *param.Default.Int)
}

// ---------------------------------------------------------------------------
// extractParamsWithTracking — local include (47 % → more branches hit)
// ---------------------------------------------------------------------------

func TestExtractParamsWithTracking_LocalInclude_HappyPath(t *testing.T) {
	dir := "/TestLocalInclude_Happy"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	parent := `
param1:
  usage: p1
  type: string
  default: v1
includes:
  - child.yaml
`
	child := `
param2:
  usage: p2
  type: int
  default: 42
`
	require.NoError(t, afero.WriteFile(app.FS, dir+"/parent.yaml", []byte(parent), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/child.yaml", []byte(child), 0o644))

	f, err := CreateFile(dir, "parent.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	params, includes, err := cfg.ExtractParams()
	require.NoError(t, err)
	assert.NotNil(t, params)
	assert.NotNil(t, params.Get("param1"))
	assert.NotNil(t, params.Get("param2"))
	assert.Contains(t, includes, "child.yaml")
}

func TestExtractParamsWithTracking_IncludedFileNotFound_Error(t *testing.T) {
	dir := "/TestIncludeMissing_Error"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	content := `
param1:
  usage: p1
  type: string
  default: v1
includes:
  - nonexistent.yaml
`
	require.NoError(t, afero.WriteFile(app.FS, dir+"/config.yaml", []byte(content), 0o644))

	f, err := CreateFile(dir, "config.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	_, _, err = cfg.ExtractParams()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// setCachedRemoteInclude / getCachedRemoteInclude — happy-path round trip.
// ---------------------------------------------------------------------------

func TestSetAndGetCachedRemoteInclude_MemFS(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	url := "ghcr.io/test/image:v1/include.yaml"
	content := []byte("key: value\n")

	err := setCachedRemoteInclude(url, content)
	require.NoError(t, err)

	got, found := getCachedRemoteInclude(url)
	assert.True(t, found)
	assert.Equal(t, content, got)
}

// ---------------------------------------------------------------------------
// ProcessOptionnalParams — non-interactive path.
//
// When default is false: cleanOptionatedParams removes children; the optional
// param itself stays in the map (only the children with its prefix are removed).
// When default is true: delete(*p, optionalParam) removes the optional marker.
// ---------------------------------------------------------------------------

func TestProcessOptionnalParams_NonInteractiveFalse_ChildrenRemoved(t *testing.T) {
	params := &Parameters{
		"feature":       &Parameter{Type: "optional", Default: CreateVariableBool(false)},
		"feature.child": &Parameter{Type: "string", Default: CreateVariableString("val")},
	}

	err := params.ProcessOptionnalParams(false)
	require.NoError(t, err)

	// feature.child must be gone (removed by cleanOptionatedParams).
	assert.Nil(t, params.Get("feature.child"))
	// feature itself stays in the map when bool is false (it's not deleted).
	// The optional param itself is NOT deleted when false, only children are.
}

func TestProcessOptionnalParams_NonInteractiveTrue_OptionalRemoved(t *testing.T) {
	params := &Parameters{
		"opt":     &Parameter{Type: "optional", Default: CreateVariableBool(true)},
		"opt.key": &Parameter{Type: "string", Default: CreateVariableString("x")},
	}

	err := params.ProcessOptionnalParams(false)
	require.NoError(t, err)

	// When true: delete(*p, optionalParam) removes the optional marker.
	assert.Nil(t, params.Get("opt"))
	// Children are preserved.
	assert.NotNil(t, params.Get("opt.key"))
}

// ---------------------------------------------------------------------------
// getInterfacesHost — cache hit path (no filesystem access)
// ---------------------------------------------------------------------------

func TestGetInterfacesHost_CacheAlreadyPopulated(t *testing.T) {
	original := interfacesCache
	interfacesCache = []Variable{
		CreateVariableString("lo"),
		CreateVariableString("eth0"),
	}
	defer func() { interfacesCache = original }()

	result, err := getInterfacesHost()
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// ---------------------------------------------------------------------------
// getInterfacesBusybox — cache hit path (no Docker required)
// ---------------------------------------------------------------------------

func TestGetInterfacesBusybox_CachePopulated(t *testing.T) {
	original := interfacesCache
	interfacesCache = []Variable{CreateVariableString("lo")}
	defer func() { interfacesCache = original }()

	result, err := getInterfacesBusybox()
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "lo", *result[0].String)
}

// ---------------------------------------------------------------------------
// getAllFilesContent — no matching extension → empty slice
// ---------------------------------------------------------------------------

func TestGetAllFilesContent_ExtensionMismatch_Returns_Empty(t *testing.T) {
	dir := "/TestGetAllFilesContent_Ext"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/data.txt", []byte("hello"), 0o644))

	contents, err := getAllFilesContent(dir, ".yaml")
	require.NoError(t, err)
	assert.Empty(t, contents)
}

// ---------------------------------------------------------------------------
// deleteEmptyFiles — Walk error on non-existent folder
// ---------------------------------------------------------------------------

func TestDeleteEmptyFiles_NonExistentFolder(t *testing.T) {
	err := deleteEmptyFiles("/NoSuchDir_deleteEmptyFiles_extra_test")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// deleteEmptyFolders — subdirectory gets cleaned up
// ---------------------------------------------------------------------------

func TestDeleteEmptyFolders_SubdirRemoved(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/TestDeleteEmptyFolders_Sub"
	sub := dir + "/empty"
	require.NoError(t, app.FS.MkdirAll(sub, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/keep.txt", []byte("x"), 0o644))

	err := deleteEmptyFolders(dir)
	assert.NoError(t, err)

	exists, _ := afero.DirExists(app.FS, sub)
	assert.False(t, exists)
}

// ---------------------------------------------------------------------------
// removeDirIfEmpty — non-empty dir is not removed
// ---------------------------------------------------------------------------

func TestRemoveDirIfEmpty_WithFile_KeepsDir(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/TestRemoveDirIfEmpty_NotEmpty2"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/file.txt", []byte("data"), 0o644))

	err := removeDirIfEmpty(dir)
	require.NoError(t, err)

	exists, _ := afero.DirExists(app.FS, dir)
	assert.True(t, exists)
}

func TestRemoveDirIfEmpty_EmptyDir_Removed(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/TestRemoveDirIfEmpty_Empty2"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	err := removeDirIfEmpty(dir)
	require.NoError(t, err)

	exists, _ := afero.DirExists(app.FS, dir)
	assert.False(t, exists)
}

// ---------------------------------------------------------------------------
// AddAsFlag — shorthand branches for string, bool, int (non-persistent)
// ---------------------------------------------------------------------------

func TestAddAsFlag_StringShorthand_NonPersistent(t *testing.T) {
	p := &Parameter{
		Name:      "output-extra",
		Shorthand: "O",
		Type:      "string",
		Usage:     "output path",
	}
	cmd := &cobra.Command{}
	p.AddAsFlag(cmd, false)
	require.NotNil(t, cmd.Flags().Lookup("output-extra"))
}

func TestAddAsFlag_BoolShorthand_NonPersistent(t *testing.T) {
	p := &Parameter{
		Name:      "verbose-extra",
		Shorthand: "V",
		Type:      "bool",
		Usage:     "verbose",
	}
	cmd := &cobra.Command{}
	p.AddAsFlag(cmd, false)
	require.NotNil(t, cmd.Flags().Lookup("verbose-extra"))
}

func TestAddAsFlag_IntShorthand_NonPersistent(t *testing.T) {
	p := &Parameter{
		Name:      "count-extra",
		Shorthand: "C",
		Type:      "int",
		Usage:     "count",
	}
	cmd := &cobra.Command{}
	p.AddAsFlag(cmd, false)
	require.NotNil(t, cmd.Flags().Lookup("count-extra"))
}

// ---------------------------------------------------------------------------
// AddAsFlag — shorthand branches (persistent)
// ---------------------------------------------------------------------------

func TestAddAsFlag_StringShorthand_Persistent(t *testing.T) {
	p := &Parameter{
		Name:      "host-extra",
		Shorthand: "H",
		Type:      "string",
		Usage:     "host",
	}
	cmd := &cobra.Command{}
	p.AddAsFlag(cmd, true)
	require.NotNil(t, cmd.PersistentFlags().Lookup("host-extra"))
}

func TestAddAsFlag_BoolShorthand_Persistent(t *testing.T) {
	p := &Parameter{
		Name:      "debug-extra",
		Shorthand: "D",
		Type:      "bool",
		Usage:     "debug",
	}
	cmd := &cobra.Command{}
	p.AddAsFlag(cmd, true)
	require.NotNil(t, cmd.PersistentFlags().Lookup("debug-extra"))
}

func TestAddAsFlag_IntShorthand_Persistent(t *testing.T) {
	p := &Parameter{
		Name:      "port-extra",
		Shorthand: "P",
		Type:      "int",
		Usage:     "port",
	}
	cmd := &cobra.Command{}
	p.AddAsFlag(cmd, true)
	require.NotNil(t, cmd.PersistentFlags().Lookup("port-extra"))
}

// ---------------------------------------------------------------------------
// isValidPath — type/name traversal sequences
// ---------------------------------------------------------------------------

func TestIsValidPath_TypeContainsDotDot(t *testing.T) {
	f := &File{Path: "/some", Name: "cfg", Type: "../evil"}
	err := f.isValidPath()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file type")
}

func TestIsValidPath_TypeContainsSlash(t *testing.T) {
	f := &File{Path: "/some", Name: "cfg", Type: "ya/ml"}
	err := f.isValidPath()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file type")
}

func TestIsValidPath_NameContainsDotDot(t *testing.T) {
	f := &File{Path: "/some", Name: "../../etc/shadow", Type: "yaml"}
	err := f.isValidPath()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file name")
}

func TestIsValidPath_NameContainsBackslash(t *testing.T) {
	f := &File{Path: "/some", Name: `cfg\bad`, Type: "yaml"}
	err := f.isValidPath()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file name")
}

// ---------------------------------------------------------------------------
// GetReleaseData — isUpgrade=true, isInstall=true path
// ---------------------------------------------------------------------------

func TestGetReleaseData_UpgradeAndInstall(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestGetRelData_Upgrade", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestGetRelData_Upgrade/config.yaml", []byte(""), 0o644))

	file, err := CreateFile("/TestGetRelData_Upgrade", "config.yaml")
	require.NoError(t, err)

	data, err := GetReleaseData(file, true, true, "newseed")
	require.NoError(t, err)
	assert.Equal(t, true, data["Release.isUpgrade"])
	assert.Equal(t, true, data["Release.isInstall"])
	assert.Equal(t, "newseed", data["Release.seed"])
}

// ---------------------------------------------------------------------------
// SetToDefaults — optional=false cleans children; true removes the marker.
// ---------------------------------------------------------------------------

func TestSetToDefaults_OptionalFalseCleanChildren(t *testing.T) {
	params := &Parameters{
		"opt":     &Parameter{Type: "optional", Default: CreateVariableBool(false)},
		"opt.sub": &Parameter{Type: "string", Default: CreateVariableString("v")},
	}
	require.NoError(t, params.SetToDefaults())
	// opt.sub must be cleaned because optional is false.
	assert.Nil(t, params.Get("opt.sub"))
}

func TestSetToDefaults_RegularParamNotSet_GetsDefault(t *testing.T) {
	params := &Parameters{
		"key": &Parameter{Type: "string", Default: CreateVariableString("mydefault")},
	}
	require.NoError(t, params.SetToDefaults())
	// The variable should now equal the default.
	assert.Equal(t, "mydefault", *params.Get("key").Variable.String)
}

// ---------------------------------------------------------------------------
// processTemplates — with a .sh file (exercises the chmod path)
// ---------------------------------------------------------------------------

func TestProcessTemplates_ShellScript(t *testing.T) {
	inDir := "/TestProcessTemplates_Sh/in"
	outDir := "/TestProcessTemplates_Sh/out"
	require.NoError(t, app.FS.MkdirAll(inDir, 0o755))
	require.NoError(t, app.FS.MkdirAll(outDir, 0o755))

	require.NoError(t, afero.WriteFile(app.FS, inDir+"/start.sh",
		[]byte("#!/bin/sh\necho {{ .Values.msg }}\n"), 0o644))

	data := map[string]interface{}{
		"Values": map[string]interface{}{"msg": "world"},
	}
	err := processTemplates(inDir, outDir, data)
	require.NoError(t, err)

	out, err := afero.ReadFile(app.FS, outDir+"/start.sh")
	require.NoError(t, err)
	assert.Contains(t, string(out), "echo world")
}

// ---------------------------------------------------------------------------
// GetData — success path with arbitrary and param values
// ---------------------------------------------------------------------------

func TestGetData_WithArbitraryAndParams(t *testing.T) {
	cfg := &Config{
		arbitrary: &Arbitrary{},
		parameters: &Parameters{
			"name": &Parameter{Type: "string", Default: CreateVariableString("alice")},
		},
	}
	cfg.arbitrary.SetArbitrary(map[string]string{"extra.key": "extval"})

	data, err := cfg.GetData()
	require.NoError(t, err)
	assert.Equal(t, "alice", data["Values.name"])
	assert.Equal(t, "extval", data["Values.extra.key"])
}

// ---------------------------------------------------------------------------
// isDirEmpty — non-existent path triggers an error
// ---------------------------------------------------------------------------

func TestIsDirEmpty_ErrorPath(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	_, err := isDirEmpty("/nonexistent/dir/isdircheck")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// LoadConfigFrom — seed branch (non-nil seed from values)
// ---------------------------------------------------------------------------

func TestLoadConfigFrom_WithSeedInValues(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestLoadCfg_Seed/src", 0o755))
	require.NoError(t, afero.WriteFile(app.FS,
		"/TestLoadCfg_Seed/src/config.yaml", []byte(""), 0o644))

	vals := "stamus:\n  config: /TestLoadCfg_Seed/src\n  project: myproj\n  seed: abcdef12345678\n"
	require.NoError(t, app.FS.MkdirAll("/TestLoadCfg_Seed/vals", 0o755))
	require.NoError(t, afero.WriteFile(app.FS,
		"/TestLoadCfg_Seed/vals/config.yaml", []byte(vals), 0o644))

	file, err := CreateFile("/TestLoadCfg_Seed/vals", "config.yaml")
	require.NoError(t, err)

	cfg, err := LoadConfigFrom(file, false)
	require.NoError(t, err)
	assert.Equal(t, "abcdef12345678", cfg.seed)
}

// ---------------------------------------------------------------------------
// cleanOptionatedParams — verifies prefix-based removal
// ---------------------------------------------------------------------------

func TestCleanOptionatedParams_RemovesChildren(t *testing.T) {
	params := &Parameters{
		"feat":         &Parameter{Type: "optional"},
		"feat.child":   &Parameter{Type: "string"},
		"feat.another": &Parameter{Type: "bool"},
		"other":        &Parameter{Type: "string"},
	}
	params.cleanOptionatedParams("feat")

	assert.Nil(t, params.Get("feat.child"))
	assert.Nil(t, params.Get("feat.another"))
	// The optional itself is NOT removed by cleanOptionatedParams.
	assert.NotNil(t, params.Get("feat"))
	// Unrelated param is preserved.
	assert.NotNil(t, params.Get("other"))
}

// ---------------------------------------------------------------------------
// getAllFiles — extension match returns correct files
// ---------------------------------------------------------------------------

func TestGetAllFiles_ReturnsMatchingFiles(t *testing.T) {
	dir := "/TestGetAllFiles_Match2"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/a.yaml", []byte(""), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/b.yaml", []byte(""), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/c.txt", []byte(""), 0o644))

	files, err := getAllFiles(dir, ".yaml")
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

// ---------------------------------------------------------------------------
// getRelease — path element trailing slash (splitted[last] == "")
// ---------------------------------------------------------------------------

func TestGetRelease_PathEndsWithSlash(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll("/TestGetRelease_Slash/dir", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/TestGetRelease_Slash/dir/config.yaml", []byte(""), 0o644))

	origName := app.Name
	app.Name = "notctl"
	defer func() { app.Name = origName }()

	file := &File{Path: "/TestGetRelease_Slash/dir/", Name: "config", Type: "yaml"}
	release := getRelease(file, "/workdir", "s", false, false)
	assert.Equal(t, "dir", release.Name)
}

// ---------------------------------------------------------------------------
// SetValuesFromFiles — valid key=file assignment exercises the full path
// ---------------------------------------------------------------------------

func TestSetValuesFromFiles_Valid(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, afero.WriteFile(app.FS, "/myval.txt", []byte("hello"), 0o644))

	cfg := &Config{
		arbitrary: &Arbitrary{},
		parameters: &Parameters{
			"greeting": &Parameter{
				Type:         "string",
				ValidateFunc: func(v Variable) bool { return true },
			},
		},
	}
	err := cfg.SetValuesFromFiles("greeting=/myval.txt")
	require.NoError(t, err)

	param := cfg.GetParams().Get("greeting")
	require.NotNil(t, param)
	assert.Equal(t, "hello", *param.Variable.String)
}

// ---------------------------------------------------------------------------
// Parameter.Copy — exercises the Copy() method
// ---------------------------------------------------------------------------

func TestParameter_Copy_AllFields(t *testing.T) {
	p := &Parameter{
		Name:      "myparam",
		Shorthand: "m",
		Usage:     "my usage",
		Type:      "string",
		Variable:  CreateVariableString("v"),
		Default:   CreateVariableString("d"),
		Hidden:    true,
	}
	c := p.Copy()
	require.NotNil(t, c)
	assert.Equal(t, p.Name, c.Name)
	assert.Equal(t, p.Shorthand, c.Shorthand)
	assert.Equal(t, p.Usage, c.Usage)
	assert.Equal(t, p.Type, c.Type)
	assert.Equal(t, *p.Variable.String, *c.Variable.String)
	assert.Equal(t, *p.Default.String, *c.Default.String)
	assert.True(t, c.Hidden)
}

// ---------------------------------------------------------------------------
// GetOrdered — alphabetical ordering guaranteed
// ---------------------------------------------------------------------------

func TestParameters_GetOrdered_Sorted(t *testing.T) {
	params := &Parameters{
		"zzz": &Parameter{Type: "string"},
		"aaa": &Parameter{Type: "string"},
		"mmm": &Parameter{Type: "string"},
	}
	got := params.GetOrdered()
	assert.Equal(t, []string{"aaa", "mmm", "zzz"}, got)
}

// ---------------------------------------------------------------------------
// SetProject / SetRegistry setter methods
// ---------------------------------------------------------------------------

func TestConfig_SetProject_And_SetRegistry(t *testing.T) {
	dir := "/TestSetProjectRegistry"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/config.yaml", []byte(""), 0o644))

	f, err := CreateFile(dir, "config.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	cfg.SetProject("proj1")
	assert.Equal(t, "proj1", cfg.project)

	cfg.SetRegistry("ghcr.io/myorg")
	assert.Equal(t, "ghcr.io/myorg", cfg.registry)
}

// ---------------------------------------------------------------------------
// MergeValues — key present in src but nil Variable is skipped
// ---------------------------------------------------------------------------

func TestMergeValues_NilVariableSkipped(t *testing.T) {
	dest := &Parameters{
		"key": &Parameter{Type: "string", Variable: CreateVariableString("old")},
	}
	src := &Parameters{
		"key": &Parameter{Type: "string"}, // Variable is nil
	}
	dest.MergeValues(src)
	// Nil variable should not overwrite existing.
	assert.Equal(t, "old", *dest.Get("key").Variable.String)
}

// ---------------------------------------------------------------------------
// ExtractValues — exercises all viper GetString/GetBool/GetInt paths
// ---------------------------------------------------------------------------

func TestExtractValues_AllTypes(t *testing.T) {
	dir := "/TestExtractValues_AllTypes"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	yaml := "name: bob\nenabled: true\ncount: 3\n"
	require.NoError(t, afero.WriteFile(app.FS, dir+"/config.yaml", []byte(yaml), 0o644))

	f, err := CreateFile(dir, "config.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	vals := cfg.ExtractValues()
	assert.NotNil(t, vals["name"])
	assert.Equal(t, "bob", *vals["name"].String)
	assert.NotNil(t, vals["enabled"])
	assert.Equal(t, true, *vals["enabled"].Bool)
}

// ---------------------------------------------------------------------------
// GetOrSetSeed — seed generation when viper has no seed
// ---------------------------------------------------------------------------

func TestGetOrSetSeed_GeneratesNewSeed(t *testing.T) {
	dir := "/TestGetOrSetSeed_Generate"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/config.yaml", []byte(""), 0o644))

	file, err := CreateFile(dir, "config.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(file)
	require.NoError(t, err)

	// A fresh empty config has no seed, so ConfigFromFile generates one.
	assert.Len(t, cfg.GetSeed(), 16)
}

// ---------------------------------------------------------------------------
// SetSeed / GetSeed / CreateSeed
// ---------------------------------------------------------------------------

func TestConfig_SetSeedAndGetSeed(t *testing.T) {
	dir := "/TestConfig_SetSeedGetSeed"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/config.yaml", []byte(""), 0o644))

	f, err := CreateFile(dir, "config.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	cfg.SetSeed("fixed-seed-value")
	assert.Equal(t, "fixed-seed-value", cfg.GetSeed())
}

// ---------------------------------------------------------------------------
// extractParamsWithTracking — recursive include error (line 254)
// A→B, B includes C which doesn't exist, causing error propagation up.
// ---------------------------------------------------------------------------

func TestExtractParamsWithTracking_RecursiveIncludeError(t *testing.T) {
	dir := "/TestRecursiveIncludeError"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	// A includes B, B includes C (nonexistent) — error should propagate.
	fileA := `
paramA:
  usage: a
  type: string
  default: a
includes:
  - b.yaml
`
	fileB := `
paramB:
  usage: b
  type: string
  default: b
includes:
  - nonexistent.yaml
`
	require.NoError(t, afero.WriteFile(app.FS, dir+"/a.yaml", []byte(fileA), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/b.yaml", []byte(fileB), 0o644))

	f, err := CreateFile(dir, "a.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	_, _, err = cfg.ExtractParams()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ExtractEmbedTo — nested directory extraction (exercises the recursive subdir
// path in getAllFilesEmbed AND the MkdirAll + WriteFile loop in ExtractEmbedTo)
// ---------------------------------------------------------------------------

func TestExtractEmbedTo_WithNestedTestdata(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	outputFolder := "/TestExtractEmbedTo_Nested"
	err := ExtractEmbedTo("testdata", testdataEmbed, outputFolder)
	require.NoError(t, err)

	// Top-level fixture file.
	exists, err := afero.Exists(app.FS, outputFolder+"/fixture.txt")
	require.NoError(t, err)
	assert.True(t, exists)

	// Nested fixture file.
	exists, err = afero.Exists(app.FS, outputFolder+"/sub/nested.txt")
	require.NoError(t, err)
	assert.True(t, exists)
}

// ---------------------------------------------------------------------------
// CompleteConfigKeysForSetFunc — UsesConfigFlag path (no-equals)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// NewRelease — exercises user.Current() path (75 %)
// The log.Fatal branch (user.Current() error) cannot be unit tested.
// This test ensures the normal path is exercised.
// ---------------------------------------------------------------------------

func TestNewRelease_NormalPath(t *testing.T) {
	r := NewRelease("myname", "/some/location", "myseed", true, false)
	assert.Equal(t, "myname", r.Name)
	assert.Equal(t, "/some/location", r.Location)
	assert.Equal(t, "myseed", r.Seed)
	assert.True(t, r.IsUpgrade)
	assert.False(t, r.IsInstall)
	assert.NotEmpty(t, r.User)
	assert.NotEmpty(t, r.Group)
	assert.NotEmpty(t, r.Service)
}

// ---------------------------------------------------------------------------
// extractParamsWithTracking — depth limit exceeded (MaxIncludeDepth)
// Creates a linear chain of includes: a→b→c→…→(depth+1) and calls ExtractParams.
// ---------------------------------------------------------------------------

func TestExtractParamsWithTracking_DepthLimitExceeded(t *testing.T) {
	dir := "/TestDepthLimit"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	// Build a chain of MaxIncludeDepth+2 files so we exceed the limit.
	depth := 12 // > MaxIncludeDepth (10)
	for i := 0; i < depth; i++ {
		var content string
		if i < depth-1 {
			content = "param" + fmt.Sprintf("%d", i) +
				":\n  usage: u\n  type: string\n  default: v\nincludes:\n  - file" + fmt.Sprintf("%d", i+1) + ".yaml\n"
		} else {
			content = "paramLast:\n  usage: u\n  type: string\n  default: v\n"
		}
		fname := dir + "/file" + fmt.Sprintf("%d", i) + ".yaml"
		require.NoError(t, afero.WriteFile(app.FS, fname, []byte(content), 0o644))
	}

	f, err := CreateFile(dir, "file0.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	_, _, err = cfg.ExtractParams()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum include depth")
}

// ---------------------------------------------------------------------------
// extractParamsWithTracking — circular include detection
// a.yaml includes b.yaml, b.yaml includes a.yaml → cycle detected.
// ---------------------------------------------------------------------------

func TestExtractParamsWithTracking_CircularInclude(t *testing.T) {
	dir := "/TestCircularInclude"
	require.NoError(t, app.FS.MkdirAll(dir, 0o755))

	fileA := "paramA:\n  usage: a\n  type: string\n  default: x\nincludes:\n  - b.yaml\n"
	fileB := "paramB:\n  usage: b\n  type: string\n  default: y\nincludes:\n  - a.yaml\n"

	require.NoError(t, afero.WriteFile(app.FS, dir+"/a.yaml", []byte(fileA), 0o644))
	require.NoError(t, afero.WriteFile(app.FS, dir+"/b.yaml", []byte(fileB), 0o644))

	f, err := CreateFile(dir, "a.yaml")
	require.NoError(t, err)
	cfg, err := ConfigFromFile(f)
	require.NoError(t, err)

	_, _, err = cfg.ExtractParams()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular include")
}
