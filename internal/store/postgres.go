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
