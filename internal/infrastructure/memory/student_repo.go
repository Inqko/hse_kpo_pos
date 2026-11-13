package memory

import (
	"errors"
	"fmt"
	"strings"

	"pos/internal/application"
	"pos/internal/domain"
)

var ErrStudentNotFound = errors.New("студент не найден в хранилище")

type InMemoryStudentRepository struct {
	students map[string]*domain.Student
	order    []string
}

var _ application.StudentRepository = (*InMemoryStudentRepository)(nil)

func NewInMemoryStudentRepository() *InMemoryStudentRepository {
	return &InMemoryStudentRepository{
		students: make(map[string]*domain.Student),
	}
}

func (r *InMemoryStudentRepository) Save(student *domain.Student) error {
	if student == nil {
		return errors.New("студент не задан")
	}
	if _, exists := r.students[student.ID()]; !exists {
		r.order = append(r.order, student.ID())
	}
	r.students[student.ID()] = student
	return nil
}

func (r *InMemoryStudentRepository) GetByID(id string) (*domain.Student, error) {
	trimmedID := strings.TrimSpace(id)
	student, exists := r.students[trimmedID]
	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrStudentNotFound, trimmedID)
	}
	return student, nil
}

func (r *InMemoryStudentRepository) GetAll() ([]*domain.Student, error) {
	students := make([]*domain.Student, 0, len(r.order))
	for _, id := range r.order {
		student, exists := r.students[id]
		if exists {
			students = append(students, student)
		}
	}
	return students, nil
}
