package alerting

import (
	"strconv"
	"strings"
	"time"

	"github.com/code-creatively/datav/backend/pkg/models"
)

type evalHandler interface {
	Eval(evalContext *models.EvalContext)
}

// DefaultEvalHandler is responsible for evaluating the alert rule.
type DefaultEvalHandler struct {
	alertJobTimeout time.Duration
}

// NewEvalHandler is the `DefaultEvalHandler` constructor.
func NewEvalHandler() *DefaultEvalHandler {
	return &DefaultEvalHandler{
		alertJobTimeout: time.Second * 5,
	}
}

// Eval evaluated the alert rule.
func (e *DefaultEvalHandler) Eval(context *models.EvalContext) {
	firing := true
	noDataFound := true
	conditionEvals := ""

	for i := 0; i < len(context.Rule.Conditions); i++ {
		condition := context.Rule.Conditions[i]
		cr, err := condition.Eval(context)
		if err != nil {
			context.Error = err
		}

		// break if condition could not be evaluated
		if context.Error != nil {
			break
		}

		if i == 0 {
			firing = cr.Firing
			noDataFound = cr.NoDataFound
		}

		// calculating Firing based on operator
		if cr.Operator == "or" {
			firing = firing || cr.Firing
			noDataFound = noDataFound || cr.NoDataFound
		} else {
			firing = firing && cr.Firing
			noDataFound = noDataFound && cr.NoDataFound
		}

		if i > 0 {
			conditionEvals = "[" + conditionEvals + " " + strings.ToUpper(cr.Operator) + " " + strconv.FormatBool(cr.Firing) + "]"
		} else {
			conditionEvals = strconv.FormatBool(firing)
		}

		context.EvalMatches = append(context.EvalMatches, cr.EvalMatches...)
	}

	context.ConditionEvals = conditionEvals + " = " + strconv.FormatBool(firing)
	context.Firing = firing
	context.NoDataFound = noDataFound
	context.EndTime = time.Now()
}