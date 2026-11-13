package memory

import (
	"errors"
	"fmt"
	"strings"

	"pos/internal/application"
	"pos/internal/domain"
)

var ErrCriterionNotFound = errors.New("критерий не найден в хранилище")

type InMemoryCriterionRepository struct {
	criteria map[string]*domain.Criterion
	order    []string
}

var _ application.CriterionRepository = (*InMemoryCriterionRepository)(nil)

func NewInMemoryCriterionRepository() *InMemoryCriterionRepository {
	return &InMemoryCriterionRepository{
		criteria: make(map[string]*domain.Criterion),
	}
}

func (r *InMemoryCriterionRepository) Save(criterion *domain.Criterion) error {
	if criterion == nil {
		return errors.New("критерий не задан")
	}
	if _, exists := r.criteria[criterion.ID()]; !exists {
		r.order = append(r.order, criterion.ID())
	}
	r.criteria[criterion.ID()] = criterion
	return nil
}

func (r *InMemoryCriterionRepository) GetByID(id string) (*domain.Criterion, error) {
	trimmedID := strings.TrimSpace(id)
	criterion, exists := r.criteria[trimmedID]
	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrCriterionNotFound, trimmedID)
	}
	return criterion, nil
}

func (r *InMemoryCriterionRepository) GetAll() ([]*domain.Criterion, error) {
	criteria := make([]*domain.Criterion, 0, len(r.order))
	for _, id := range r.order {
		criterion, exists := r.criteria[id]
		if exists {
			criteria = append(criteria, criterion)
		}
	}
	return criteria, nil
}
