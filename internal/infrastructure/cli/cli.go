package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"pos/internal/application"
	"pos/internal/domain"
)

type ConsoleApp struct {
	facade  *application.POSFacade
	scanner *bufio.Scanner
	out     io.Writer
}

func NewConsoleApp(facade *application.POSFacade) *ConsoleApp {
	return &ConsoleApp{
		facade:  facade,
		scanner: bufio.NewScanner(os.Stdin),
		out:     os.Stdout,
	}
}

func (a *ConsoleApp) Run() error {
	for {
		a.printMenu()
		choice, ok := a.ask("Пункт меню: ")
		if !ok {
			fmt.Fprintln(a.out)
			return nil
		}
		switch choice {
		case "0":
			fmt.Fprintln(a.out, "До свидания!")
			return nil
		case "1":
			a.addStudent()
		case "2":
			a.listStudents()
		case "3":
			a.addCriterion()
		case "4":
			a.listCriteria()
		case "5":
			a.assignEvaluation()
		case "6":
			a.changeOrCancelEvaluation()
		case "7":
			a.showStudentEvaluations()
		case "8":
			a.switchStrategy()
		case "9":
			a.exportReport()
		case "10":
			a.runDemo()
		default:
			fmt.Fprintf(a.out, "Неизвестный пункт меню: %q\n", choice)
		}
	}
}

func (a *ConsoleApp) printMenu() {
	fmt.Fprintln(a.out)
	fmt.Fprintln(a.out, "=== ПОС: Преподавательская Оценка Студентов ===")
	fmt.Fprintln(a.out, "1. Добавить студента")
	fmt.Fprintln(a.out, "2. Список студентов")
	fmt.Fprintln(a.out, "3. Добавить критерий в рубрикатор")
	fmt.Fprintln(a.out, "4. Показать рубрикатор критериев")
	fmt.Fprintln(a.out, "5. Выставить оценку студенту")
	fmt.Fprintln(a.out, "6. Изменить или отменить оценку")
	fmt.Fprintln(a.out, "7. Показать оценки студента и рассчитать его итог")
	fmt.Fprintln(a.out, "8. Сменить стратегию расчёта итога (Weighted / RawSum)")
	fmt.Fprintln(a.out, "9. Экспортировать итоговую ведомость (CSV или JSON)")
	fmt.Fprintln(a.out, "10. Запустить автоматический сквозной демо-сценарий")
	fmt.Fprintln(a.out, "0. Выход")
}

func (a *ConsoleApp) ask(label string) (string, bool) {
	fmt.Fprint(a.out, label)
	if !a.scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(a.scanner.Text()), true
}

func (a *ConsoleApp) askFloat(label string) (float64, bool) {
	value, ok := a.ask(label)
	if !ok {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		fmt.Fprintf(a.out, "Ожидалось число, получено %q\n", value)
		return 0, false
	}
	return parsed, true
}

func (a *ConsoleApp) printError(err error) {
	fmt.Fprintf(a.out, "Ошибка: %v\n", err)
}

func (a *ConsoleApp) addStudent() {
	id, ok := a.ask("ID студента: ")
	if !ok {
		return
	}
	name, ok := a.ask("Имя студента: ")
	if !ok {
		return
	}
	student, err := a.facade.RegisterStudent(id, name)
	if err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintf(a.out, "Студент добавлен: %s (%s)\n", student.Name(), student.ID())
}

func (a *ConsoleApp) listStudents() {
	students, err := a.facade.ListStudents()
	if err != nil {
		a.printError(err)
		return
	}
	if len(students) == 0 {
		fmt.Fprintln(a.out, "Студентов пока нет")
		return
	}
	fmt.Fprintf(a.out, "%-12s %s\n", "ID", "Имя")
	for _, student := range students {
		fmt.Fprintf(a.out, "%-12s %s\n", student.ID(), student.Name())
	}
}

