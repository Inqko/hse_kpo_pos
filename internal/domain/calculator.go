package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrInvalidStrategy   = errors.New("некорректная стратегия расчёта")
	ErrUnknownStrategy   = errors.New("неизвестная стратегия расчёта")
	ErrCriterionNotFound = errors.New("критерий отсутствует в рубрикаторе")
)

const (
	WeightedTenPointStrategyName = "weighted-ten-point"
	RawSumStrategyName           = "raw-sum"
)

type CalculationStrategy interface {
	Name() string
	Calculate(evaluations []*Evaluation, criteria map[string]*Criterion) (Score, error)
}

type WeightedTenPointStrategy struct{}

func NewWeightedTenPointStrategy() *WeightedTenPointStrategy {
	return &WeightedTenPointStrategy{}
}

func (s *WeightedTenPointStrategy) Name() string {
	return WeightedTenPointStrategyName
}

func (s *WeightedTenPointStrategy) Calculate(evaluations []*Evaluation, criteria map[string]*Criterion) (Score, error) {
	if err := validateEvaluations(evaluations, criteria); err != nil {
		return Score{}, err
	}
	total := 0.0
	for _, evaluation := range evaluations {
		criterion := criteria[evaluation.CriterionID()]
		total += evaluation.Score().Value() / criterion.MaxScore().Value() * criterion.Weight().Value() * 10
	}
	return NewScore(roundToTwo(total))
}

type RawSumStrategy struct{}

func NewRawSumStrategy() *RawSumStrategy {
	return &RawSumStrategy{}
}

func (s *RawSumStrategy) Name() string {
	return RawSumStrategyName
}

func (s *RawSumStrategy) Calculate(evaluations []*Evaluation, criteria map[string]*Criterion) (Score, error) {
	if err := validateEvaluations(evaluations, criteria); err != nil {
		return Score{}, err
	}
	total := 0.0
	for _, evaluation := range evaluations {
		criterion := criteria[evaluation.CriterionID()]
		total += evaluation.Score().Value() * criterion.Weight().Value()
	}
	return NewScore(roundToTwo(total))
}

type GradeCalculator struct {
	strategies map[string]CalculationStrategy
	order      []string
	current    CalculationStrategy
}

func NewGradeCalculator(strategies []CalculationStrategy, defaultStrategyName string) (*GradeCalculator, error) {
	if len(strategies) == 0 {
		return nil, fmt.Errorf("%w: список стратегий пуст", ErrInvalidStrategy)
	}
	available := make(map[string]CalculationStrategy, len(strategies))
	order := make([]string, 0, len(strategies))
	for _, strategy := range strategies {
		if strategy == nil {
			return nil, fmt.Errorf("%w: стратегия не задана", ErrInvalidStrategy)
		}
		name := strings.TrimSpace(strategy.Name())
		if name == "" {
			return nil, fmt.Errorf("%w: имя стратегии не может быть пустым", ErrInvalidStrategy)
		}
		if _, exists := available[name]; exists {
			return nil, fmt.Errorf("%w: стратегия %q указана дважды", ErrInvalidStrategy, name)
		}
		available[name] = strategy
		order = append(order, name)
	}
	current, ok := available[strings.TrimSpace(defaultStrategyName)]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownStrategy, defaultStrategyName)
	}
	return &GradeCalculator{strategies: available, order: order, current: current}, nil
}

func (g *GradeCalculator) SetStrategy(name string) error {
	strategy, ok := g.strategies[strings.TrimSpace(name)]
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownStrategy, name)
	}
	g.current = strategy
	return nil
}

func (g *GradeCalculator) AvailableStrategies() []string {
	names := make([]string, len(g.order))
	copy(names, g.order)
	return names
}

func (g *GradeCalculator) CurrentStrategy() string {
	return g.current.Name()
}

func (g *GradeCalculator) CalculateFinalGrade(evaluations []*Evaluation, criteria []*Criterion) (Score, error) {
	rubric := make(map[string]*Criterion, len(criteria))
	for _, criterion := range criteria {
		if criterion == nil {
			return Score{}, fmt.Errorf("%w: критерий не задан", ErrInvalidCriterion)
		}
		if _, exists := rubric[criterion.ID()]; exists {
			return Score{}, fmt.Errorf("%w: критерий %q указан дважды", ErrInvalidCriterion, criterion.ID())
		}
		rubric[criterion.ID()] = criterion
	}
	return g.current.Calculate(evaluations, rubric)
}

func validateEvaluations(evaluations []*Evaluation, criteria map[string]*Criterion) error {
	for index, evaluation := range evaluations {
		if evaluation == nil {
			return fmt.Errorf("%w: оценка #%d не задана", ErrInvalidEvaluation, index+1)
		}
		criterion, ok := criteria[evaluation.CriterionID()]
		if !ok || criterion == nil {
			return fmt.Errorf("%w: %q", ErrCriterionNotFound, evaluation.CriterionID())
		}
		if !evaluation.Score().LessOrEqual(criterion.MaxScore()) {
			return fmt.Errorf(
				"%w: балл %.2f превышает максимум %.2f у критерия %q",
				ErrScoreExceedsMax,
				evaluation.Score().Value(),
				criterion.MaxScore().Value(),
				criterion.ID(),
			)
		}
	}
	return nil
}

func roundToTwo(value float64) float64 {
	return math.Round(value*100) / 100
}
