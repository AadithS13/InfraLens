package core

import (
	"context"
	"fmt"
)

// ProjectRepo defines what the core layer needs from the data layer.
// The repo package implements this.
type ProjectRepo interface {
	Search(ctx context.Context, f SearchFilter) ([]ProjectListItem, int, error)
	GetByID(ctx context.Context, id int) (*ProjectDetail, error)
	GetChanges(ctx context.Context, projectID int) ([]ChangeItem, error)
	ListCrawlRuns(ctx context.Context, limit int) ([]CrawlRunItem, error)
	StatusDistribution(ctx context.Context) ([]StatusDistributionItem, error)
	TopBuilders(ctx context.Context, limit int) ([]TopBuilderItem, error)
	ByDistrict(ctx context.Context, limit int) ([]DistrictCountItem, error)
	Suggestions(ctx context.Context, q string, limit int) ([]SuggestionItem, error)
}

type ProjectService struct {
	repo ProjectRepo
}

func NewProjectService(r ProjectRepo) *ProjectService {
	return &ProjectService{repo: r}
}

func (s *ProjectService) Search(ctx context.Context, f SearchFilter) (*ListResponse[ProjectListItem], error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}

	items, total, err := s.repo.Search(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("search projects: %w", err)
	}
	if items == nil {
		items = []ProjectListItem{}
	}

	return &ListResponse[ProjectListItem]{
		Data: items,
		Meta: Meta{Page: f.Page, Limit: f.Limit, Total: total},
	}, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id int) (*ProjectDetail, error) {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return project, nil
}

func (s *ProjectService) GetChanges(ctx context.Context, projectID int) ([]ChangeItem, error) {
	changes, err := s.repo.GetChanges(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get changes: %w", err)
	}
	if changes == nil {
		changes = []ChangeItem{}
	}
	return changes, nil
}

func (s *ProjectService) ListCrawlRuns(ctx context.Context, limit int) ([]CrawlRunItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	runs, err := s.repo.ListCrawlRuns(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list crawl runs: %w", err)
	}
	if runs == nil {
		runs = []CrawlRunItem{}
	}
	return runs, nil
}

func (s *ProjectService) StatusDistribution(ctx context.Context) ([]StatusDistributionItem, error) {
	items, err := s.repo.StatusDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("status distribution: %w", err)
	}
	if items == nil {
		items = []StatusDistributionItem{}
	}
	return items, nil
}

func (s *ProjectService) TopBuilders(ctx context.Context, limit int) ([]TopBuilderItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	items, err := s.repo.TopBuilders(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("top builders: %w", err)
	}
	if items == nil {
		items = []TopBuilderItem{}
	}
	return items, nil
}

func (s *ProjectService) ByDistrict(ctx context.Context, limit int) ([]DistrictCountItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	items, err := s.repo.ByDistrict(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("by district: %w", err)
	}
	if items == nil {
		items = []DistrictCountItem{}
	}
	return items, nil
}

// Suggestions returns autocomplete results for a query string.
// Matches project names and promoter names using pg_trgm word_similarity.
func (s *ProjectService) Suggestions(ctx context.Context, q string, limit int) ([]SuggestionItem, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	items, err := s.repo.Suggestions(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("suggestions: %w", err)
	}
	if items == nil {
		items = []SuggestionItem{}
	}
	return items, nil
}