func (a *ConsoleApp) addCriterion() {
	id, ok := a.ask("ID критерия: ")
	if !ok {
		return
	}
	name, ok := a.ask("Название критерия: ")
	if !ok {
		return
	}
	maxScore, ok := a.askFloat("Максимальный балл: ")
	if !ok {
		return
	}
	weight, ok := a.askFloat("Вес критерия (0, 1]: ")
	if !ok {
		return
	}
	criterion, err := a.facade.AddCriterion(id, name, maxScore, weight)
	if err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintf(
		a.out,
		"Критерий добавлен: %s (%s), максимум %.2f, вес %.2f\n",
		criterion.Name(),
		criterion.ID(),
		criterion.MaxScore().Value(),
		criterion.Weight().Value(),
	)
}

func (a *ConsoleApp) listCriteria() {
	criteria, err := a.facade.ListCriteria()
	if err != nil {
		a.printError(err)
		return
	}
	if len(criteria) == 0 {
		fmt.Fprintln(a.out, "Рубрикатор пуст")
		return
	}
	totalWeight := 0.0
	fmt.Fprintf(a.out, "%-12s %-24s %8s %8s\n", "ID", "Название", "Максимум", "Вес")
	for _, criterion := range criteria {
		fmt.Fprintf(
			a.out,
			"%-12s %-24s %8.2f %8.2f\n",
			criterion.ID(),
			criterion.Name(),
			criterion.MaxScore().Value(),
			criterion.Weight().Value(),
		)
		totalWeight += criterion.Weight().Value()
	}
	fmt.Fprintf(a.out, "Суммарный вес: %.2f\n", totalWeight)
}

func (a *ConsoleApp) assignEvaluation() {
	id, ok := a.ask("ID оценки (Enter — сгенерировать): ")
	if !ok {
		return
	}
	if id == "" {
		id = a.nextEvaluationID()
	}
	studentID, ok := a.ask("ID студента: ")
	if !ok {
		return
	}
	criterionID, ok := a.ask("ID критерия: ")
	if !ok {
		return
	}
	score, ok := a.askFloat("Балл: ")
	if !ok {
		return
	}
	evaluation, err := a.facade.AssignEvaluation(id, studentID, criterionID, score)
	if err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintf(
		a.out,
		"Оценка выставлена: %s, студент %s, критерий %s, балл %.2f\n",
		evaluation.ID(),
		evaluation.StudentID(),
		evaluation.CriterionID(),
		evaluation.Score().Value(),
	)
}

func (a *ConsoleApp) changeOrCancelEvaluation() {
	action, ok := a.ask("1 — изменить, 2 — отменить, 0 — назад: ")
	if !ok || action == "0" {
		return
	}
	studentID, ok := a.ask("ID студента: ")
	if !ok {
		return
	}
	criterionID, ok := a.ask("ID критерия: ")
	if !ok {
		return
	}
	switch action {
	case "1":
		score, ok := a.askFloat("Новый балл: ")
		if !ok {
			return
		}
		evaluation, err := a.facade.UpdateEvaluation(studentID, criterionID, score)
		if err != nil {
			a.printError(err)
			return
		}
		fmt.Fprintf(a.out, "Оценка обновлена: %s, балл %.2f\n", evaluation.ID(), evaluation.Score().Value())
	case "2":
		if err := a.facade.CancelEvaluation(studentID, criterionID); err != nil {
			a.printError(err)
			return
		}
		fmt.Fprintf(a.out, "Оценка студента %s по критерию %s отменена\n", studentID, criterionID)
	default:
		fmt.Fprintf(a.out, "Неизвестное действие: %q\n", action)
	}
}

