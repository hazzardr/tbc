package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/wowsims/tbc/sim"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	balance "github.com/wowsims/tbc/sim/druid/balance"
)

func init() {
	sim.RegisterAll()
}

// buildRequest constructs a RaidSimRequest using the Balance Druid P1 preset.
func buildRequest() *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Race:      proto.Race_RaceTauren,
				Class:     proto.Class_ClassDruid,
				Equipment: balance.P1Gear,
				Consumes:  balance.FullConsumes,
				Spec:      balance.PlayerOptionsAdaptive,
				Buffs:     balance.FullIndividualBuffs,
			},
			balance.FullPartyBuffs,
			balance.FullRaidBuffs,
			balance.FullDebuffs,
		),
		Encounter: &proto.Encounter{
			Duration: 300,
			Targets:  []*proto.Target{core.NewDefaultTarget()},
		},
		SimOptions: &proto.SimOptions{
			Iterations: 500,
			RandomSeed: 101,
		},
	}
}

// --- bubbletea model ---

type simResult struct {
	dps float64
	err string
}

type model struct {
	result *simResult
}

func (m model) Init() tea.Cmd {
	return runSim
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case simResult:
		m.result = &msg
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var s string
	if m.result == nil {
		s = "Running Balance Druid sim...\n"
	} else if m.result.err != "" {
		s = fmt.Sprintf("Error: %s\n\nPress q to quit.\n", m.result.err)
	} else {
		s = fmt.Sprintf("Balance Druid (P1 Adaptive) — 300s, 500 iterations\n\nDPS: %.2f\n\nPress q to quit.\n", m.result.dps)
	}
	return tea.NewView(s)
}

// runSim is a bubbletea command: runs in the background and returns a simResult message.
func runSim() tea.Msg {
	result := core.RunRaidSim(buildRequest())
	if result.GetErrorResult() != "" {
		return simResult{err: result.GetErrorResult()}
	}
	dps := result.GetRaidMetrics().GetDps().GetAvg()
	return simResult{dps: dps}
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
