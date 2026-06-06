package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infralens/infralens/internal/core"
)

type ProjectRepo struct {
	db *pgxpool.Pool
}

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return pool, nil
}

// Search returns a paginated list of projects matching the given filters.
func (r *ProjectRepo) Search(ctx context.Context, f core.SearchFilter) ([]core.ProjectListItem, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			p.id, p.maha_id, p.rera_registration_no, p.project_name,
			COALESCE(p.project_type, ''), COALESCE(p.project_status, ''),
			COALESCE(p.project_current_status, ''),
			p.rera_registration_date, p.proposed_completion_date,
			p.total_units, p.total_sold_units,
			COALESCE(pr.name, '')    AS promoter_name,
			COALESCE(a.city, '')     AS city,
			COALESCE(a.district, '') AS district,
			COALESCE(a.state, '')    AS state,
			COALESCE(a.pincode, '')  AS pincode,
			COUNT(*) OVER()          AS total_count
		FROM projects p
		LEFT JOIN promoters pr ON p.promoter_id = pr.id
		LEFT JOIN addresses a  ON a.entity_type = 'project' AND a.entity_id = p.id
		WHERE ($1::text IS NULL OR a.city     ILIKE '%' || $1 || '%')
		  AND ($2::text IS NULL OR a.district ILIKE '%' || $2 || '%')
		  AND ($3::text IS NULL OR a.state    ILIKE '%' || $3 || '%')
		  AND ($4::text IS NULL OR pr.name    ILIKE '%' || $4 || '%')
		  AND ($5::text IS NULL OR p.project_status ILIKE $5)
		  AND ($6::text IS NULL OR p.project_type   ILIKE '%' || $6 || '%')
		ORDER BY p.id DESC
		LIMIT $7 OFFSET $8`,
		f.City, f.District, f.State, f.Promoter, f.Status, f.Type,
		f.Limit, f.Offset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var (
		items []core.ProjectListItem
		total int
	)
	for rows.Next() {
		var item core.ProjectListItem
		if err := rows.Scan(
			&item.ID, &item.MahaID, &item.ReraRegistrationNo, &item.ProjectName,
			&item.ProjectType, &item.ProjectStatus, &item.ProjectCurrentStatus,
			&item.ReraRegistrationDate, &item.ProposedCompletion,
			&item.TotalUnits, &item.TotalSoldUnits,
			&item.PromoterName, &item.City, &item.District, &item.State, &item.Pincode,
			&total,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// GetByID returns full project detail including promoter, addresses, and contacts.
func (r *ProjectRepo) GetByID(ctx context.Context, id int) (*core.ProjectDetail, error) {
	var d core.ProjectDetail
	var (
		projAddr core.ProjectAddress
		promAddr core.ProjectAddress
		promoter core.PromoterDetail
	)

	err := r.db.QueryRow(ctx, `
		SELECT
			p.id, p.maha_id, p.rera_registration_no, p.acknowledgement_number,
			p.project_name,
			COALESCE(p.project_type, ''), COALESCE(p.project_status, ''),
			COALESCE(p.project_current_status, ''),
			p.rera_registration_date, p.proposed_completion_date, p.original_completion_date,
			p.total_units, p.total_sold_units, p.is_migrated,
			COALESCE(pr.name, ''),         COALESCE(pr.pan, ''),
			COALESCE(pr.gstin, ''),        COALESCE(pr.promoter_type, ''),
			COALESCE(pa.line1, ''),        COALESCE(pa.line2, ''),
			COALESCE(pa.city, ''),         COALESCE(pa.district, ''),
			COALESCE(pa.state, ''),        COALESCE(pa.pincode, ''),
			COALESCE(pra.line1, ''),       COALESCE(pra.line2, ''),
			COALESCE(pra.city, ''),        COALESCE(pra.district, ''),
			COALESCE(pra.state, ''),       COALESCE(pra.pincode, '')
		FROM projects p
		LEFT JOIN promoters pr  ON p.promoter_id   = pr.id
		LEFT JOIN addresses pa  ON pa.entity_type  = 'project'  AND pa.entity_id = p.id
		LEFT JOIN addresses pra ON pra.entity_type = 'promoter' AND pra.entity_id = pr.id
		WHERE p.id = $1`, id,
	).Scan(
		&d.ID, &d.MahaID, &d.ReraRegistrationNo, &d.AcknowledgementNumber,
		&d.ProjectName, &d.ProjectType, &d.ProjectStatus, &d.ProjectCurrentStatus,
		&d.ReraRegistrationDate, &d.ProposedCompletion, &d.OriginalCompletion,
		&d.TotalUnits, &d.TotalSoldUnits, &d.IsMigrated,
		&promoter.Name, &promoter.Pan, &promoter.Gstin, &promoter.PromoterType,
		&projAddr.Line1, &projAddr.Line2, &projAddr.City, &projAddr.District,
		&projAddr.State, &projAddr.Pincode,
		&promAddr.Line1, &promAddr.Line2, &promAddr.City, &promAddr.District,
		&promAddr.State, &promAddr.Pincode,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	d.ProjectAddress = &projAddr
	promoter.Address = &promAddr
	d.Promoter = &promoter

	// Fetch contacts
	contacts, err := r.getContacts(ctx, id)
	if err != nil {
		return nil, err
	}
	d.Contacts = contacts

	return &d, nil
}

func (r *ProjectRepo) getContacts(ctx context.Context, projectID int) ([]core.ContactItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT COALESCE(role,''), COALESCE(name,''), COALESCE(phone,''), COALESCE(email,'')
		FROM contacts WHERE project_id = $1 ORDER BY role`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []core.ContactItem
	for rows.Next() {
		var c core.ContactItem
		if err := rows.Scan(&c.Role, &c.Name, &c.Phone, &c.Email); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

// GetChanges returns the change history for a project, most recent first.
func (r *ProjectRepo) GetChanges(ctx context.Context, projectID int) ([]core.ChangeItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT field_name, COALESCE(old_value,''), COALESCE(new_value,''), detected_at
		FROM project_changes
		WHERE project_id = $1
		ORDER BY detected_at DESC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []core.ChangeItem
	for rows.Next() {
		var c core.ChangeItem
		if err := rows.Scan(&c.FieldName, &c.OldValue, &c.NewValue, &c.DetectedAt); err != nil {
			return nil, err
		}
		changes = append(changes, c)
	}
	return changes, rows.Err()
}
