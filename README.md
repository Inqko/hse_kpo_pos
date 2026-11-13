# ПОС — Преподавательская Оценка Студентов

Консольное приложение на Go для учёта преподавательской оценки студентов: студенты, рубрикатор критериев, выставленные оценки и расчёт итогового балла. Проект демонстрирует Domain-Driven Design, архитектуру Ports & Adapters (Hexagonal) и три уместных паттерна GoF: Factory, Facade и Strategy.

Приложение однопоточное: горутины и `sync.Mutex` не используются. Данные хранятся в памяти и живут до завершения процесса; итоговая ведомость выгружается в файл CSV или JSON.

## Возможности

- добавление студентов и просмотр списка;
- ведение рубрикатора: название критерия, максимальный балл и вес;
- выставление, изменение и отмена оценок;
- запрет оценки по критерию, которого нет в рубрикаторе;
- запрет балла, превышающего максимум критерия (нельзя поставить 12 из 10);
- запрет суммарного веса критериев больше 1.0;
- расчёт итога по двум стратегиям: взвешенная 10-балльная шкала и сырая сумма;
- экспорт итоговой ведомости в CSV и JSON;
- автоматический сквозной демо-сценарий для показа всех ключевых правил.

## Запуск

Требуется Go версии, указанной в `go.mod`.

```bash
go run ./cmd/pos
```

Сборка бинарника:

```bash
go build ./...
go build -o pos ./cmd/pos
./pos
```

Проверка тестов:

```bash
go test -v ./...
```

Дополнительно:

```bash
go vet ./...
gofmt -l .
```

## Быстрый сценарий

1. Запустить `go run ./cmd/pos`.
2. Выбрать пункт `10` — демо само создаст двух студентов и три критерия с весами `0.3`, `0.3`, `0.4`.
3. Посмотреть в выводе отказы при `12 из 10` и при несуществующем критерии.
4. Увидеть итоги по стратегиям `weighted-ten-point` и `raw-sum`.
5. В корне проекта появится `report.csv`.
6. Выбрать `0` для выхода.

### Меню

| Пункт | Действие |
|---|---|
| 1 | Добавить студента |
| 2 | Список студентов |
| 3 | Добавить критерий в рубрикатор |
| 4 | Показать рубрикатор критериев |
| 5 | Выставить оценку студенту |
| 6 | Изменить или отменить оценку |
| 7 | Показать оценки студента и рассчитать его итог |
| 8 | Сменить стратегию расчёта итога |
| 9 | Экспортировать итоговую ведомость |
| 10 | Запустить автоматический сквозной демо-сценарий |
| 0 | Выход |

## Словарь предметной области (Ubiquitous Language)

| Термин | Значение | Реализация |
|---|---|---|
| Студент | Кого оценивают на курсе | `domain.Student`, Entity |
| Критерий | За что ставят балл: название, максимум, вклад в итог | `domain.Criterion`, Entity |
| Рубрикатор | Текущая совокупность критериев курса | Набор `*domain.Criterion` в `CriterionRepository`; отдельного класса рубрики нет |
| Оценка | Факт «этому студенту по этому критерию поставили такой балл» | `domain.Evaluation`, Entity |
| Балл | Числовое значение оценки | `domain.Score`, Value Object |
| Вес | Вклад критерия в итог, `(0, 1]` | `domain.Weight`, Value Object |
| Итог | Результат по студенту, посчитанный из его оценок и рубрикатора | `domain.GradeCalculator`, Domain Service; результат — `domain.Score` |

## Карта DDD

### Entity

Entity обладает идентичностью, которая важнее набора атрибутов, и имеет жизненный цикл: объект можно найти по ID, изменить или удалить.

| Сущность | Идентификатор | Почему Entity |
|---|---|---|
| `Student` | `id` | Два студента с одинаковым именем — разные люди; запись живёт в хранилище и находится по ID |
| `Criterion` | `id` | Критерий входит в рубрикатор и на него ссылаются оценки; после создания его максимум и вес не меняются |
| `Evaluation` | `id` | Оценка имеет собственный ID, ссылается на студента и критерий по их ID и может быть изменена или отменена |

