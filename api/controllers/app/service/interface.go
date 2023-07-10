package service

import (
	"context"
)

type Interface interface {
	// Create creates a service
	Reconcile(ctx context.Context) error
	// Delete deletes a service
	Delete(ctx context.Context) (bool, error)
}
