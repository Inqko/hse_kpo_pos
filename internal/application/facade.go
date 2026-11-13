package application

import (
	"errors"
	"fmt"
	"strings"

	"pos/internal/domain"
)

var (
	ErrStudentNotFound         = errors.New("студент не найден")
	ErrStudentAlreadyExists    = errors.New("студент уже существует")
	ErrCriterionNotFound       = errors.New("критерий не найден")
	ErrCriterionAlreadyExists  = errors.New("критерий уже существует")
	ErrEvaluationNotFound      = errors.New("оценка не найдена")
	ErrEvaluationAlreadyExists = errors.New("оценка по этому критерию уже выставлена")
	ErrWeightSumExceeded       = errors.New("суммарный вес критериев превышает 1")
	ErrUnsupportedExportFormat = errors.New("неподдерживаемый формат экспорта")
)

const weightSumTolerance = 1e-9

type POSFacade struct {
	studentRepo   StudentRepository
	criterionRepo CriterionRepository
	evalRepo      EvaluationRepository
	factory       *domain.EvaluationFactory
	calculator    *domain.GradeCalculator
	exporters     map[string]ReportExporter
	exportFormats []string
}

func NewPOSFacade(
	studentRepo StudentRepository,
	criterionRepo CriterionRepository,
	evalRepo EvaluationRepository,
	factory *domain.EvaluationFactory,
	calculator *domain.GradeCalculator,
	exporters []ReportExporter,
) *POSFacade {
	exportersByName := make(map[string]ReportExporter, len(exporters))
	formats := make([]string, 0, len(exporters))
	for _, exporter := range exporters {
		if exporter == nil {
			continue
		}
		name := strings.TrimSpace(exporter.Format())
		if name == "" {
			continue
		}
		if _, exists := exportersByName[name]; !exists {
			formats = append(formats, name)
		}
		exportersByName[name] = exporter
	}
	return &POSFacade{
		studentRepo:   studentRepo,
		criterionRepo: criterionRepo,
		evalRepo:      evalRepo,
		factory:       factory,
		calculator:    calculator,
		exporters:     exportersByName,
		exportFormats: formats,
	}
}

func (f *POSFacade) RegisterStudent(id, name string) (*domain.Student, error) {
	trimmedID := strings.TrimSpace(id)
	existing, err := f.studentRepo.GetByID(trimmedID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%w: %q", ErrStudentAlreadyExists, trimmedID)
	}
	student, err := domain.NewStudent(id, name)
	if err != nil {
		return nil, err
	}
	if err := f.studentRepo.Save(student); err != nil {
		return nil, fmt.Errorf("сохранить студента %q: %w", student.ID(), err)
	}
	return student, nil
}

func (f *POSFacade) ListStudents() ([]*domain.Student, error) {
	students, err := f.studentRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("получить список студентов: %w", err)
	}
	return students, nil
}

func (f *POSFacade) AddCriterion(id, name string, maxScoreVal, weightVal float64) (*domain.Criterion, error) {
	maxScore, err := domain.NewScore(maxScoreVal)
	if err != nil {
		return nil, err
	}
	weight, err := domain.NewWeight(weightVal)
	if err != nil {
		return nil, err
	}
	trimmedID := strings.TrimSpace(id)
	existing, err := f.criterionRepo.GetByID(trimmedID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%w: %q", ErrCriterionAlreadyExists, trimmedID)
	}
	criteria, err := f.criterionRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("получить список критериев: %w", err)
	}
	currentWeight := 0.0
	for _, criterion := range criteria {
		currentWeight += criterion.Weight().Value()
	}
	if currentWeight+weight.Value() > 1+weightSumTolerance {
		return nil, fmt.Errorf(
			"%w: текущая сумма %.2f, новый вес %.2f",
			ErrWeightSumExceeded,
			currentWeight,
			weight.Value(),
		)
	}
	criterion, err := domain.NewCriterion(id, name, maxScore, weight)
	if err != nil {
		return nil, err
	}
	if err := f.criterionRepo.Save(criterion); err != nil {
		return nil, fmt.Errorf("сохранить критерий %q: %w", criterion.ID(), err)
	}
	return criterion, nil
}

func (f *POSFacade) ListCriteria() ([]*domain.Criterion, error) {
	criteria, err := f.criterionRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("получить список критериев: %w", err)
	}
	return criteria, nil
}

func (f *POSFacade) AssignEvaluation(id, studentID, criterionID string, scoreVal float64) (*domain.Evaluation, error) {
	student, err := f.studentRepo.GetByID(strings.TrimSpace(studentID))
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %w", ErrStudentNotFound, studentID, err)
	}
	criterion, err := f.criterionRepo.GetByID(strings.TrimSpace(criterionID))
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %w", ErrCriterionNotFound, criterionID, err)
	}
	existing, err := f.evalRepo.GetByStudentAndCriterion(student.ID(), criterion.ID())
	if err == nil && existing != nil {
		return nil, fmt.Errorf(
			"%w: студент %q, критерий %q",
			ErrEvaluationAlreadyExists,
			student.ID(),
			criterion.ID(),
		)
	}
	score, err := domain.NewScore(scoreVal)
	if err != nil {
		return nil, err
	}
	evaluation, err := f.factory.CreateEvaluation(id, student, criterion, score)
	if err != nil {
		return nil, err
	}
	if err := f.evalRepo.Save(evaluation); err != nil {
		return nil, fmt.Errorf("сохранить оценку %q: %w", evaluation.ID(), err)
	}
	return evaluation, nil
}

