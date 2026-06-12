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
// When f.Q is set, results are ranked by word_similarity (pg_trgm) across
// project name, promoter name, and district. Each result includes a relevance
// score (0–1) so the caller can show how well each result matched.
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
			CASE WHEN $9::text IS NULL THEN 0::float8 ELSE
				GREATEST(
					word_similarity($9, COALESCE(p.project_name, '')),
					word_similarity($9, COALESCE(pr.name, '')),
					word_similarity($9, COALESCE(a.district, ''))
				)
			END AS relevance,
			COUNT(*) OVER() AS total_count
		FROM projects p
		LEFT JOIN promoters pr ON p.promoter_id = pr.id
		LEFT JOIN addresses a  ON a.entity_type = 'project' AND a.entity_id = p.id
		WHERE ($1::text IS NULL OR a.city     ILIKE '%' || $1 || '%')
		  AND ($2::text IS NULL OR a.district ILIKE '%' || $2 || '%')
		  AND ($3::text IS NULL OR a.state    ILIKE '%' || $3 || '%')
		  AND ($4::text IS NULL OR pr.name    ILIKE '%' || $4 || '%')
		  AND ($5::text IS NULL OR p.project_status ILIKE $5)
		  AND ($6::text IS NULL OR p.project_type   ILIKE '%' || $6 || '%')
		  AND ($9::text IS NULL OR (
		      word_similarity($9, COALESCE(p.project_name, '')) > 0.1
		      OR word_similarity($9, COALESCE(pr.name, ''))      > 0.1
		      OR word_similarity($9, COALESCE(a.district, ''))   > 0.1
		  ))
		ORDER BY relevance DESC, p.id DESC
		LIMIT $7 OFFSET $8`,
		f.City, f.District, f.State, f.Promoter, f.Status, f.Type,
		f.Limit, f.Offset(), f.Q,
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
		var (
			item      core.ProjectListItem
			relevance float64
		)
		if err := rows.Scan(
			&item.ID, &item.MahaID, &item.ReraRegistrationNo, &item.ProjectName,
			&item.ProjectType, &item.ProjectStatus, &item.ProjectCurrentStatus,
			&item.ReraRegistrationDate, &item.ProposedCompletion,
			&item.TotalUnits, &item.TotalSoldUnits,
			&item.PromoterName, &item.City, &item.District, &item.State, &item.Pincode,
			&relevance,
			&total,
		); err != nil {
			return nil, 0, err
		}
		if relevance > 0 {
			item.Relevance = &relevance
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// Suggestions returns autocomplete results for a partial query.
// Returns ranked project names and promoter names using word_similarity.
func (r *ProjectRepo) Suggestions(ctx context.Context, q string, limit int) ([]core.SuggestionItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT text, type, score FROM (
			(
				SELECT p.project_name AS text, 'project' AS type,
				       word_similarity($1, p.project_name) AS score
				FROM projects p
				WHERE p.project_name IS NOT NULL
				  AND word_similarity($1, p.project_name) > 0.15
				ORDER BY score DESC
				LIMIT 8
			)
			UNION
			(
				SELECT pr.name AS text, 'promoter' AS type,
				       word_similarity($1, pr.name) AS score
				FROM promoters pr
				WHERE pr.name IS NOT NULL
				  AND word_similarity($1, pr.name) > 0.15
				ORDER BY score DESC
				LIMIT 8
			)
		) sub
		ORDER BY score DESC
		LIMIT $2`,
		q, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []core.SuggestionItem
	for rows.Next() {
		var item core.SuggestionItem
		if err := rows.Scan(&item.Text, &item.Type, &item.Score); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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

// ListCrawlRuns returns recent crawl runs, most recent first.
func (r *ProjectRepo) ListCrawlRuns(ctx context.Context, limit int) ([]core.CrawlRunItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, started_at, finished_at, status, start_id, end_id, processed, failed, error
		FROM crawl_runs
		ORDER BY started_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []core.CrawlRunItem
	for rows.Next() {
		var r core.CrawlRunItem
		if err := rows.Scan(&r.ID, &r.StartedAt, &r.FinishedAt, &r.Status,
			&r.StartID, &r.EndID, &r.Processed, &r.Failed, &r.Error); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// StatusDistribution returns project counts grouped by project_status.
func (r *ProjectRepo) StatusDistribution(ctx context.Context) ([]core.StatusDistributionItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT COALESCE(project_status, 'Unknown') AS status, COUNT(*) AS count
		FROM projects
		GROUP BY project_status
		ORDER BY count DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []core.StatusDistributionItem
	for rows.Next() {
		var item core.StatusDistributionItem
		if err := rows.Scan(&item.Status, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// TopBuilders returns the top promoters ranked by project count.
func (r *ProjectRepo) TopBuilders(ctx context.Context, limit int) ([]core.TopBuilderItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			COALESCE(pr.name, 'Unknown') AS promoter_name,
			COUNT(p.id)                  AS project_count,
			COALESCE(SUM(p.total_units), 0) AS total_units
		FROM projects p
		LEFT JOIN promoters pr ON p.promoter_id = pr.id
		GROUP BY pr.name
		ORDER BY project_count DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []core.TopBuilderItem
	for rows.Next() {
		var item core.TopBuilderItem
		if err := rows.Scan(&item.PromoterName, &item.ProjectCount, &item.TotalUnits); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ByDistrict returns project counts grouped by district.
func (r *ProjectRepo) ByDistrict(ctx context.Context, limit int) ([]core.DistrictCountItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT COALESCE(a.district, 'Unknown') AS district, COUNT(p.id) AS count
		FROM projects p
		LEFT JOIN addresses a ON a.entity_type = 'project' AND a.entity_id = p.id
		WHERE a.district IS NOT NULL AND a.district != ''
		GROUP BY a.district
		ORDER BY count DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []core.DistrictCountItem
	for rows.Next() {
		var item core.DistrictCountItem
		if err := rows.Scan(&item.District, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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
