package handlers

import (
    "os"
    "testing"

    "github.com/spf13/afero"
    "stamus-ctl/internal/app"
)

func TestSplitImageRef(t *testing.T) {
    cases := []struct {
        in       string
        wantReg  string
        wantName string
    }{
        {"ghcr.io/org/repo:tag", "ghcr.io/org/", "repo:tag"},
        {"jasonish/suricata:master-amd64", "jasonish/", "suricata:master-amd64"},
        {"repo:tag", "", "repo:tag"},
        {"localhost:5000/repo:tag", "localhost:5000/", "repo:tag"},
        {"myreg.io/team/app@sha256:deadbeef", "myreg.io/team/", "app@sha256:deadbeef"},
    }
    for _, c := range cases {
        reg, name := splitImageRef(c.in)
        if reg != c.wantReg || name != c.wantName {
            t.Fatalf("splitImageRef(%q) = (%q,%q), want (%q,%q)", c.in, reg, name, c.wantReg, c.wantName)
        }
    }
}

func TestResolveSuricataImage_FromCompose(t *testing.T) {
    // Ensure tests use in-memory FS
    app.FS = afero.NewMemMapFs()

    // Create instance path and compose file
    instPath := "/instance-a"
    if err := app.FS.MkdirAll(instPath, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }

    compose := `services:
  suricata:
    image: ghcr.io/stamusnetworks/suricata:1.2.3
  other:
    image: busybox
`
    if err := afero.WriteFile(app.FS, instPath+"/docker-compose.yaml", []byte(compose), 0o644); err != nil {
        t.Fatalf("write compose: %v", err)
    }

    got, err := resolveSuricataImage(instPath)
    if err != nil {
        t.Fatalf("resolveSuricataImage error: %v", err)
    }
    want := "ghcr.io/stamusnetworks/suricata:1.2.3"
    if got != want {
        t.Fatalf("resolveSuricataImage got %q, want %q", got, want)
    }
}

func TestResolveSuricataImage_EnvOverride(t *testing.T) {
    // Ensure tests use in-memory FS
    app.FS = afero.NewMemMapFs()

    // Set env override
    os.Setenv("STAMUSCTL_SURICATA_IMAGE", "override/repo:latest")
    defer os.Unsetenv("STAMUSCTL_SURICATA_IMAGE")

    got, err := resolveSuricataImage("/does-not-matter")
    if err != nil {
        t.Fatalf("resolveSuricataImage error with env override: %v", err)
    }
    if got != "override/repo:latest" {
        t.Fatalf("env override not applied, got %q", got)
    }
}

func TestResolveSuricataImage_Fallback(t *testing.T) {
    // Ensure tests use in-memory FS
    app.FS = afero.NewMemMapFs()

    // No compose present -> fallback to defaultSuricataImage
    got, _ := resolveSuricataImage("/missing-instance")
    if got != defaultSuricataImage {
        t.Fatalf("expected fallback to %q, got %q", defaultSuricataImage, got)
    }
}

