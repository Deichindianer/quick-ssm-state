package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	termWidth, termHeight := ui.TerminalDimensions()

	ls := widgets.NewList()
	ls.SelectedRowStyle = ui.NewStyle(ui.ColorCyan)
	ls.Rows = prepareAssociationList()
	ls.Title = "State Associations"
	ls.WrapText = true

	ils := widgets.NewList()
	initialILSRow, err := getAssociationTargets(ls.Rows[0])
	if err != nil {
		log.Fatal(err)
	}
	ils.Rows = initialILSRow
	ils.Title = "Target Selector"

	bc := widgets.NewBarChart()
	bc.BarWidth = (termWidth - 10) / 2 / 4
	bc.Title = "All target states of the association"
	bc.Labels = []string{"Success", "Failed", "Pending", "Skipped"}
	bc.Data = calculateStatusBarChartData(ls.Rows[0])
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
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Down>", "j":
				ls.ScrollDown()
				bc.Data = calculateStatusBarChartData(ls.Rows[ls.SelectedRow])
				ilsRows, err := getAssociationTargets(ls.Rows[ls.SelectedRow])
				if err != nil {
					log.Fatal(err)
				}
				ils.Rows = ilsRows
			case "<Up>", "k":
				ls.ScrollUp()
				bc.Data = calculateStatusBarChartData(ls.Rows[ls.SelectedRow])
				ilsRows, err := getAssociationTargets(ls.Rows[ls.SelectedRow])
				if err != nil {
					log.Fatal(err)
				}
				ils.Rows = ilsRows
			case "<Home>":
				ls.ScrollTop()
				bc.Data = calculateStatusBarChartData(ls.Rows[ls.SelectedRow])
				ilsRows, err := getAssociationTargets(ls.Rows[ls.SelectedRow])
				if err != nil {
					log.Fatal(err)
				}
				ils.Rows = ilsRows
			case "g":
				if previousKey == "g" {
					previousKey = ""
					ls.ScrollTop()
					bc.Data = calculateStatusBarChartData(ls.Rows[ls.SelectedRow])
					ilsRows, err := getAssociationTargets(ls.Rows[ls.SelectedRow])
					if err != nil {
						log.Fatal(err)
					}
					ils.Rows = ilsRows
				}
			case "<End>", "G":
				ls.ScrollBottom()
				bc.Data = calculateStatusBarChartData(ls.Rows[ls.SelectedRow])
				ilsRows, err := getAssociationTargets(ls.Rows[ls.SelectedRow])
				if err != nil {
					log.Fatal(err)
				}
				ils.Rows = ilsRows
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				bc.BarWidth = (payload.Width - 10) / 2 / 4
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
			case "r":
				ls.Rows = prepareAssociationList()
			}
			previousKey = e.ID
			ui.Render(grid)
		case <-ticker.C:
			bc.Data = calculateStatusBarChartData(ls.Rows[ls.SelectedRow])
			ilsRows, err := getAssociationTargets(ls.Rows[ls.SelectedRow])
			if err != nil {
				log.Fatal(err)
			}
			ils.Rows = ilsRows
			ui.Render(grid)
		}
	}
}

func prepareAssociationList() []string {
	associations, err := listAssociations()
	if err != nil {
		log.Fatal(err)
	}

	var associationNames []string
	for _, a := range associations.Associations {
		associationNames = append(associationNames, fmt.Sprintf("%s %s", *a.AssociationId, *a.AssociationName))
	}
	return associationNames
}

func calculateStatusBarChartData(associationString string) []float64 {
	associationID := strings.Split(associationString, " ")[0]
	a, err := getAssociation(associationID)
	if err != nil {
		log.Fatal(err)
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
	return data
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