`Evaluation` ссылается на `Student` и `Criterion` по строковым ID, а не хранит указатели. Это снижает связность и позволяет хранить оценки отдельно от студентов и рубрикатора. Согласованность ссылок обеспечивают сценарии приложения и проверка в доменном сервисе.

### Value Object

Value Object определяется значением, а не идентичностью. Он неизменяем: после создания его нельзя поменять, можно только заменить новым значением.

| Объект | Поле | Почему Value Object |
|---|---|---|
| `Score` | `value float64` | Балл 8 — это просто 8; два балла 8 неразличимы. Методов изменения нет, есть только `Value()` и `LessOrEqual()` |
| `Weight` | `value float64` | Вес 0.3 определяется числом; важна только корректность диапазона `(0, 1]` |

`Score` не ограничивает значение сверху, потому что максимум зависит от критерия: у одного критерия 10 баллов, у другого 20 или 100. Верхняя граница проверяется там, где известен критерий: в `EvaluationFactory`, `Evaluation.UpdateScore` и при расчёте итога.

### Инварианты и места их проверки

| Инвариант | Где охраняется |
|---|---|
| `Score` не отрицательный и не `NaN`/`Inf` | `domain.NewScore` |
| `Weight` в диапазоне `(0, 1]` | `domain.NewWeight` и дополнительно `domain.NewCriterion` |
| `id` и `name` не пустые после `TrimSpace` | `NewStudent`, `NewCriterion`, `newEvaluation` |
| Максимальный балл критерия больше нуля | `NewCriterion` |
| Балл не превышает максимум критерия | `EvaluationFactory.CreateEvaluation`, `Evaluation.UpdateScore`, стратегии расчёта |
| Оценку можно выставить только по существующему критерию | `POSFacade.AssignEvaluation` ищет критерий в репозитории; `GradeCalculator` повторно сверяет оценки с рубрикатором |
| Нельзя создать оценку в обход проверок | конструктор `newEvaluation` не экспортируется из пакета `domain`; снаружи доступна только `EvaluationFactory` |
| Нельзя изменить оценку по чужому критерию | `Evaluation.UpdateScore` сверяет `criterion.ID()` с `criterionID` оценки |
| Не более одной оценки студента по одному критерию | `POSFacade.AssignEvaluation` |
| Суммарный вес критериев не больше 1.0 | `POSFacade.AddCriterion` (с допуском `1e-9` на погрешность `float64`) |
| Дублирование ID студента или критерия запрещено | `POSFacade.RegisterStudent`, `POSFacade.AddCriterion` |

Проверка суммы весов находится в сценарии `AddCriterion`, потому что для неё нужно прочитать весь рубрикатор из репозитория. Это правило уровня агрегата-рубрикатора; отдельный класс `Rubric` не вводился, чтобы не усложнять модель без необходимости.

### Domain Service

`GradeCalculator` — доменный сервис, потому что расчёт итога:

- не принадлежит одной сущности: нужны все оценки студента и весь рубрикатор;
- не является Value Object: это поведение над набором объектов;
- не должен жить в UI или инфраструктуре: правило расчёта — часть предметной области.

Сервис хранит доступные стратегии и активную стратегию, умеет переключать правило и рассчитывать итог. Перед расчётом он проверяет, что каждая оценка ссылается на критерий из переданного рубрикатора.

Формулы:

- `WeightedTenPointStrategy`: `round2(Σ (score_i / max_i) * weight_i * 10)`;
- `RawSumStrategy`: `round2(Σ score_i * weight_i)`;
- `round2(x) = math.Round(x * 100) / 100`.

### Repository (порты хранения)

Интерфейсы репозиториев объявлены в слое приложения (`internal/application/ports.go`), а реализации — в инфраструктуре (`internal/infrastructure/memory`). Домен ничего не знает о способе хранения.

