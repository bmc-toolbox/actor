package executor

import "fmt"

type (
	PlanMaker struct {
		executorFactory ExecutorFactory
	}

	ActionResult struct {
		Action  string
		Status  bool
		Message string
		Error   error
	}

	ExecutorFactory interface {
		New(params map[string]interface{}) (Executor, error)
	}

	Executor interface {
		ActionToFn(string) (ActionFn, error)
		Cleanup()
	}

	ExecutionPlan struct {
		executor  Executor
		actionFns []ActionFn
	}

	ActionFn func() ActionResult
)

func NewPlanMaker(executorFactory ExecutorFactory) *PlanMaker {
	return &PlanMaker{executorFactory: executorFactory}
}

func (e *PlanMaker) MakePlan(actions []string, params map[string]interface{}) (*ExecutionPlan, error) {
	executor, err := e.executorFactory.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to make an execution plan: %w", err)
	}

	plan := newExecutionPlan(executor)

	for _, action := range actions {
		fn, err := executor.ActionToFn(action)
		if err != nil {
			return nil, err
		}
		plan.actionFns = append(plan.actionFns, fn)
	}

	return plan, nil
}

func newExecutionPlan(executor Executor) *ExecutionPlan {
	return &ExecutionPlan{executor: executor, actionFns: make([]ActionFn, 0)}
}

func (p *ExecutionPlan) Run() ([]ActionResult, error) {
	defer p.executor.Cleanup()

	results := make([]ActionResult, 0)

	for _, fn := range p.actionFns {
		result := fn()
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
