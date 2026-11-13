package domain_test

import (
	"errors"
	"math"
	"testing"

	"pos/internal/domain"
)

func TestScoreRejectsNegativeAndNonFiniteValues(t *testing.T) {
	for _, value := range []float64{-10, -0.01} {
		if _, err := domain.NewScore(value); !errors.Is(err, domain.ErrInvalidScore) {
			t.Fatalf("NewScore(%v) error = %v, want ErrInvalidScore", value, err)
		}
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := domain.NewScore(value); !errors.Is(err, domain.ErrInvalidScore) {
			t.Fatalf("NewScore(%v) error = %v, want ErrInvalidScore", value, err)
		}
	}
	for _, value := range []float64{0, 0.01, 5, 10, 100} {
		score, err := domain.NewScore(value)
		if err != nil {
			t.Fatalf("NewScore(%v) error = %v", value, err)
		}
		if score.Value() != value {
			t.Fatalf("Score.Value() = %v, want %v", score.Value(), value)
		}
	}
}

func TestWeightRejectsValuesOutsideRange(t *testing.T) {
	for _, value := range []float64{-1, 0, 1.01, 2} {
		if _, err := domain.NewWeight(value); !errors.Is(err, domain.ErrInvalidWeight) {
			t.Fatalf("NewWeight(%v) error = %v, want ErrInvalidWeight", value, err)
		}
	}
	for _, value := range []float64{0.01, 0.3, 0.5, 1} {
		weight, err := domain.NewWeight(value)
		if err != nil {
			t.Fatalf("NewWeight(%v) error = %v", value, err)
		}
		if weight.Value() != value {
			t.Fatalf("Weight.Value() = %v, want %v", weight.Value(), value)
		}
	}
}

func TestEvaluationFactoryRejectsScoreAboveCriterionMax(t *testing.T) {
	criterion := mustCriterion(t, "c1", "Домашняя работа", 10, 0.5)
	student := mustStudent(t, "s1", "Иван")
	factory := domain.NewEvaluationFactory()

	if _, err := factory.CreateEvaluation("e1", student, criterion, mustScore(t, 12)); !errors.Is(err, domain.ErrScoreExceedsMax) {
		t.Fatalf("CreateEvaluation(12 из 10) error = %v, want ErrScoreExceedsMax", err)
	}
	evaluation, err := factory.CreateEvaluation("e2", student, criterion, mustScore(t, 10))
	if err != nil {
		t.Fatalf("CreateEvaluation(10 из 10) error = %v", err)
	}
	if evaluation.Score().Value() != 10 {
		t.Fatalf("Score = %v, want 10", evaluation.Score().Value())
	}
	if evaluation.StudentID() != student.ID() || evaluation.CriterionID() != criterion.ID() {
		t.Fatalf("evaluation references student %q criterion %q", evaluation.StudentID(), evaluation.CriterionID())
	}
}

func TestEvaluationRejectsScoreAboveMaxOnUpdate(t *testing.T) {
	criterion := mustCriterion(t, "c1", "Домашняя работа", 10, 0.5)
	student := mustStudent(t, "s1", "Иван")
	evaluation, err := domain.NewEvaluationFactory().CreateEvaluation("e1", student, criterion, mustScore(t, 8))
	if err != nil {
		t.Fatalf("CreateEvaluation error = %v", err)
	}
	if err := evaluation.UpdateScore(mustScore(t, 11), criterion); !errors.Is(err, domain.ErrScoreExceedsMax) {
		t.Fatalf("UpdateScore(11 из 10) error = %v, want ErrScoreExceedsMax", err)
	}
	if evaluation.Score().Value() != 8 {
		t.Fatalf("score changed after rejected update: %v", evaluation.Score().Value())
	}
	if err := evaluation.UpdateScore(mustScore(t, 9), criterion); err != nil {
		t.Fatalf("UpdateScore(9) error = %v", err)
	}
	if evaluation.Score().Value() != 9 {
		t.Fatalf("Score = %v, want 9", evaluation.Score().Value())
	}
}

