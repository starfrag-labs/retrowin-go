package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/starfrag-lab/retrowin-go/internal/core/object"
)

const (
	// DefaultPendingExpiry is the default time after which pending objects are considered expired.
	DefaultPendingExpiry = 24 * time.Hour
)

// GarbageCollector handles cleanup of expired pending objects.
type GarbageCollector struct {
	objectSvc object.ObjectService
	expiry    time.Duration
}

// NewGarbageCollector creates a new garbage collector.
func NewGarbageCollector(objectSvc object.ObjectService, expiry time.Duration) *GarbageCollector {
	if expiry == 0 {
		expiry = DefaultPendingExpiry
	}
	return &GarbageCollector{
		objectSvc: objectSvc,
		expiry:    expiry,
	}
}

// Run performs garbage collection of expired pending objects.
// Returns the number of objects cleaned up.
func (gc *GarbageCollector) Run(ctx context.Context) (int, error) {
	pendingObjects, err := gc.objectSvc.FindPendingOlderThan(ctx, gc.expiry)
	if err != nil {
		return 0, err
	}

	cleaned := 0
	for _, obj := range pendingObjects {
		if err := gc.objectSvc.Delete(ctx, obj.ID()); err != nil {
			slog.Warn("failed to delete expired pending object",
				"object_id", obj.ID(),
				"error", err,
			)
			continue
		}
		cleaned++
		slog.Info("cleaned up expired pending object",
			"object_id", obj.ID(),
			"age", time.Since(obj.CreatedAt()),
		)
	}

	return cleaned, nil
}

// RunPeriodically starts periodic garbage collection at the given interval.
// This is a blocking function that runs until the context is cancelled.
func (gc *GarbageCollector) RunPeriodically(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleaned, err := gc.Run(ctx)
			if err != nil {
				slog.Error("garbage collection failed", "error", err)
			} else if cleaned > 0 {
				slog.Info("garbage collection completed", "cleaned", cleaned)
			}
		}
	}
}
