package datagen

import (
	"mcolomerc/synth-payment-producer/pkg/config"
	"strings"

	"github.com/mroth/weightedrand"
)

type Workflow struct {
	Chooser *weightedrand.Chooser
}

func NewWorkflowHandler(cfg config.Config) Workflow {
	var choices []weightedrand.Choice
	for v, k := range cfg.Datagen.Workflows {
		s := strings.Split(v, ",")
		var workflow []Status
		for st := range s {
			status := strings.TrimSpace(s[st])
			workflow = append(workflow, GetStatus(status))
		}
		choices = append(choices, weightedrand.NewChoice(workflow, uint(k)))
	}
	// Distribution by workflow
	chooser, _ := weightedrand.NewChooser(choices...)

	return Workflow{
		Chooser: chooser,
	}
}

/*
*
Randomly selects an element from some kind of list, where the chances of each element to be selected are not equal,
but rather defined by relative "weights" (or probabilities). This is called weighted random selection.
*
*/
func (w Workflow) GetWorkflow() []Status {
	return w.Chooser.Pick().([]Status)
}
