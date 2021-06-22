package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Deichindianer/quick-ssm-state/internal/data"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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

	bc, err := data.NewStatusBarChart(termWidth, associationList.Rows[0])
	if err != nil {
		exit(1, err)
	}

	grid := ui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(0.5, associationList),
			ui.NewCol(0.5,
				ui.NewRow(0.5, bc),
				ui.NewRow(0.5, targetList),
			),
		),
	)
	mainScreen := &mainScreen{
		grid:            grid,
		associationList: associationList,
		targetList:      targetList,
		statusBarChart:  bc,
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

func prepareAssociationList() ([]string, error) {
	associations, err := listAssociations()
	if err != nil {
		return nil, err
	}

	if associations == nil {
		return nil, fmt.Errorf("associations list is empty")
	}

	var associationNames []string
	for _, a := range associations.Associations {
		if a.AssociationName == nil {
			a.AssociationName = aws.String("None")
		}
		if a.AssociationId == nil {
			exit(1, errors.New("AssociationID is nil, wtf man"))
		}
		associationNames = append(associationNames, fmt.Sprintf("%s %s", *a.AssociationId, *a.AssociationName))
	}
	return associationNames, nil
}

func listAssociations() (*ssm.ListAssociationsOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	awsClient := ssm.NewFromConfig(cfg)
	associations, err := awsClient.ListAssociations(context.Background(), &ssm.ListAssociationsInput{})
	if err != nil {
		return nil, err
	}
	return associations, nil
}

func getAssociation(associationID string) (*ssm.DescribeAssociationOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	awsClient := ssm.NewFromConfig(cfg)
	association, err := awsClient.DescribeAssociation(context.Background(), &ssm.DescribeAssociationInput{
		AssociationId: aws.String(associationID),
	})
	if err != nil {
		return nil, err
	}
	return association, nil
}

func getAssociationPendingTargets(association *ssm.DescribeAssociationOutput) float64 {
	return float64(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Pending"])
}

func getAssociationSuccessTargets(association *ssm.DescribeAssociationOutput) float64 {
	return float64(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Success"])
}

func getAssociationFailedTargets(association *ssm.DescribeAssociationOutput) float64 {
	return float64(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Failed"])
}

func getAssociationSkippedTargets(association *ssm.DescribeAssociationOutput) float64 {
	return float64(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Skipped"])
}

func getAssociationTargets(associationString string) ([]string, error) {
	associationID := strings.Split(associationString, " ")[0]
	a, err := getAssociation(associationID)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, t := range a.AssociationDescription.Targets {
		vals := strings.Join(t.Values, ", ")
		result = append(result, fmt.Sprintf("%s:%s", *t.Key, vals))
	}
	return result, nil
}

func exit(exitCode int, err error) {
	ui.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(exitCode)
}
