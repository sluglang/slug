package runtime

import (
	"fmt"
	"strings"
)

const (
	RuntimeTreewalk = "treewalk"
	RuntimeVM       = "vm"
)

func NormalizeRuntimeMode(mode string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return RuntimeTreewalk, nil
	}
	switch normalized {
	case RuntimeTreewalk, RuntimeVM:
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid runtime mode %q (expected %q or %q)", mode, RuntimeTreewalk, RuntimeVM)
	}
}
