package actions

import "fmt"

type (
	Action struct {
		value    string
		executor Executor
	}

	PlanMaker struct {
		executorFactories []ExecutorFactory
	}

	ActionResult struct {
		Action  string
		Status  bool
		Message string
		Error   error
	}

	ExecutorFactory interface {
		New(map[string]interface{}) (Executor, error)
	}

	Executor interface {
		Validate(string) error
		Run(string) ActionResult
		Cleanup()
	}

	ExecutionPlan struct {
		actions    []Action
		cleanupFns []func()
	}

	ActionFn func() ActionResult
)

func NewPlanMaker(executorFactories ...ExecutorFactory) *PlanMaker {
	return &PlanMaker{executorFactories: executorFactories}
}

func (e *PlanMaker) MakePlan(actionsRaw []string, params map[string]interface{}) (*ExecutionPlan, error) {
	executors := make([]Executor, 0)
	cleanupFns := make([]func(), 0)

	for _, executorFactory := range e.executorFactories {
		executor, err := executorFactory.New(params)
		if err != nil {
			return nil, fmt.Errorf("failed to make an execution plan: %w", err)
		}
		executors = append(executors, executor)
		cleanupFns = append(cleanupFns, executor.Cleanup)
	}

	actions := make([]Action, len(actionsRaw))

	for i, action := range actionsRaw {
		executor := findExecutor(action, executors)
		if executor == nil {
			return nil, fmt.Errorf("action %q is unknown", action)
		}
		actions[i] = Action{value: action, executor: executor}
	}

	return &ExecutionPlan{actions: actions, cleanupFns: cleanupFns}, nil
}

func (p *ExecutionPlan) Run() ([]ActionResult, error) {
	defer func() {
		for _, cleanupFn := range p.cleanupFns {
			cleanupFn()
		}
	}()

	results := make([]ActionResult, 0)

	for _, action := range p.actions {
		result := action.executor.Run(action.value)
		results = append(results, result)
		if result.Error != nil {
			return results, result.Error
		}
	}

	return results, nil
}

func NewActionResult(action string, status bool, message string, err error) ActionResult {
	return ActionResult{
		Action:  action,
		Status:  status,
		Message: message,
		Error:   err,
	}
}

func findExecutor(action string, executors []Executor) Executor {
	for _, executor := range executors {
		if err := executor.Validate(action); err == nil {
			return executor
		}
	}
	return nil
}
