package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infralens/infralens/internal/model"
)

type Store struct {
	db *pgxpool.Pool
}

func New(dsn string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return &Store{db: pool}, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) UpsertPromoter(ctx context.Context, p *model.Promoter) (int, error) {
	raw, _ := json.Marshal(p.RawJSON)
	var id int
	err := s.db.QueryRow(ctx, `
		INSERT INTO promoters (user_profile_id, name, pan, gstin, promoter_type, raw_json, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_profile_id) DO UPDATE SET
			name         = EXCLUDED.name,
			pan          = EXCLUDED.pan,
			gstin        = EXCLUDED.gstin,
			promoter_type = EXCLUDED.promoter_type,
			raw_json     = EXCLUDED.raw_json,
			updated_at   = NOW()
		RETURNING id`,
		p.UserProfileID, p.Name, p.Pan, p.Gstin, p.PromoterType, raw,
	).Scan(&id)
	return id, err
}

func (s *Store) UpsertProject(ctx context.Context, p *model.Project) (int, error) {
	raw, _ := json.Marshal(p.RawJSON)
	var id int
	err := s.db.QueryRow(ctx, `
		INSERT INTO projects (
			maha_id, rera_registration_no, acknowledgement_number,
			project_name, project_type, project_status, project_current_status,
			rera_registration_date, proposed_completion_date, original_completion_date,
			total_units, total_sold_units, user_profile_id, promoter_id,
			is_migrated, raw_json, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,NOW()
		)
		ON CONFLICT (maha_id) DO UPDATE SET
			rera_registration_no     = EXCLUDED.rera_registration_no,
			project_name             = EXCLUDED.project_name,
			project_type             = EXCLUDED.project_type,
			project_status           = EXCLUDED.project_status,
			project_current_status   = EXCLUDED.project_current_status,
			rera_registration_date   = EXCLUDED.rera_registration_date,
			proposed_completion_date = EXCLUDED.proposed_completion_date,
			original_completion_date = EXCLUDED.original_completion_date,
			total_units              = EXCLUDED.total_units,
			total_sold_units         = EXCLUDED.total_sold_units,
			promoter_id              = EXCLUDED.promoter_id,
			raw_json                 = EXCLUDED.raw_json,
			updated_at               = NOW()
		RETURNING id`,
		p.MahaID, p.ReraRegistrationNo, p.AcknowledgementNumber,
		p.ProjectName, p.ProjectType, p.ProjectStatus, p.ProjectCurrentStatus,
		p.ReraRegistrationDate, p.ProposedCompletionDate, p.OriginalCompletionDate,
		p.TotalUnits, p.TotalSoldUnits, p.UserProfileID, p.PromoterID,
		p.IsMigrated, raw,
	).Scan(&id)
	return id, err
}

func (s *Store) InsertAddress(ctx context.Context, a *model.Address) error {
	raw, _ := json.Marshal(a.RawJSON)
	_, err := s.db.Exec(ctx, `
		INSERT INTO addresses (entity_type, entity_id, line1, line2, city, district, state, pincode, raw_json)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.EntityType, a.EntityID, a.Line1, a.Line2, a.City, a.District, a.State, a.Pincode, raw,
	)
	return err
}

func (s *Store) InsertContacts(ctx context.Context, contacts []model.Contact) error {
	if len(contacts) == 0 {
		return nil
	}
	rows := make([][]any, len(contacts))
	for i, c := range contacts {
		raw, _ := json.Marshal(c.RawJSON)
		rows[i] = []any{c.ProjectID, c.Role, c.Name, c.Phone, c.Email, raw}
	}
	_, err := s.db.CopyFrom(ctx,
		pgx.Identifier{"contacts"},
		[]string{"project_id", "role", "name", "phone", "email", "raw_json"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (s *Store) InsertSnapshot(ctx context.Context, snap *model.ProjectSnapshot) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO project_snapshots (project_id, fetched_at, checksum, raw_json)
		VALUES ($1, $2, $3, $4)`,
		snap.ProjectID, snap.FetchedAt, snap.Checksum, snap.RawJSON,
	)
	return err
}

func (s *Store) SnapshotExists(ctx context.Context, projectID int, checksum string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM project_snapshots WHERE project_id=$1 AND checksum=$2)`,
		projectID, checksum,
	).Scan(&exists)
	return exists, err
}

// GetLatestSnapshot returns the most recent snapshot for a project, or nil if none exists.
func (s *Store) GetLatestSnapshot(ctx context.Context, projectID int) (*model.ProjectSnapshot, error) {
	var snap model.ProjectSnapshot
	err := s.db.QueryRow(ctx, `
		SELECT id, project_id, fetched_at, checksum, raw_json
		FROM project_snapshots
		WHERE project_id = $1
		ORDER BY fetched_at DESC
		LIMIT 1`,
		projectID,
	).Scan(&snap.ID, &snap.ProjectID, &snap.FetchedAt, &snap.Checksum, &snap.RawJSON)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

// InsertCrawlRun creates a new crawl_run record and returns its ID.
func (s *Store) InsertCrawlRun(ctx context.Context, startID, endID int) (int, error) {
	var id int
	err := s.db.QueryRow(ctx, `
		INSERT INTO crawl_runs (start_id, end_id)
		VALUES ($1, $2)
		RETURNING id`,
		startID, endID,
	).Scan(&id)
	return id, err
}

// UpdateCrawlRun marks a crawl run as finished with final stats.
func (s *Store) UpdateCrawlRun(ctx context.Context, id, processed, failed int, status string, crawlErr error) error {
	var errMsg *string
	if crawlErr != nil {
		msg := crawlErr.Error()
		errMsg = &msg
	}
	_, err := s.db.Exec(ctx, `
		UPDATE crawl_runs
		SET finished_at = NOW(), status = $2, processed = $3, failed = $4, error = $5
		WHERE id = $1`,
		id, status, processed, failed, errMsg,
	)
	return err
}

// ListCrawlRuns returns the most recent crawl runs.
func (s *Store) ListCrawlRuns(ctx context.Context, limit int) ([]model.CrawlRun, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, started_at, finished_at, status, start_id, end_id, processed, failed, error
		FROM crawl_runs
		ORDER BY started_at DESC
		LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []model.CrawlRun
	for rows.Next() {
		var r model.CrawlRun
		if err := rows.Scan(&r.ID, &r.StartedAt, &r.FinishedAt, &r.Status,
			&r.StartID, &r.EndID, &r.Processed, &r.Failed, &r.Error); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// InsertProjectChanges batch-inserts detected field-level changes.
func (s *Store) InsertProjectChanges(ctx context.Context, changes []model.ProjectChange) error {
	if len(changes) == 0 {
		return nil
	}
	rows := make([][]any, len(changes))
	for i, c := range changes {
		rows[i] = []any{c.ProjectID, c.FieldName, c.OldValue, c.NewValue, c.DetectedAt}
	}
	_, err := s.db.CopyFrom(ctx,
		pgx.Identifier{"project_changes"},
		[]string{"project_id", "field_name", "old_value", "new_value", "detected_at"},
		pgx.CopyFromRows(rows),
	)
	return err
}
