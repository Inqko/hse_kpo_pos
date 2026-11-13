package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidEvaluation = errors.New("некорректная оценка")
	ErrCriterionMismatch = errors.New("оценка относится к другому критерию")
	ErrScoreExceedsMax   = errors.New("балл превышает максимум критерия")
)

type Evaluation struct {
	id          string
	studentID   string
	criterionID string
	score       Score
}

func newEvaluation(id, studentID, criterionID string, score Score) (*Evaluation, error) {
	trimmedID := strings.TrimSpace(id)
	trimmedStudentID := strings.TrimSpace(studentID)
	trimmedCriterionID := strings.TrimSpace(criterionID)
	if trimmedID == "" {
		return nil, fmt.Errorf("%w: идентификатор не может быть пустым", ErrInvalidEvaluation)
	}
	if trimmedStudentID == "" {
		return nil, fmt.Errorf("%w: студент не задан", ErrInvalidEvaluation)
	}
	if trimmedCriterionID == "" {
		return nil, fmt.Errorf("%w: критерий не задан", ErrInvalidEvaluation)
	}
	return &Evaluation{
		id:          trimmedID,
		studentID:   trimmedStudentID,
		criterionID: trimmedCriterionID,
		score:       score,
	}, nil
}

func (e *Evaluation) ID() string {
	return e.id
}

func (e *Evaluation) StudentID() string {
	return e.studentID
}

func (e *Evaluation) CriterionID() string {
	return e.criterionID
}

func (e *Evaluation) Score() Score {
	return e.score
}

func (e *Evaluation) UpdateScore(newScore Score, criterion *Criterion) error {
	if criterion == nil {
		return fmt.Errorf("%w: критерий не задан", ErrInvalidEvaluation)
	}
	if criterion.ID() != e.criterionID {
		return fmt.Errorf("%w: ожидался %q, получен %q", ErrCriterionMismatch, e.criterionID, criterion.ID())
	}
	if !newScore.LessOrEqual(criterion.MaxScore()) {
		return fmt.Errorf(
			"%w: балл %.2f превышает максимум %.2f у критерия %q",
			ErrScoreExceedsMax,
			newScore.Value(),
			criterion.MaxScore().Value(),
			criterion.ID(),
		)
	}
	e.score = newScore
	return nil
}
