package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Deichindianer/quick-ssm-state/internal/data"

	ui "github.com/gizak/termui/v3"
)

type mainScreen struct {
	grid            *ui.Grid
	associationList *data.AssociationList
	targetList      *data.TargetList
	statusBarChart  *data.StatusBarChart
}

func main() {
	var err error

	if err = ui.Init(); err != nil {
		exit(1, err)
	}

	UIBusyloop(generateGrid())
}

func generateGrid() *mainScreen {
	var err error
	termWidth, termHeight := ui.TerminalDimensions()

	associationList, err := data.NewAssociationList()
	if err != nil {
		exit(1, err)
	}

	targetList, err := data.NewTargetList(associationList.Rows[0])
	if err != nil {
		exit(1, err)
	}

	statusBarChart, err := data.NewStatusBarChart(termWidth, associationList.Rows[0])
	if err != nil {
		exit(1, err)
	}

	grid := ui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(0.5, associationList),
			ui.NewCol(0.5,
				ui.NewRow(0.5, statusBarChart),
				ui.NewRow(0.5, targetList),
			),
		),
	)
	mainScreen := &mainScreen{
		grid:            grid,
		associationList: associationList,
		targetList:      targetList,
		statusBarChart:  statusBarChart,
	}
	return mainScreen
}

func UIBusyloop(ms *mainScreen) {
	ui.Render(ms.grid)
	uiEvents := ui.PollEvents()
	var previousKey string
	selectedAssociation := ms.associationList.Rows[0]
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				exit(0, nil)
			case "<Down>", "j":
				ms.associationList.ScrollDown()
			case "<Up>", "k":
				ms.associationList.ScrollUp()
			case "<Home>":
				ms.associationList.ScrollTop()
			case "g":
				if previousKey == "g" {
					previousKey = ""
					ms.associationList.ScrollTop()
				}
			case "<End>", "G":
				ms.associationList.ScrollBottom()
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				ms.statusBarChart.BarWidth = (payload.Width - 10) / 2 / 4
				ms.grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
			case "r":
				if err := ms.associationList.Reload(); err != nil {
					exit(1, err)
				}
			case "<Enter>":
				selectedAssociation = ms.associationList.Rows[ms.associationList.SelectedRow]
				if err := ms.statusBarChart.Reload(selectedAssociation); err != nil {
					exit(1, err)
				}
				if err := ms.targetList.Reload(selectedAssociation); err != nil {
					exit(1, err)
				}
			}
			previousKey = e.ID
			ui.Render(ms.grid)
		case <-ticker.C:
			if err := ms.statusBarChart.Reload(selectedAssociation); err != nil {
				exit(1, err)
			}
			if err := ms.targetList.Reload(selectedAssociation); err != nil {
				exit(1, err)
			}
			ui.Render(ms.grid)
		}
	}
}

func exit(exitCode int, err error) {
	ui.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(exitCode)
}
