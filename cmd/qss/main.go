package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {
	var err error

	if err = ui.Init(); err != nil {
		exit(1, err)
	}

	termWidth, termHeight := ui.TerminalDimensions()

	ls := widgets.NewList()
	ls.SelectedRowStyle = ui.NewStyle(ui.ColorCyan)
	ls.Rows, err = prepareAssociationList()
	if err != nil {
		exit(1, err)
	}
	ls.Title = "State Associations"
	ls.WrapText = true

	ils := widgets.NewList()
	initialILSRow, err := getAssociationTargets(ls.Rows[0])
	if err != nil {
		exit(1, err)
	}
	ils.Rows = initialILSRow
	ils.Title = "Target Selector"

	bc := widgets.NewBarChart()
	bc.BarWidth = (termWidth - 10) / 2 / 4
	bc.Title = fmt.Sprintf("Target states of the association: %s", ls.Rows[0])
	bc.Labels = []string{"Success", "Failed", "Pending", "Skipped"}
	bc.Data, err = calculateStatusBarChartData(ls.Rows[0])
	if err != nil {
		exit(1, err)
	}
	bc.BarColors = []ui.Color{ui.ColorGreen, ui.ColorRed, ui.ColorYellow, ui.ColorCyan}
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorBlack)}
	bc.PaddingRight = 1
	bc.PaddingLeft = 1

	grid := ui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(0.5, ls),
			ui.NewCol(0.5,
				ui.NewRow(0.5, bc),
				ui.NewRow(0.5, ils),
			),
		),
	)

	ui.Render(grid)

	uiEvents := ui.PollEvents()
	var previousKey string
	var selectedAssociation string
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				exit(0, nil)
			case "<Down>", "j":
				ls.ScrollDown()
			case "<Up>", "k":
				ls.ScrollUp()
			case "<Home>":
				ls.ScrollTop()
			case "g":
				if previousKey == "g" {
					previousKey = ""
					ls.ScrollTop()
				}
			case "<End>", "G":
				ls.ScrollBottom()
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				bc.BarWidth = (payload.Width - 10) / 2 / 4
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
			case "r":
				ls.Rows, err = prepareAssociationList()
				if err != nil {
					exit(1, err)
				}
			case "<Enter>":
				selectedAssociation = ls.Rows[ls.SelectedRow]
				bc.Data, err = calculateStatusBarChartData(selectedAssociation)
				if err != nil {
					exit(1, err)
				}
				ilsRows, err := getAssociationTargets(selectedAssociation)
				if err != nil {
					exit(1, err)
				}
				ils.Rows = ilsRows
				bc.Title = fmt.Sprintf("Target states of the association: %s", selectedAssociation)
			}
			previousKey = e.ID
			ui.Render(grid)
		case <-ticker.C:
			bc.Data, err = calculateStatusBarChartData(selectedAssociation)
			if err != nil {
				exit(1, err)
			}
			ilsRows, err := getAssociationTargets(selectedAssociation)
			if err != nil {
				exit(1, err)
			}
			ils.Rows = ilsRows
			ui.Render(grid)
		}
	}
}

func prepareAssociationList() ([]string, error) {
	associations, err := listAssociations()
	if err != nil {
		return nil, err
	}

	var associationNames []string
	for _, a := range associations.Associations {
		associationNames = append(associationNames, fmt.Sprintf("%s %s", *a.AssociationId, *a.AssociationName))
	}
	return associationNames, nil
}

func calculateStatusBarChartData(associationString string) ([]float64, error) {
	associationID := strings.Split(associationString, " ")[0]
	a, err := getAssociation(associationID)
	if err != nil {
		return nil, err
	}
	s := getAssociationSuccessTargets(a)
	f := getAssociationFailedTargets(a)
	p := getAssociationPendingTargets(a)
	sk := getAssociationSkippedTargets(a)
	data := []float64{s, f, p, sk}
	if s+f+p == 0 {
		//data = append(data, 1)
		data = []float64{}
	}
	return data, nil
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
