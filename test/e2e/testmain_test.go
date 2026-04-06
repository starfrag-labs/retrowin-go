package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	if err := startSharedContainers(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start shared containers: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	stopSharedContainers(context.Background())
	os.Exit(code)
}
