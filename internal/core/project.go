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
