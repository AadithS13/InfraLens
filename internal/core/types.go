package core

import "time"

// --- Request types ---

type SearchFilter struct {
	City     *string
	District *string
	State    *string
	Promoter *string
	Status   *string
	Type     *string
	Page     int
	Limit    int
}

func (f *SearchFilter) Offset() int {
	if f.Page <= 0 {
		f.Page = 1
	}
	return (f.Page - 1) * f.Limit
}

// --- Response types ---

type ProjectListItem struct {
	ID                   int        `json:"id"`
	MahaID               int        `json:"maha_id"`
	ReraRegistrationNo   string     `json:"rera_registration_no"`
	ProjectName          string     `json:"project_name"`
	ProjectType          string     `json:"project_type"`
	ProjectStatus        string     `json:"project_status"`
	ProjectCurrentStatus string     `json:"project_current_status"`
	ReraRegistrationDate *time.Time `json:"rera_registration_date"`
	ProposedCompletion   *time.Time `json:"proposed_completion_date"`
	TotalUnits           int        `json:"total_units"`
	TotalSoldUnits       int        `json:"total_sold_units"`
	PromoterName         string     `json:"promoter_name"`
	City                 string     `json:"city"`
	District             string     `json:"district"`
	State                string     `json:"state"`
	Pincode              string     `json:"pincode"`
}

type ProjectAddress struct {
	Line1   string `json:"line1"`
	Line2   string `json:"line2"`
	City    string `json:"city"`
	District string `json:"district"`
	State   string `json:"state"`
	Pincode string `json:"pincode"`
}

type PromoterDetail struct {
	Name         string         `json:"name"`
	Pan          string         `json:"pan"`
	Gstin        string         `json:"gstin"`
	PromoterType string         `json:"promoter_type"`
	Address      *ProjectAddress `json:"address,omitempty"`
}

type ContactItem struct {
	Role  string `json:"role"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type ProjectDetail struct {
	ProjectListItem
	AcknowledgementNumber  string         `json:"acknowledgement_number"`
	OriginalCompletion     *time.Time     `json:"original_completion_date"`
	IsMigrated             bool           `json:"is_migrated"`
	ProjectAddress         *ProjectAddress `json:"project_address,omitempty"`
	Promoter               *PromoterDetail `json:"promoter,omitempty"`
	Contacts               []ContactItem  `json:"contacts"`
}

type ChangeItem struct {
	FieldName  string    `json:"field_name"`
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	DetectedAt time.Time `json:"detected_at"`
}

// --- Generic list response ---

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type ListResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}