func TestGradeCalculatorSwitchesStrategy(t *testing.T) {
	first := mustCriterion(t, "c1", "Домашняя работа", 10, 0.5)
	second := mustCriterion(t, "c2", "Контрольная работа", 20, 0.5)
	student := mustStudent(t, "s1", "Иван")
	factory := domain.NewEvaluationFactory()

	firstEvaluation, err := factory.CreateEvaluation("e1", student, first, mustScore(t, 8))
	if err != nil {
		t.Fatalf("CreateEvaluation e1 error = %v", err)
	}
	secondEvaluation, err := factory.CreateEvaluation("e2", student, second, mustScore(t, 10))
	if err != nil {
		t.Fatalf("CreateEvaluation e2 error = %v", err)
	}

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

	evaluations := []*domain.Evaluation{firstEvaluation, secondEvaluation}
	criteria := []*domain.Criterion{first, second}

	if calculator.CurrentStrategy() != domain.WeightedTenPointStrategyName {
		t.Fatalf("CurrentStrategy() = %q", calculator.CurrentStrategy())
	}
	weighted, err := calculator.CalculateFinalGrade(evaluations, criteria)
	if err != nil {
		t.Fatalf("CalculateFinalGrade weighted error = %v", err)
	}
	if weighted.Value() != 6.5 {
		t.Fatalf("weighted total = %v, want 6.5", weighted.Value())
	}

	if err := calculator.SetStrategy(domain.RawSumStrategyName); err != nil {
		t.Fatalf("SetStrategy(RawSum) error = %v", err)
	}
	raw, err := calculator.CalculateFinalGrade(evaluations, criteria)
	if err != nil {
		t.Fatalf("CalculateFinalGrade raw error = %v", err)
	}
	if raw.Value() != 9 {
		t.Fatalf("raw total = %v, want 9", raw.Value())
	}
	if weighted.Value() == raw.Value() {
		t.Fatal("strategies produced the same total")
	}

	if err := calculator.SetStrategy("missing"); !errors.Is(err, domain.ErrUnknownStrategy) {
		t.Fatalf("SetStrategy(missing) error = %v, want ErrUnknownStrategy", err)
	}
	if calculator.CurrentStrategy() != domain.RawSumStrategyName {
		t.Fatalf("strategy changed after failed switch: %q", calculator.CurrentStrategy())
	}
}

func TestGradeCalculatorRejectsEvaluationWithUnknownCriterion(t *testing.T) {
	criterion := mustCriterion(t, "c1", "Домашняя работа", 10, 0.5)
	foreign := mustCriterion(t, "c2", "Экзамен", 10, 0.5)
	student := mustStudent(t, "s1", "Иван")
	evaluation, err := domain.NewEvaluationFactory().CreateEvaluation("e1", student, foreign, mustScore(t, 5))
	if err != nil {
		t.Fatalf("CreateEvaluation error = %v", err)
	}
	calculator, err := domain.NewGradeCalculator(
		[]domain.CalculationStrategy{domain.NewWeightedTenPointStrategy()},
		domain.WeightedTenPointStrategyName,
	)
	if err != nil {
		t.Fatalf("NewGradeCalculator error = %v", err)
	}
	if _, err := calculator.CalculateFinalGrade([]*domain.Evaluation{evaluation}, []*domain.Criterion{criterion}); !errors.Is(err, domain.ErrCriterionNotFound) {
		t.Fatalf("CalculateFinalGrade error = %v, want ErrCriterionNotFound", err)
	}
}

func mustScore(t *testing.T, value float64) domain.Score {
	t.Helper()
	score, err := domain.NewScore(value)
	if err != nil {
		t.Fatalf("NewScore(%v) error = %v", value, err)
	}
	return score
}

func mustWeight(t *testing.T, value float64) domain.Weight {
	t.Helper()
	weight, err := domain.NewWeight(value)
	if err != nil {
		t.Fatalf("NewWeight(%v) error = %v", value, err)
	}
	return weight
}

func mustStudent(t *testing.T, id, name string) *domain.Student {
	t.Helper()
	student, err := domain.NewStudent(id, name)
	if err != nil {
		t.Fatalf("NewStudent(%q, %q) error = %v", id, name, err)
	}
	return student
}

func mustCriterion(t *testing.T, id, name string, maxScore, weight float64) *domain.Criterion {
	t.Helper()
	criterion, err := domain.NewCriterion(id, name, mustScore(t, maxScore), mustWeight(t, weight))
	if err != nil {
		t.Fatalf("NewCriterion(%q, %q) error = %v", id, name, err)
	}
	return criterion
}
