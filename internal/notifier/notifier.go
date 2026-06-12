package notifier

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/infralens/infralens/internal/model"
	"github.com/infralens/infralens/internal/store"
)

// watchedFields are the only fields that trigger notifications.
// These have direct business value — status shifts, deadline slips, unit updates.
var watchedFields = []string{
	"project_status",
	"project_current_status",
	"proposed_completion_date",
}

// fieldLabels makes the terminal output human-readable.
var fieldLabels = map[string]string{
	"project_status":           "Project Status",
	"project_current_status":   "Current Status",
	"proposed_completion_date": "Completion Date",
}

type Notifier struct {
	store *store.Store
}

func New(s *store.Store) *Notifier {
	return &Notifier{store: s}
}

// Notify finds all unnotified changes for watched fields, logs them, and marks them done.
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
		n.logChange(c)
		notified = append(notified, c.ID)
	}

	if err := n.store.MarkChangesNotified(ctx, notified); err != nil {
		return fmt.Errorf("mark changes notified: %w", err)
	}

	log.Printf("[NOTIFIER] dispatched %d notification(s)", len(notified))
	return nil
}

func (n *Notifier) logChange(c model.NotifiableChange) {
	label := fieldLabels[c.FieldName]
	if label == "" {
		label = c.FieldName
	}

	border := strings.Repeat("─", 50)
	log.Printf("\n[NOTIFY] %s\n  Project : %s\n  Field   : %s\n  Old     : %s\n  New     : %s\n  Detected: %s\n%s",
		border,
		c.ProjectName,
		label,
		c.OldValue,
		c.NewValue,
		c.DetectedAt.UTC().Format(time.RFC3339),
		border,
	)
}
