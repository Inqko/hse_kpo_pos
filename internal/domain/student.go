package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidStudent = errors.New("некорректный студент")

type Student struct {
	id   string
	name string
}

func NewStudent(id, name string) (*Student, error) {
	trimmedID := strings.TrimSpace(id)
	trimmedName := strings.TrimSpace(name)
	if trimmedID == "" {
		return nil, fmt.Errorf("%w: идентификатор не может быть пустым", ErrInvalidStudent)
	}
	if trimmedName == "" {
		return nil, fmt.Errorf("%w: имя не может быть пустым", ErrInvalidStudent)
	}
	return &Student{id: trimmedID, name: trimmedName}, nil
}

func (s *Student) ID() string {
	return s.id
}

func (s *Student) Name() string {
	return s.name
}
