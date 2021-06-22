package data

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type AssociationList struct {
	*widgets.List
}

func NewAssociationList() (*AssociationList, error) {
	al := &AssociationList{widgets.NewList()}
	al.SelectedRowStyle = ui.NewStyle(ui.ColorCyan)
	al.Title = "State Associations"
	al.WrapText = true
	if err := al.Reload(); err != nil {
		return nil, err
	}
	return al, nil
}

func (al *AssociationList) Reload() error {
	var err error
	al.Rows, err = prepareAssociationList()
	if err != nil {
		return err
	}
	return nil
}
