package version

import "testing"

func TestString_defaults(t *testing.T) {
	// With default build-time values
	want := "dev (none) unknown"
	got := String()
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestString_with_build_values(t *testing.T) {
	origV, origC, origD := Version, Commit, Date
	t.Cleanup(func() { Version, Commit, Date = origV, origC, origD })

	Version = "1.2.3"
	Commit = "abc1234"
	Date = "2025-01-15"

	want := "1.2.3 (abc1234) 2025-01-15"
	got := String()
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
