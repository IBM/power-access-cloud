package service

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type Interface interface {
	// Reconcile reconciles a service
	Reconcile(ctx context.Context) (ctrl.Result, error)
	// Delete deletes a service
	Delete(ctx context.Context) (bool, error)
}