func (f *POSFacade) UpdateEvaluation(studentID, criterionID string, newScoreVal float64) (*domain.Evaluation, error) {
	criterion, err := f.criterionRepo.GetByID(strings.TrimSpace(criterionID))
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %w", ErrCriterionNotFound, criterionID, err)
	}
	evaluation, err := f.evalRepo.GetByStudentAndCriterion(strings.TrimSpace(studentID), criterion.ID())
	if err != nil || evaluation == nil {
		return nil, fmt.Errorf("%w: студент %q, критерий %q", ErrEvaluationNotFound, studentID, criterion.ID())
	}
	score, err := domain.NewScore(newScoreVal)
	if err != nil {
		return nil, err
	}
	if err := evaluation.UpdateScore(score, criterion); err != nil {
		return nil, err
	}
	if err := f.evalRepo.Save(evaluation); err != nil {
		return nil, fmt.Errorf("сохранить оценку %q: %w", evaluation.ID(), err)
	}
	return evaluation, nil
}

func (f *POSFacade) CancelEvaluation(studentID, criterionID string) error {
	evaluation, err := f.evalRepo.GetByStudentAndCriterion(strings.TrimSpace(studentID), strings.TrimSpace(criterionID))
	if err != nil || evaluation == nil {
		return fmt.Errorf("%w: студент %q, критерий %q", ErrEvaluationNotFound, studentID, criterionID)
	}
	if err := f.evalRepo.Delete(evaluation.ID()); err != nil {
		return fmt.Errorf("удалить оценку %q: %w", evaluation.ID(), err)
	}
	return nil
}

func (f *POSFacade) ListStudentEvaluations(studentID string) ([]*domain.Evaluation, error) {
	evaluations, err := f.evalRepo.GetByStudentID(strings.TrimSpace(studentID))
	if err != nil {
		return nil, fmt.Errorf("получить оценки студента %q: %w", studentID, err)
	}
	return evaluations, nil
}

func (f *POSFacade) ListAllEvaluations() ([]*domain.Evaluation, error) {
	evaluations, err := f.evalRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("получить список оценок: %w", err)
	}
	return evaluations, nil
}

func (f *POSFacade) CalculateStudentTotal(studentID string) (domain.Score, error) {
	evaluations, err := f.evalRepo.GetByStudentID(strings.TrimSpace(studentID))
	if err != nil {
		return domain.Score{}, fmt.Errorf("получить оценки студента %q: %w", studentID, err)
	}
	criteria, err := f.criterionRepo.GetAll()
	if err != nil {
		return domain.Score{}, fmt.Errorf("получить список критериев: %w", err)
	}
	total, err := f.calculator.CalculateFinalGrade(evaluations, criteria)
	if err != nil {
		return domain.Score{}, fmt.Errorf("рассчитать итог студента %q: %w", studentID, err)
	}
	return total, nil
}

func (f *POSFacade) SwitchCalculationStrategy(strategyName string) error {
	if err := f.calculator.SetStrategy(strategyName); err != nil {
		return fmt.Errorf("переключить стратегию расчёта: %w", err)
	}
	return nil
}

func (f *POSFacade) GetAvailableStrategies() []string {
	return f.calculator.AvailableStrategies()
}

func (f *POSFacade) GetAvailableExportFormats() []string {
	formats := make([]string, len(f.exportFormats))
	copy(formats, f.exportFormats)
	return formats
}

func (f *POSFacade) ExportSummaryReport(format, filePath string) error {
	exporter, ok := f.exporters[strings.TrimSpace(format)]
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnsupportedExportFormat, format)
	}
	students, err := f.studentRepo.GetAll()
	if err != nil {
		return fmt.Errorf("получить список студентов: %w", err)
	}
	criteria, err := f.criterionRepo.GetAll()
	if err != nil {
		return fmt.Errorf("получить список критериев: %w", err)
	}
	strategyName := f.calculator.CurrentStrategy()
	rows := make([]StudentReportRow, 0, len(students))
	for _, student := range students {
		evaluations, err := f.evalRepo.GetByStudentID(student.ID())
		if err != nil {
			return fmt.Errorf("получить оценки студента %q: %w", student.ID(), err)
		}
		total, err := f.calculator.CalculateFinalGrade(evaluations, criteria)
		if err != nil {
			return fmt.Errorf("рассчитать итог студента %q: %w", student.ID(), err)
		}
		rows = append(rows, StudentReportRow{
			StudentID:   student.ID(),
			StudentName: student.Name(),
			FinalGrade:  total.Value(),
			Strategy:    strategyName,
		})
	}
	if err := exporter.Export(rows, filePath); err != nil {
		return fmt.Errorf("экспортировать отчёт в формате %q: %w", exporter.Format(), err)
	}
	return nil
}
