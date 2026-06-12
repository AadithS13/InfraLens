package notifier

import (
	"context"
	"fmt"
	"log"

	"github.com/infralens/infralens/internal/store"
)

// watchedFields are the only fields that trigger notifications.
// These have direct business value — status shifts, approval milestones, deadline slips.
var watchedFields = []string{
	"project_status",
	"project_current_status",
	"proposed_completion_date",
}

// fieldLabels makes output human-readable across all adapters.
var fieldLabels = map[string]string{
	"project_status":           "Project Status",
	"project_current_status":   "Current Status",
	"proposed_completion_date": "Completion Date",
}

// Notifier fetches unnotified changes and fans them out to every registered adapter.
type Notifier struct {
	store    *store.Store
	adapters []Adapter
}

// New creates a Notifier. Pass at least one adapter — LogAdapter is always recommended.
// adapters are tried in order; a failed adapter logs a warning but does not block others.
func New(s *store.Store, adapters ...Adapter) *Notifier {
	return &Notifier{store: s, adapters: adapters}
}

// Notify finds all unnotified changes for watched fields, fans out to all adapters,
// then marks the changes as notified regardless of adapter failures (best-effort delivery).
func (n *Notifier) Notify(ctx context.Context) error {
	changes, err := n.store.GetUnnotifiedChanges(ctx, watchedFields)
	if err != nil {
		return fmt.Errorf("fetch unnotified changes: %w", err)
	}
	if len(changes) == 0 {
		return nil
	}

	var notified []int
	for _, c := range changes {
		for _, a := range n.adapters {
			if err := a.Send(ctx, c); err != nil {
				log.Printf("[NOTIFIER] adapter %q failed for change %d: %v", a.Name(), c.ID, err)
			}
		}
		notified = append(notified, c.ID)
	}

	if err := n.store.MarkChangesNotified(ctx, notified); err != nil {
		return fmt.Errorf("mark changes notified: %w", err)
	}

	log.Printf("[NOTIFIER] dispatched %d notification(s) via %d adapter(s)", len(notified), len(n.adapters))
	return nil
}