func (a *ConsoleApp) showStudentEvaluations() {
	studentID, ok := a.ask("ID студента: ")
	if !ok {
		return
	}
	student, err := a.findStudent(studentID)
	if err != nil {
		a.printError(err)
		return
	}
	evaluations, err := a.facade.ListStudentEvaluations(student.ID())
	if err != nil {
		a.printError(err)
		return
	}
	criteria, err := a.facade.ListCriteria()
	if err != nil {
		a.printError(err)
		return
	}
	criteriaByID := make(map[string]*domain.Criterion, len(criteria))
	for _, criterion := range criteria {
		criteriaByID[criterion.ID()] = criterion
	}
	fmt.Fprintf(a.out, "Оценки студента %s (%s):\n", student.Name(), student.ID())
	if len(evaluations) == 0 {
		fmt.Fprintln(a.out, "Оценок пока нет")
	}
	for _, evaluation := range evaluations {
		criterion, exists := criteriaByID[evaluation.CriterionID()]
		if !exists {
			fmt.Fprintf(a.out, "%-12s %8.2f\n", evaluation.CriterionID(), evaluation.Score().Value())
			continue
		}
		fmt.Fprintf(
			a.out,
			"%-12s %-24s %8.2f из %8.2f\n",
			criterion.ID(),
			criterion.Name(),
			evaluation.Score().Value(),
			criterion.MaxScore().Value(),
		)
	}
	total, err := a.facade.CalculateStudentTotal(student.ID())
	if err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintf(a.out, "Итог: %.2f\n", total.Value())
}

func (a *ConsoleApp) switchStrategy() {
	strategies := a.facade.GetAvailableStrategies()
	fmt.Fprintf(a.out, "Доступные стратегии: %s\n", strings.Join(strategies, ", "))
	name, ok := a.ask("Стратегия: ")
	if !ok {
		return
	}
	if err := a.facade.SwitchCalculationStrategy(name); err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintf(a.out, "Активная стратегия: %s\n", name)
}

func (a *ConsoleApp) exportReport() {
	formats := a.facade.GetAvailableExportFormats()
	if len(formats) == 0 {
		fmt.Fprintln(a.out, "Нет доступных форматов экспорта")
		return
	}
	fmt.Fprintf(a.out, "Доступные форматы: %s\n", strings.Join(formats, ", "))
	format, ok := a.ask("Формат: ")
	if !ok || format == "" {
		return
	}
	path, ok := a.ask("Путь к файлу (Enter — report.<формат>): ")
	if !ok {
		return
	}
	if path == "" {
		path = "report." + format
	}
	if err := a.facade.ExportSummaryReport(format, path); err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintf(a.out, "Ведомость сохранена: %s\n", path)
}

