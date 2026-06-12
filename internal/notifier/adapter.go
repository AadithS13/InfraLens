package notifier

import (
	"context"

	"github.com/infralens/infralens/internal/model"
)

// Adapter is the delivery interface. Every notification channel implements this.
// Send is called once per NotifiableChange — return an error to signal delivery failure.
// A failed adapter does NOT block other adapters or prevent the change from being marked notified.
type Adapter interface {
	Name() string
	Send(ctx context.Context, c model.NotifiableChange) error
}
