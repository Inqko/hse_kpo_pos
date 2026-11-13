package memory

import (
	"errors"
	"fmt"
	"strings"

	"pos/internal/application"
	"pos/internal/domain"
)

var ErrEvaluationNotFound = errors.New("оценка не найдена в хранилище")

type InMemoryEvaluationRepository struct {
	evaluations map[string]*domain.Evaluation
	order       []string
}

var _ application.EvaluationRepository = (*InMemoryEvaluationRepository)(nil)

func NewInMemoryEvaluationRepository() *InMemoryEvaluationRepository {
	return &InMemoryEvaluationRepository{
		evaluations: make(map[string]*domain.Evaluation),
	}
}

func (r *InMemoryEvaluationRepository) Save(evaluation *domain.Evaluation) error {
	if evaluation == nil {
		return errors.New("оценка не задана")
	}
	if _, exists := r.evaluations[evaluation.ID()]; !exists {
		r.order = append(r.order, evaluation.ID())
	}
	r.evaluations[evaluation.ID()] = evaluation
	return nil
}

func (r *InMemoryEvaluationRepository) GetByID(id string) (*domain.Evaluation, error) {
	trimmedID := strings.TrimSpace(id)
	evaluation, exists := r.evaluations[trimmedID]
	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrEvaluationNotFound, trimmedID)
	}
	return evaluation, nil
}

func (r *InMemoryEvaluationRepository) GetByStudentAndCriterion(studentID, criterionID string) (*domain.Evaluation, error) {
	trimmedStudentID := strings.TrimSpace(studentID)
	trimmedCriterionID := strings.TrimSpace(criterionID)
	for _, id := range r.order {
		evaluation, exists := r.evaluations[id]
		if !exists {
			continue
		}
		if evaluation.StudentID() == trimmedStudentID && evaluation.CriterionID() == trimmedCriterionID {
			return evaluation, nil
		}
	}
	return nil, fmt.Errorf(
		"%w: студент %q, критерий %q",
		ErrEvaluationNotFound,
		trimmedStudentID,
		trimmedCriterionID,
	)
}

func (r *InMemoryEvaluationRepository) GetByStudentID(studentID string) ([]*domain.Evaluation, error) {
	trimmedStudentID := strings.TrimSpace(studentID)
	evaluations := make([]*domain.Evaluation, 0)
	for _, id := range r.order {
		evaluation, exists := r.evaluations[id]
		if !exists {
			continue
		}
		if evaluation.StudentID() == trimmedStudentID {
			evaluations = append(evaluations, evaluation)
		}
	}
	return evaluations, nil
}

func (r *InMemoryEvaluationRepository) GetAll() ([]*domain.Evaluation, error) {
	evaluations := make([]*domain.Evaluation, 0, len(r.order))
	for _, id := range r.order {
		evaluation, exists := r.evaluations[id]
		if exists {
			evaluations = append(evaluations, evaluation)
		}
	}
	return evaluations, nil
}

func (r *InMemoryEvaluationRepository) Delete(id string) error {
	trimmedID := strings.TrimSpace(id)
	if _, exists := r.evaluations[trimmedID]; !exists {
		return fmt.Errorf("%w: %q", ErrEvaluationNotFound, trimmedID)
	}
	delete(r.evaluations, trimmedID)
	for index, existingID := range r.order {
		if existingID == trimmedID {
			r.order = append(r.order[:index], r.order[index+1:]...)
			break
		}
	}
	return nil
}
