package application

import "pos/internal/domain"

type StudentRepository interface {
	Save(student *domain.Student) error
	GetByID(id string) (*domain.Student, error)
	GetAll() ([]*domain.Student, error)
}

type CriterionRepository interface {
	Save(criterion *domain.Criterion) error
	GetByID(id string) (*domain.Criterion, error)
	GetAll() ([]*domain.Criterion, error)
}

type EvaluationRepository interface {
	Save(evaluation *domain.Evaluation) error
	GetByID(id string) (*domain.Evaluation, error)
	GetByStudentAndCriterion(studentID, criterionID string) (*domain.Evaluation, error)
	GetByStudentID(studentID string) ([]*domain.Evaluation, error)
	GetAll() ([]*domain.Evaluation, error)
	Delete(id string) error
}

type StudentReportRow struct {
	StudentID   string
	StudentName string
	FinalGrade  float64
	Strategy    string
}

type ReportExporter interface {
	Format() string
	Export(rows []StudentReportRow, destination string) error
}