func (a *ConsoleApp) runDemo() {
	fmt.Fprintln(a.out, "=== Автоматический демо-сценарий ===")
	prefix := a.nextDemoPrefix()

	student1, err := a.facade.RegisterStudent(prefix+"-s1", "Иван Петров")
	if err != nil {
		a.printError(err)
		return
	}
	student2, err := a.facade.RegisterStudent(prefix+"-s2", "Мария Смирнова")
	if err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintf(a.out, "Добавлены студенты: %s (%s), %s (%s)\n", student1.Name(), student1.ID(), student2.Name(), student2.ID())

	homework, err := a.facade.AddCriterion(prefix+"-hw", "Домашняя работа", 10, 0.3)
	if err != nil {
		a.printError(err)
		return
	}
	test, err := a.facade.AddCriterion(prefix+"-test", "Контрольная работа", 20, 0.3)
	if err != nil {
		a.printError(err)
		return
	}
	exam, err := a.facade.AddCriterion(prefix+"-exam", "Экзамен", 10, 0.4)
	if err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintf(
		a.out,
		"Созданы критерии: %s (макс. %.0f, вес %.1f), %s (макс. %.0f, вес %.1f), %s (макс. %.0f, вес %.1f)\n",
		homework.Name(), homework.MaxScore().Value(), homework.Weight().Value(),
		test.Name(), test.MaxScore().Value(), test.Weight().Value(),
		exam.Name(), exam.MaxScore().Value(), exam.Weight().Value(),
	)

	fmt.Fprintln(a.out, "Проверка инвариантов:")
	if _, err := a.facade.AssignEvaluation(prefix+"-invalid-max", student1.ID(), homework.ID(), 12); err != nil {
		fmt.Fprintf(a.out, "  Отказ при 12 из 10: %v\n", err)
	} else {
		fmt.Fprintln(a.out, "  ОШИБКА: балл 12 из 10 был принят")
	}
	if _, err := a.facade.AssignEvaluation(prefix+"-invalid-criterion", student1.ID(), prefix+"-missing", 5); err != nil {
		fmt.Fprintf(a.out, "  Отказ по несуществующему критерию: %v\n", err)
	} else {
		fmt.Fprintln(a.out, "  ОШИБКА: оценка по несуществующему критерию была принята")
	}

	assign := func(id string, student *domain.Student, criterion *domain.Criterion, score float64) bool {
		if _, err := a.facade.AssignEvaluation(id, student.ID(), criterion.ID(), score); err != nil {
			a.printError(err)
			return false
		}
		return true
	}
	if !assign(prefix+"-e1", student1, homework, 9) ||
		!assign(prefix+"-e2", student1, test, 16) ||
		!assign(prefix+"-e3", student1, exam, 7) {
		return
	}
	if !assign(prefix+"-e4", student2, homework, 6) ||
		!assign(prefix+"-e5", student2, test, 12) ||
		!assign(prefix+"-e6", student2, exam, 8) {
		return
	}
	fmt.Fprintln(a.out, "Выставлены валидные оценки")

	students := []*domain.Student{student1, student2}
	a.printDemoTotals(students, domain.WeightedTenPointStrategyName)
	if err := a.facade.SwitchCalculationStrategy(domain.RawSumStrategyName); err != nil {
		a.printError(err)
		return
	}
	a.printDemoTotals(students, domain.RawSumStrategyName)
	if err := a.facade.SwitchCalculationStrategy(domain.WeightedTenPointStrategyName); err != nil {
		a.printError(err)
		return
	}
	if err := a.facade.ExportSummaryReport("csv", "report.csv"); err != nil {
		a.printError(err)
		return
	}
	fmt.Fprintln(a.out, "Ведомость сохранена в report.csv (стратегия weighted-ten-point)")
}

func (a *ConsoleApp) printDemoTotals(students []*domain.Student, strategy string) {
	fmt.Fprintf(a.out, "Итоги по стратегии %q:\n", strategy)
	for _, student := range students {
		total, err := a.facade.CalculateStudentTotal(student.ID())
		if err != nil {
			a.printError(err)
			continue
		}
		fmt.Fprintf(a.out, "  %s (%s): %.2f\n", student.Name(), student.ID(), total.Value())
	}
}

func (a *ConsoleApp) findStudent(id string) (*domain.Student, error) {
	students, err := a.facade.ListStudents()
	if err != nil {
		return nil, err
	}
	trimmedID := strings.TrimSpace(id)
	for _, student := range students {
		if student.ID() == trimmedID {
			return student, nil
		}
	}
	return nil, fmt.Errorf("%w: %q", application.ErrStudentNotFound, trimmedID)
}

func (a *ConsoleApp) nextEvaluationID() string {
	evaluations, err := a.facade.ListAllEvaluations()
	if err != nil {
		return "e1"
	}
	maxNumber := 0
	for _, evaluation := range evaluations {
		id := evaluation.ID()
		if !strings.HasPrefix(id, "e") {
			continue
		}
		number, err := strconv.Atoi(strings.TrimPrefix(id, "e"))
		if err != nil {
			continue
		}
		if number > maxNumber {
			maxNumber = number
		}
	}
	return fmt.Sprintf("e%d", maxNumber+1)
}

func (a *ConsoleApp) nextDemoPrefix() string {
	students, err := a.facade.ListStudents()
	if err != nil {
		return "demo-1"
	}
	maxRun := 0
	for _, student := range students {
		parts := strings.SplitN(student.ID(), "-", 3)
		if len(parts) < 2 || parts[0] != "demo" {
			continue
		}
		run, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		if run > maxRun {
			maxRun = run
		}
	}
	return fmt.Sprintf("demo-%d", maxRun+1)
}