- `StudentRepository` — сохранение, поиск по ID, список студентов;
- `CriterionRepository` — сохранение, поиск по ID, список критериев;
- `EvaluationRepository` — сохранение, поиск по ID, поиск по студенту и критерию, оценки студента, все оценки, удаление.

## Архитектура Ports & Adapters

```
+------------------------------------------------------+
| cmd/pos + internal/di  (Composition Root)            |
|   собирает граф зависимостей через uber-go/dig       |
+---------------------------+--------------------------+
                            |
+---------------------------v--------------------------+
| infrastructure (адаптеры)                            |
|   cli.ConsoleApp                                     |
|   memory.InMemory*Repository                         |
|   export.CSVReportExporter / JSONReportExporter      |
+---------------------------+--------------------------+
                            | реализует порты
+---------------------------v--------------------------+
| application (use cases + порты)                      |
|   POSFacade                                          |
|   StudentRepository, CriterionRepository,            |
|   EvaluationRepository, ReportExporter               |
|   StudentReportRow                                   |
+---------------------------+--------------------------+
                            |
+---------------------------v--------------------------+
| domain (ядро)                                        |
|   Student, Criterion, Evaluation                     |
|   Score, Weight                                      |
|   EvaluationFactory, GradeCalculator                 |
|   CalculationStrategy, стратегии                     |
+------------------------------------------------------+
```

Стрелки зависимостей направлены внутрь: инфраструктура зависит от приложения и домена, приложение — от домена, домен не зависит ни от чего, кроме стандартной библиотеки (`errors`, `fmt`, `strings`, `math`).

### Слои и папки

```
cmd/pos/main.go                 точка входа, запуск через DI
internal/
  domain/                       ядро: сущности, VO, инварианты, фабрика, расчёт
    score.go
    weight.go
    student.go
    criterion.go
    evaluation.go
    factory.go
    calculator.go
  application/                  сценарии использования и выходные порты
    ports.go
    facade.go
  infrastructure/               адаптеры
    memory/                     хранение в памяти
      student_repo.go
      criterion_repo.go
      evaluation_repo.go
    export/                     выгрузка ведомости
      csv_exporter.go
      json_exporter.go
    cli/                        входной адаптер — консоль
      cli.go
  di/                           composition root
    container.go
```

Роли адаптеров:

- `cli.ConsoleApp` — входной (driving) адаптер. Он знает только `*application.POSFacade`: читает ввод, вызывает сценарии, печатает результат. Ни один репозиторий или доменный сервис напрямую из CLI не используется.
- `memory.*Repository` — выходные (driven) адаптеры хранения, реализуют порты приложения.
- `export.CSVReportExporter` и `export.JSONReportExporter` — выходные адаптеры выгрузки, реализуют `application.ReportExporter`.
- `internal/di` — единственное место, которое знает все слои одновременно; зависимости собираются на краю приложения.

## Паттерны GoF

| Группа | Паттерн | Где | Какую боль закрывает |
|---|---|---|---|
| Порождающий | Factory | `domain.EvaluationFactory` | Единая точка создания оценки с проверкой `score <= max_score` и обязательных ссылок на студента и критерий |
| Структурный | Facade | `application.POSFacade` | Изолирует CLI от оркестрации трёх репозиториев, фабрики, калькулятора и экспортёров |
| Поведенческий | Strategy | `domain.CalculationStrategy` и реализации | Позволяет менять формулу итога без изменения `GradeCalculator` и без ветвлений по типу расчёта |

### Factory

Если бы `Evaluation` можно было создать публичным конструктором, любой адаптер мог бы обойти правило «балл не больше максимума критерия». Поэтому конструктор `newEvaluation` приватен для пакета `domain`, а единственный внешний путь создания — `EvaluationFactory.CreateEvaluation`. Фабрика проверяет:

