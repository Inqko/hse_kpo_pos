package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidCriterion = errors.New("некорректный критерий")

type Criterion struct {
	id       string
	name     string
	maxScore Score
	weight   Weight
}

func NewCriterion(id, name string, maxScore Score, weight Weight) (*Criterion, error) {
	trimmedID := strings.TrimSpace(id)
	trimmedName := strings.TrimSpace(name)
	if trimmedID == "" {
		return nil, fmt.Errorf("%w: идентификатор не может быть пустым", ErrInvalidCriterion)
	}
	if trimmedName == "" {
		return nil, fmt.Errorf("%w: название не может быть пустым", ErrInvalidCriterion)
	}
	if maxScore.Value() <= 0 {
		return nil, fmt.Errorf("%w: максимальный балл должен быть больше нуля", ErrInvalidCriterion)
	}
	if weight.Value() <= 0 || weight.Value() > 1 {
		return nil, fmt.Errorf("%w: вес должен быть в интервале (0, 1]", ErrInvalidCriterion)
	}
	return &Criterion{
		id:       trimmedID,
		name:     trimmedName,
		maxScore: maxScore,
		weight:   weight,
	}, nil
}

func (c *Criterion) ID() string {
	return c.id
}

func (c *Criterion) Name() string {
	return c.name
}

func (c *Criterion) MaxScore() Score {
	return c.maxScore
}

func (c *Criterion) Weight() Weight {
	return c.weight
}
