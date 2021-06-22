package data

import (
	"fmt"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type StatusBarChart struct {
	*widgets.BarChart
}

func NewStatusBarChart(termWidth int, initialAssociation string) (*StatusBarChart, error) {
	bc := &StatusBarChart{widgets.NewBarChart()}
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
	a, err := getAssociation(associationID)
	if err != nil {
		return err
	}
	s := getAssociationSuccessTargets(a)
	f := getAssociationFailedTargets(a)
	p := getAssociationPendingTargets(a)
	sk := getAssociationSkippedTargets(a)
	data := []float64{s, f, p, sk}
	if s+f+p == 0 {
		data = []float64{}
	}
	bc.Data = data
	return nil
}
