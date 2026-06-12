package notifier

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/infralens/infralens/internal/model"
)

// LogAdapter prints a formatted [NOTIFY] block to the process log.
// It is always registered — Phase 1 behaviour.
type LogAdapter struct{}

func NewLogAdapter() *LogAdapter { return &LogAdapter{} }

func (a *LogAdapter) Name() string { return "log" }

func (a *LogAdapter) Send(_ context.Context, c model.NotifiableChange) error {
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
	return nil
}
