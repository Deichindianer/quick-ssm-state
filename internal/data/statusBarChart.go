package data

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type StatusBarChart struct {
	*widgets.BarChart
	ssmClient *ssm.Client
}

func NewStatusBarChart(ssmClient *ssm.Client, termWidth int, initialAssociation string) (*StatusBarChart, error) {
	bc := &StatusBarChart{widgets.NewBarChart(), ssmClient}
	bc.BarColors = []ui.Color{ui.ColorGreen, ui.ColorRed, ui.ColorYellow, ui.ColorCyan}
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorBlack)}
	bc.PaddingRight = 1
	bc.PaddingLeft = 1
	bc.BarWidth = (termWidth - 10) / 2 / 4
	bc.Title = "Target states of the association:"
	bc.Labels = []string{"Success", "Failed", "Pending", "Skipped"}
	bc.Title = fmt.Sprintf("Target states of the association: %s", initialAssociation)
	if err := bc.Reload(initialAssociation); err != nil {
		return nil, err
	}
	return bc, nil
}

func (bc *StatusBarChart) Reload(associationString string) error {
	bc.Title = fmt.Sprintf("Target states of the association: %s", associationString)
	associationID := strings.Split(associationString, " ")[0]
	a, err := getAssociation(bc.ssmClient, associationID)
	if err != nil {
		return err
	}
	s := float64(a.AssociationDescription.Overview.AssociationStatusAggregatedCount["Success"])
	f := float64(a.AssociationDescription.Overview.AssociationStatusAggregatedCount["Failed"])
	p := float64(a.AssociationDescription.Overview.AssociationStatusAggregatedCount["Pending"])
	sk := float64(a.AssociationDescription.Overview.AssociationStatusAggregatedCount["Skipped"])
	data := []float64{s, f, p, sk}
	if s+f+p == 0 {
		data = []float64{}
	}
	bc.Data = data
	return nil
}
