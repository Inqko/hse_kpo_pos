package domain

import (
	"errors"
	"fmt"
	"math"
)

var ErrInvalidWeight = errors.New("некорректный вес")

type Weight struct {
	value float64
}

func NewWeight(val float64) (Weight, error) {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return Weight{}, fmt.Errorf("%w: значение должно быть конечным числом", ErrInvalidWeight)
	}
	if val <= 0 || val > 1 {
		return Weight{}, fmt.Errorf("%w: значение %.2f должно быть в интервале (0, 1]", ErrInvalidWeight, val)
	}
	return Weight{value: val}, nil
}

func (w Weight) Value() float64 {
	return w.value
}
