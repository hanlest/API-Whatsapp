package storage

import "testing"

func TestSanitizeDeviceFolder(t *testing.T) {
	if got := SanitizeDeviceFolder("Estar Siempre", ""); got != "Estar-Siempre" {
		t.Fatalf("got %q", got)
	}
	if got := SanitizeDeviceFolder("dev", "Custom/Name"); got != "CustomName" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatParseMC(t *testing.T) {
	ref := FormatMC("b1", "folder/incoming/a.jpg")
	parsed, err := ParseMC(ref)
	if err != nil || parsed.Bucket != "b1" || parsed.ObjectKey != "folder/incoming/a.jpg" {
		t.Fatalf("parse failed: %v %+v", err, parsed)
	}
}
