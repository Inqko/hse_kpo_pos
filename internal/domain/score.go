package domain

import (
	"errors"
	"fmt"
	"math"
)

var ErrInvalidScore = errors.New("некорректный балл")

type Score struct {
	value float64
}

func NewScore(val float64) (Score, error) {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return Score{}, fmt.Errorf("%w: значение должно быть конечным числом", ErrInvalidScore)
	}
	if val < 0 {
		return Score{}, fmt.Errorf("%w: значение %.2f не может быть отрицательным", ErrInvalidScore, val)
	}
	return Score{value: val}, nil
}

func (s Score) Value() float64 {
	return s.value
}

func (s Score) LessOrEqual(other Score) bool {
	return s.value <= other.value
}
