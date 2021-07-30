package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/Deichindianer/quick-ssm-state/internal/data"

	ui "github.com/gizak/termui/v3"
)

type mainScreen struct {
	grid            *ui.Grid
	associationList *data.AssociationList
	targetList      *data.TargetList
	statusBarChart  *data.StatusBarChart
	outputParagraph *data.OutputParagraph
}

func main() {
	var err error

	if err = ui.Init(); err != nil {
		exit(1, err)
	}
	mainScreen, err := generateMainScreen()
	if err != nil {
		exit(1, err)
	}
	UIBusyloop(mainScreen)
}

func generateMainScreen() (*mainScreen, error) {
	var err error
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	aws.String("test")
	ssmClient := ssm.NewFromConfig(cfg)
	termWidth, termHeight := ui.TerminalDimensions()

	associationList, err := data.NewAssociationList(ssmClient)
	if err != nil {
		return nil, err
	}

	targetList, err := data.NewTargetList(ssmClient, associationList.Rows[0])
	if err != nil {
		return nil, err
	}

	statusBarChart, err := data.NewStatusBarChart(ssmClient, termWidth, associationList.Rows[0])
	if err != nil {
		return nil, err
	}

	outputParagraph, err := data.NewOutputParagraph(ssmClient, associationList.Rows[0])
	if err != nil {
		return nil, err
	}

	grid := ui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(0.5,
				ui.NewRow(0.4, associationList),
				ui.NewRow(0.6, outputParagraph),
			),
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
		outputParagraph: outputParagraph,
	}
	return mainScreen, nil
}

func UIBusyloop(ms *mainScreen) {
	defer recoverPanic()
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
				if err := ms.outputParagraph.Reload(selectedAssociation); err != nil {
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
			if err := ms.outputParagraph.Reload(selectedAssociation); err != nil {
				exit(1, err)
			}
			ui.Render(ms.grid)
		}
	}
}

func recoverPanic() {
	r := recover()
	var err error
	if r != nil {
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("unknown panic")
		}
		exit(1, err)
	}
}

func exit(exitCode int, err error) {
	ui.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(exitCode)
}
