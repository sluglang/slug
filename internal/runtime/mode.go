package runtime

import (
	"fmt"
	"strings"
)

const (
	// RuntimeTreewalk is retained as a legacy alias during treewalker removal.
	RuntimeTreewalk = "treewalk"
	RuntimeVM       = "vm"
)

func NormalizeRuntimeMode(mode string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return RuntimeVM, nil
	}
	switch normalized {
	case RuntimeVM:
		return RuntimeVM, nil
	case RuntimeTreewalk:
		return RuntimeVM, nil
	default:
		return "", fmt.Errorf("invalid runtime mode %q (expected %q)", mode, RuntimeVM)
	}
}
