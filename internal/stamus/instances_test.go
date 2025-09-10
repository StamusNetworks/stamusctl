package stamus

import "testing"

func TestIsComposeUpFromPs_JSONRunning(t *testing.T) {
	out := `[
        {"Name":"a","State":"running","Status":"running"},
        {"Name":"b","State":"exited","Status":"exited"}
    ]`
	if !isComposeUpFromPs(out) {
		t.Fatalf("expected running=true for JSON output")
	}
}

func TestIsComposeUpFromPs_JSONExited(t *testing.T) {
	out := `[{"Name":"a","State":"exited","Status":"exited"}]`
	if isComposeUpFromPs(out) {
		t.Fatalf("expected running=false for exited JSON output")
	}
}

func TestIsComposeUpFromPs_TextUp(t *testing.T) {
	out := "NAME\tSERVICE\tSTATUS\nweb\tweb\tUp 2 minutes"
	if !isComposeUpFromPs(out) {
		t.Fatalf("expected running=true for text containing Up")
	}
}

func TestIsComposeUpFromPs_TextRunning(t *testing.T) {
	out := "NAME SERVICE STATUS\nweb web running 2m"
	if !isComposeUpFromPs(out) {
		t.Fatalf("expected running=true for text containing running")
	}
}

func TestIsComposeUpFromPs_Empty(t *testing.T) {
	if isComposeUpFromPs("") {
		t.Fatalf("expected running=false for empty output")
	}
}

func TestIsComposeUpFromPs_TextNoRunning(t *testing.T) {
	out := "NAME\tSERVICE\tSTATUS\nweb\tweb\tExited (0) 3 minutes ago"
	if isComposeUpFromPs(out) {
		t.Fatalf("expected running=false for exited text output")
	}
}

func TestIsComposeUpFromPs_JSONLines(t *testing.T) {
	out := `{"Name":"a","State":"exited","Status":"exited"}
{"Name":"b","State":"running","Status":"running"}`
	if !isComposeUpFromPs(out) {
		t.Fatalf("expected running=true for JSON-per-line output")
	}
}
