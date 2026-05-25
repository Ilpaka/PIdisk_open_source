package usecase

import (
	"errors"
	"testing"

	"github.com/pidisk/pidisk/internal/domain"
)

func TestValidatePath(t *testing.T) {
	cases := []struct {
		in   string
		want error
	}{
		{"/srv/data", nil},
		{"/srv/../etc/passwd", domain.ErrPathTraversal},
		{"", domain.ErrInvalidPath},
		{"/srv/with\x00null", domain.ErrInvalidPath},
	}
	for _, tc := range cases {
		err := validatePath(tc.in)
		if !errors.Is(err, tc.want) {
			t.Fatalf("validatePath(%q) = %v, want %v", tc.in, err, tc.want)
		}
	}
}

func TestValidateName(t *testing.T) {
	if err := validateName(""); err == nil {
		t.Fatalf("empty name should fail")
	}
	if err := validateName("good_name.txt"); err != nil {
		t.Fatalf("good name unexpectedly failed: %v", err)
	}
	if err := validateName("with/slash"); err == nil {
		t.Fatalf("slash should be rejected")
	}
}

func TestParseDf(t *testing.T) {
	out := "Filesystem     1K-blocks      Used Available Use% Mounted on\n" +
		"/dev/sda1      102400000  51200000  51200000  50% /"
	du, err := parseDf(out)
	if err != nil {
		t.Fatalf("parseDf: %v", err)
	}
	if du.Total == 0 || du.Used == 0 || du.Free == 0 {
		t.Fatalf("expected non-zero values, got %+v", du)
	}
	if du.Percent < 49 || du.Percent > 51 {
		t.Fatalf("percent off: %f", du.Percent)
	}
}
