package application_test

import (
	"errors"
	"testing"

	"pos/internal/application"
	"pos/internal/domain"
	"pos/internal/infrastructure/memory"
)

func TestPOSFacadeRejectsEvaluationForUnknownCriterion(t *testing.T) {
	facade := newTestFacade(t)

	if _, err := facade.RegisterStudent("s1", "Иван"); err != nil {
		t.Fatalf("RegisterStudent error = %v", err)
	}
	if _, err := facade.AssignEvaluation("e1", "s1", "missing", 5); !errors.Is(err, application.ErrCriterionNotFound) {
		t.Fatalf("AssignEvaluation(unknown criterion) error = %v, want ErrCriterionNotFound", err)
	}
	if _, err := facade.AddCriterion("c1", "Домашняя работа", 10, 0.5); err != nil {
		t.Fatalf("AddCriterion error = %v", err)
	}
	evaluation, err := facade.AssignEvaluation("e1", "s1", "c1", 5)
	if err != nil {
		t.Fatalf("AssignEvaluation(valid) error = %v", err)
	}
	if evaluation.StudentID() != "s1" || evaluation.CriterionID() != "c1" {
		t.Fatalf("unexpected evaluation references: %q, %q", evaluation.StudentID(), evaluation.CriterionID())
	}
}

func TestPOSFacadeRejectsWeightSumAboveOne(t *testing.T) {
	facade := newTestFacade(t)

	if _, err := facade.AddCriterion("c1", "Домашняя работа", 10, 0.6); err != nil {
		t.Fatalf("AddCriterion(c1) error = %v", err)
	}
	if _, err := facade.AddCriterion("c2", "Экзамен", 10, 0.5); !errors.Is(err, application.ErrWeightSumExceeded) {
		t.Fatalf("AddCriterion(0.6 + 0.5) error = %v, want ErrWeightSumExceeded", err)
	}
	if _, err := facade.AddCriterion("c2", "Экзамен", 10, 0.4); err != nil {
		t.Fatalf("AddCriterion(exact 1.0) error = %v", err)
	}
	criteria, err := facade.ListCriteria()
	if err != nil {
		t.Fatalf("ListCriteria error = %v", err)
	}
	if len(criteria) != 2 {
		t.Fatalf("criteria count = %d, want 2", len(criteria))
	}
}

func TestPOSFacadeRejectsUnknownStudentWhenAssigning(t *testing.T) {
	facade := newTestFacade(t)

	if _, err := facade.AddCriterion("c1", "Домашняя работа", 10, 0.5); err != nil {
		t.Fatalf("AddCriterion error = %v", err)
	}
	if _, err := facade.AssignEvaluation("e1", "missing", "c1", 5); !errors.Is(err, application.ErrStudentNotFound) {
		t.Fatalf("AssignEvaluation(unknown student) error = %v, want ErrStudentNotFound", err)
	}
}

func TestPOSFacadeRejectsDuplicateEvaluation(t *testing.T) {
	facade := newTestFacade(t)

	if _, err := facade.RegisterStudent("s1", "Иван"); err != nil {
		t.Fatalf("RegisterStudent error = %v", err)
	}
	if _, err := facade.AddCriterion("c1", "Домашняя работа", 10, 0.5); err != nil {
		t.Fatalf("AddCriterion error = %v", err)
	}
	if _, err := facade.AssignEvaluation("e1", "s1", "c1", 5); err != nil {
		t.Fatalf("AssignEvaluation error = %v", err)
	}
	if _, err := facade.AssignEvaluation("e2", "s1", "c1", 6); !errors.Is(err, application.ErrEvaluationAlreadyExists) {
		t.Fatalf("AssignEvaluation(duplicate) error = %v, want ErrEvaluationAlreadyExists", err)
	}
}

func TestPOSFacadeUpdatesAndCancelsEvaluation(t *testing.T) {
	facade := newTestFacade(t)

	if _, err := facade.RegisterStudent("s1", "Иван"); err != nil {
		t.Fatalf("RegisterStudent error = %v", err)
	}
	if _, err := facade.AddCriterion("c1", "Домашняя работа", 10, 0.5); err != nil {
		t.Fatalf("AddCriterion error = %v", err)
	}
	if _, err := facade.AssignEvaluation("e1", "s1", "c1", 5); err != nil {
		t.Fatalf("AssignEvaluation error = %v", err)
	}
	if _, err := facade.UpdateEvaluation("s1", "c1", 11); !errors.Is(err, domain.ErrScoreExceedsMax) {
		t.Fatalf("UpdateEvaluation(11 из 10) error = %v, want ErrScoreExceedsMax", err)
	}
	updated, err := facade.UpdateEvaluation("s1", "c1", 9)
	if err != nil {
		t.Fatalf("UpdateEvaluation(9) error = %v", err)
	}
	if updated.Score().Value() != 9 {
		t.Fatalf("updated score = %v, want 9", updated.Score().Value())
	}
	if err := facade.CancelEvaluation("s1", "c1"); err != nil {
		t.Fatalf("CancelEvaluation error = %v", err)
	}
	evaluations, err := facade.ListStudentEvaluations("s1")
	if err != nil {
		t.Fatalf("ListStudentEvaluations error = %v", err)
	}
	if len(evaluations) != 0 {
		t.Fatalf("evaluations after cancel = %d, want 0", len(evaluations))
	}
}

func newTestFacade(t *testing.T) *application.POSFacade {
	t.Helper()
	calculator, err := domain.NewGradeCalculator(
		[]domain.CalculationStrategy{
			domain.NewWeightedTenPointStrategy(),
			domain.NewRawSumStrategy(),
		},
		domain.WeightedTenPointStrategyName,
	)
	if err != nil {
		t.Fatalf("NewGradeCalculator error = %v", err)
	}
	return application.NewPOSFacade(
		memory.NewInMemoryStudentRepository(),
		memory.NewInMemoryCriterionRepository(),
		memory.NewInMemoryEvaluationRepository(),
		domain.NewEvaluationFactory(),
		calculator,
		nil,
	)
}
