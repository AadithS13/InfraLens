package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/infralens/infralens/internal/model"
)

// WebhookAdapter delivers notifications as an HTTP POST with a JSON payload.
// Any endpoint that accepts POST JSON can receive InfraLens alerts.
type WebhookAdapter struct {
	url    string
	client *http.Client
}

func NewWebhookAdapter(url string) *WebhookAdapter {
	return &WebhookAdapter{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *WebhookAdapter) Name() string { return "webhook" }

type webhookPayload struct {
	ProjectName string `json:"project_name"`
	Field       string `json:"field"`
	FieldLabel  string `json:"field_label"`
	OldValue    string `json:"old_value"`
	NewValue    string `json:"new_value"`
	DetectedAt  string `json:"detected_at"`
}

func (a *WebhookAdapter) Send(ctx context.Context, c model.NotifiableChange) error {
	label := fieldLabels[c.FieldName]
	if label == "" {
		label = c.FieldName
	}

	payload := webhookPayload{
		ProjectName: c.ProjectName,
		Field:       c.FieldName,
		FieldLabel:  label,
		OldValue:    c.OldValue,
		NewValue:    c.NewValue,
		DetectedAt:  c.DetectedAt.UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "InfraLens-Notifier/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