- непустой ID оценки;
- что студент и критерий переданы;
- что балл не превышает максимум критерия.

Фабрика используется в `POSFacade.AssignEvaluation`, поэтому и консоль, и демо-сценарий проходят через одну точку с инвариантами.

### Facade

Консольному интерфейсу не нужно знать, что для выставления оценки надо найти студента, найти критерий, проверить отсутствие дубликата, создать VO, вызвать фабрику и сохранить результат. Всё это скрыто за методами `POSFacade`:

- `RegisterStudent`, `ListStudents`;
- `AddCriterion`, `ListCriteria`;
- `AssignEvaluation`, `UpdateEvaluation`, `CancelEvaluation`, `ListStudentEvaluations`, `ListAllEvaluations`;
- `CalculateStudentTotal`, `SwitchCalculationStrategy`, `GetAvailableStrategies`;
- `ExportSummaryReport`, `GetAvailableExportFormats`.

Фасад не подменяет домен: бизнес-правила остаются в сущностях, фабрике и `GradeCalculator`. Он только задаёт порядок вызовов и границы транзакций сценария.

Экспортёры хранятся в `map[string]ReportExporter`. Чтобы добавить новый формат, достаточно реализовать `ReportExporter` и зарегистрировать его в DI; код фасада менять не нужно (принцип открытости/закрытости).

### Strategy

У расчёта итога есть два взаимозаменяемых правила. Если зашить их в `if/else` внутри калькулятора, добавление третьей формулы потребует правки готового класса и смешает правила в одном месте. Вместо этого:

- `CalculationStrategy` объявляет `Name()` и `Calculate(evaluations, criteria)`;
- `WeightedTenPointStrategy` считает взвешенный итог по 10-балльной шкале;
- `RawSumStrategy` считает простую сумму набранных баллов с учётом весов;
- `GradeCalculator` — контекст, который хранит карту стратегий и активную стратегию, умеет переключаться и делегирует расчёт.

Стратегии не знают ни о хранилище, ни о консоли, поэтому их легко тестировать отдельно.

## DI-контейнер и Composition Root

Используется `go.uber.org/dig`. `internal/di.BuildContainer()` создаёт контейнер и регистрирует провайдеры:

- `memory.NewInMemoryStudentRepository` как `application.StudentRepository`;
- `memory.NewInMemoryCriterionRepository` как `application.CriterionRepository`;
- `memory.NewInMemoryEvaluationRepository` как `application.EvaluationRepository`;
- `domain.NewEvaluationFactory`;
- `[]domain.CalculationStrategy` с `WeightedTenPointStrategy` и `RawSumStrategy`;
- имя стратегии по умолчанию для `domain.NewGradeCalculator`;
- `[]application.ReportExporter` с CSV- и JSON-экспортёрами;
- `application.NewPOSFacade`;
- `cli.NewConsoleApp`.

Для приведения конкретных типов к интерфейсам используется `dig.As(new(application.StudentRepository))` и аналогичные опции. Это позволяет адаптерам зависеть от абстракций, а контейнеру — собирать граф без ручного `new` в точке входа.

Запуск:

```go
container.Invoke(func(app *cli.ConsoleApp) error {
    return app.Run()
})
```

`cmd/pos/main.go` не создаёт зависимости вручную и не содержит бизнес-логики. Он только собирает контейнер, запускает приложение и сообщает об ошибке в `stderr`. Такое разделение даёт тестируемость (в тестах можно собрать фасад с in-memory репозиториями) и отсутствие глобального состояния.

## Обработка ошибок

Домен и приложение возвращают именованные sentinel-ошибки, обёрнутые через `%w`:

