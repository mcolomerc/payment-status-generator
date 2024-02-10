package stats

import (
	"fmt"
	"os"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Stats struct {
	sync      sync.Mutex
	states    map[string]int
	workflows map[string]int
	banks     map[string]int
}

func NewStats() *Stats {
	s := Stats{}
	s.states = make(map[string]int)
	s.workflows = make(map[string]int)
	s.banks = make(map[string]int)
	return &s
}

func (s *Stats) AddState(state string) {
	s.states[state] = 0
}

func (s *Stats) IncState(state string) {
	s.sync.Lock()
	s.states[state] += 1
	s.sync.Unlock()
}

func (s *Stats) AddWorkflow(workflow string) {
	s.sync.Lock()
	s.workflows[workflow] += 1
	s.sync.Unlock()
}

func (s *Stats) AddBank(bank string) {
	s.sync.Lock()
	s.banks[bank] += 1
	s.sync.Unlock()
}

func (s *Stats) Print() {
	s.PrintWorkflows()
	s.PrintStates()
	s.PrintBanks()
	fmt.Println("\n ")
}

func (s *Stats) PrintWorkflows() {
	fmt.Println("\n ")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Workflow", "Count"})
	total := 0
	for workflow, count := range s.workflows {
		t.AppendRow([]interface{}{workflow, count})
		total += count
	}
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

	for state, count := range s.states {
		t.AppendRow([]interface{}{state, count})
		total += count
	}
	t.AppendSeparator()
	t.AppendFooter(table.Row{"Total", total})
	t.Render()
}

func (s *Stats) PrintBanks() {
	fmt.Println("\n ")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Bank", "Updated times"})

	for bank, count := range s.banks {
		t.AppendRow([]interface{}{bank, count})

	}
	t.AppendSeparator()

	t.Render()
}
