package domain

import (
	"fmt"
	"strings"
)

type EvaluationFactory struct{}

func NewEvaluationFactory() *EvaluationFactory {
	return &EvaluationFactory{}
}

func (f *EvaluationFactory) CreateEvaluation(id string, student *Student, criterion *Criterion, score Score) (*Evaluation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: идентификатор не может быть пустым", ErrInvalidEvaluation)
	}
	if student == nil {
		return nil, fmt.Errorf("%w: студент не задан", ErrInvalidEvaluation)
	}
	if criterion == nil {
		return nil, fmt.Errorf("%w: критерий не задан", ErrInvalidEvaluation)
	}
	if !score.LessOrEqual(criterion.MaxScore()) {
		return nil, fmt.Errorf(
			"%w: балл %.2f превышает максимум %.2f у критерия %q",
			ErrScoreExceedsMax,
			score.Value(),
			criterion.MaxScore().Value(),
			criterion.ID(),
		)
	}
	return newEvaluation(id, student.ID(), criterion.ID(), score)
}
