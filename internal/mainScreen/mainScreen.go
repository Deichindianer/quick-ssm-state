package mainScreen

import (
	"github.com/Deichindianer/quick-ssm-state/internal/data"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ui "github.com/gizak/termui/v3"
)

type MainScreen struct {
	termHeight      int
	termWidth       int
	Grid            *ui.Grid
	AssociationList *data.AssociationList
	TargetList      *data.TargetList
	StatusBarChart  *data.StatusBarChart
}

func New(ssmClient *ssm.Client) (*MainScreen, error) {
	ms := MainScreen{}
	ms.termWidth, ms.termHeight = ui.TerminalDimensions()

	associationList, err := data.NewAssociationList(ssmClient)
	if err != nil {
		return nil, err
	}

	targetList, err := data.NewTargetList(ssmClient, associationList.Rows[0])
	if err != nil {
		return nil, err
	}

	statusBarChart, err := data.NewStatusBarChart(ssmClient, ms.termWidth, associationList.Rows[0])
	if err != nil {
		return nil, err
	}

	ms.AssociationList = associationList
	ms.TargetList = targetList
	ms.StatusBarChart = statusBarChart
	ms.generateGrid(associationList, statusBarChart, targetList)
	return &ms, nil
}

func (ms *MainScreen) generateGrid(associationList *data.AssociationList, statusBarChart *data.StatusBarChart, targetList *data.TargetList) {
	grid := ui.NewGrid()
	grid.SetRect(0, 0, ms.termWidth, ms.termHeight)

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(0.5,
				ui.NewRow(0.4, associationList),
			),
			ui.NewCol(0.5,
				ui.NewRow(0.5, statusBarChart),
				ui.NewRow(0.5, targetList),
			),
		),
	)
	ms.Grid = grid
}
