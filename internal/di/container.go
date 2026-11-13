package di

import (
	"go.uber.org/dig"

	"pos/internal/application"
	"pos/internal/domain"
	"pos/internal/infrastructure/cli"
	"pos/internal/infrastructure/export"
	"pos/internal/infrastructure/memory"
)

func BuildContainer() (*dig.Container, error) {
	container := dig.New()

	if err := container.Provide(memory.NewInMemoryStudentRepository, dig.As(new(application.StudentRepository))); err != nil {
		return nil, err
	}
	if err := container.Provide(memory.NewInMemoryCriterionRepository, dig.As(new(application.CriterionRepository))); err != nil {
		return nil, err
	}
	if err := container.Provide(memory.NewInMemoryEvaluationRepository, dig.As(new(application.EvaluationRepository))); err != nil {
		return nil, err
	}
	if err := container.Provide(domain.NewEvaluationFactory); err != nil {
		return nil, err
	}
	if err := container.Provide(func() []domain.CalculationStrategy {
		return []domain.CalculationStrategy{
			domain.NewWeightedTenPointStrategy(),
			domain.NewRawSumStrategy(),
		}
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(func() string {
		return domain.WeightedTenPointStrategyName
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(domain.NewGradeCalculator); err != nil {
		return nil, err
	}
	if err := container.Provide(func() []application.ReportExporter {
		return []application.ReportExporter{
			export.NewCSVReportExporter(),
			export.NewJSONReportExporter(),
		}
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(application.NewPOSFacade); err != nil {
		return nil, err
	}
	if err := container.Provide(cli.NewConsoleApp); err != nil {
		return nil, err
	}
	return container, nil
}
