package stats

import (
	"fmt"
	"os"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Stats struct {
	states    sync.Map
	workflows sync.Map
}

func NewStats() *Stats {
	s := Stats{}
	s.states = sync.Map{}
	s.workflows = sync.Map{}
	return &s
}

func (s *Stats) AddState(state string) {
	s.states.Store(state, 0)
}

func (s *Stats) IncState(state string) {
	cr, ok := s.states.Load(state)
	if !ok {
		s.states.Store(state, 1)
	}
	s.states.Store(state, cr.(int)+1)
}

func (s *Stats) AddWorkflow(workflow string) {
	cr, ok := s.workflows.Load(workflow)
	if !ok {
		s.workflows.Store(workflow, 1)
		return
	}
	s.workflows.Store(workflow, cr.(int)+1)
}

func (s *Stats) PrintWorkflows() {
	fmt.Println("\n ")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Workflow", "Count"})
	total := 0
	s.workflows.Range(func(key, value interface{}) bool {
		t.AppendRow([]interface{}{key, value})
		total += value.(int)
		return true
	})
	t.AppendSeparator()
	t.AppendFooter(table.Row{"Total", total})
	t.Render()
}

func (s *Stats) PrintStates() {
	fmt.Println("\n ")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Status", "Produced Events"})
	total := 0
	s.states.Range(func(key, value interface{}) bool {
		t.AppendRow([]interface{}{key, value})
		total += value.(int)
		return true
	})
	t.AppendSeparator()
	t.AppendFooter(table.Row{"Total", total})
	t.Render()
}