- `domain.ErrInvalidScore`, `domain.ErrInvalidWeight`, `domain.ErrInvalidStudent`, `domain.ErrInvalidCriterion`, `domain.ErrInvalidEvaluation`;
- `domain.ErrScoreExceedsMax`, `domain.ErrCriterionMismatch`, `domain.ErrCriterionNotFound`;
- `domain.ErrInvalidStrategy`, `domain.ErrUnknownStrategy`;
- `application.ErrStudentNotFound`, `application.ErrStudentAlreadyExists`, `application.ErrCriterionNotFound`, `application.ErrCriterionAlreadyExists`;
- `application.ErrEvaluationNotFound`, `application.ErrEvaluationAlreadyExists`, `application.ErrWeightSumExceeded`, `application.ErrUnsupportedExportFormat`.

Это позволяет тестам проверять причину через `errors.Is`, а CLI — показывать понятное сообщение пользователю, не разбирая текст ошибки.

## Экспорт ведомости

CSV (`report.csv`):

```csv
student_id,student_name,final_grade,strategy
demo-1-s1,Иван Петров,7.90,weighted-ten-point
demo-1-s2,Мария Смирнова,6.80,weighted-ten-point
```

JSON содержит те же данные в отформатированном виде:

```json
[
  {
    "StudentID": "demo-1-s1",
    "StudentName": "Иван Петров",
    "FinalGrade": 7.9,
    "Strategy": "weighted-ten-point"
  }
]
```

При экспорте для каждого студента заново считаются оценки, итог и фиксируется имя активной стратегии. Итог не хранится как отдельная цифра и всегда согласован с текущим набором оценок и рубрикатором.

## Тесты

Запуск:

```bash
go test -v ./...
```

`internal/domain/domain_test.go`:

- `TestScoreRejectsNegativeAndNonFiniteValues` — отрицательные, `NaN` и `Inf` запрещены, корректные баллы создаются;
- `TestWeightRejectsValuesOutsideRange` — `0`, отрицательные и значения больше `1` запрещены, `1` и `0.3` допустимы;
- `TestEvaluationFactoryRejectsScoreAboveCriterionMax` — фабрика отклоняет `12 из 10` и принимает `10 из 10`;
- `TestEvaluationRejectsScoreAboveMaxOnUpdate` — `UpdateScore` не меняет балл при превышении максимума;
- `TestGradeCalculatorSwitchesStrategy` — `WeightedTenPointStrategy` даёт `6.5`, `RawSumStrategy` — `9`, неизвестная стратегия отклоняется;
- `TestGradeCalculatorRejectsEvaluationWithUnknownCriterion` — расчёт не принимает оценку по критерию вне рубрикатора.

`internal/application/facade_test.go`:

- `TestPOSFacadeRejectsEvaluationForUnknownCriterion` — сценарий выставления оценки по несуществующему критерию возвращает `ErrCriterionNotFound`;
- `TestPOSFacadeRejectsWeightSumAboveOne` — сумма весов `0.6 + 0.5` отклоняется, ровно `1.0` разрешена;
- `TestPOSFacadeRejectsUnknownStudentWhenAssigning` — неизвестный студент отклоняется;
- `TestPOSFacadeRejectsDuplicateEvaluation` — повторная оценка по тому же критерию отклоняется;
- `TestPOSFacadeUpdatesAndCancelsEvaluation` — обновление, запрет `11 из 10` и отмена работают.

## Ограничения и возможные расширения

- Хранилище только in-memory: после выхода данные теряются. Замена на JSON/CSV-хранилище потребует только нового адаптера репозитория и регистрации в DI.
- Нет редактирования и удаления студентов и критериев: по ТЗ это необязательно; при необходимости добавляются методы репозиториев и сценарии фасада.
- Нет импорта ведомостей: общий каркас с разным парсингом можно реализовать через Template Method.
- Нет апелляций, пересдач и нескольких курсов: текущая модель — ядро системы.
- Приложение сознательно однопоточное, без мьютексов; конкурентный доступ не поддерживается и не требуется.

## Политика использования ИИ

ИИ-ассистент использовался как инструмент разработки.

Что сгенерировано агентом:

- юнит-тесты доменных инвариантов и сценариев фасада;
- этот README;
- техническая помощь в реализации
