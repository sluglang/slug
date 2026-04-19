package util

import (
	"strings"
	"testing"
)

func TestGetLineAndColumnMidRunePosition(t *testing.T) {
	src := "a\néx\nz"
	line, col := GetLineAndColumn(src, 3) // middle byte of 'é'
	if line != 2 || col != 1 {
		t.Fatalf("expected line 2 col 1, got line %d col %d", line, col)
	}
}

func TestGetContextLinesUnicodeSafeColumn(t *testing.T) {
	src := "first\néxample\nthird\n"
	ctx := GetContextLines(src, 2, 2)
	if !strings.Contains(ctx, "^ unexpected here") {
		t.Fatalf("expected caret marker, got %q", ctx)
	}
}
