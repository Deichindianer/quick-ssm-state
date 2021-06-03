package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {

	association, err := getAssociation("")
	if err != nil {
		log.Fatal(err)
	}
	total := getAssociationTotalTargets(association)
	success := getAssociationSuccessTargets(association)
	failed := getAssociationFailedTargets(association)
	pending := getAssociationPendingTargets(association)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	gauge := widgets.NewGauge()
	gauge.Title = "Percent of successful runs"
	gauge.SetRect(100, 10, 150, 14)
	gauge.Percent = (success / total) * 100
	gauge.Label = fmt.Sprintf("%v%%)", gauge.Percent)

	g2 := widgets.NewGauge()
	g2.Title = "Percent of failed runs"
	g2.SetRect(0, 20, 50, 24)
	g2.Percent = (failed / total) * 100
	g2.Label = fmt.Sprintf("%v%%)", g2.Percent)

	g3 := widgets.NewGauge()
	g3.Title = "Percent of pending runs"
	g3.SetRect(0, 10, 50, 14)
	g3.Percent = (pending / total) * 100
	g3.Label = fmt.Sprintf("%v%%)", g3.Percent)

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	associations, err := listAssociations()
	if err != nil {
		log.Fatal(err)
	}

	var associationNames []string
	for _, a := range associations.Associations {
		associationNames = append(associationNames, *a.Name)
	}
	ls := widgets.NewList()
	ls.Rows = associationNames

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(1.0/2, ls),
			ui.NewCol(1.0/2,
				ui.NewRow(1.0/3, gauge),
				ui.NewRow(1.0/3, g2),
				ui.NewRow(1.0/3, g3),
				),
			),
		)

	ui.Render(grid)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}
}

func listAssociations() (*ssm.ListAssociationsOutput, error){
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

func getAssociationTotalTargets(association *ssm.DescribeAssociationOutput) int {
	var total int
	for _, count := range association.AssociationDescription.Overview.AssociationStatusAggregatedCount {
		total += int(count)
	}
	return total
}

func getAssociationPendingTargets(association *ssm.DescribeAssociationOutput) int {
	return int(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Pending"])
}

func getAssociationSuccessTargets(association *ssm.DescribeAssociationOutput) int {
	return int(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Success"])
}

func getAssociationFailedTargets(association *ssm.DescribeAssociationOutput) int {
	return int(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Failed"])
}