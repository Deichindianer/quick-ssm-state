package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type AssociationList struct {
	*widgets.List
	ssmClient *ssm.Client
}

func NewAssociationList(ssmClient *ssm.Client) (*AssociationList, error) {
	al := &AssociationList{widgets.NewList(), ssmClient}
	al.SelectedRowStyle = ui.NewStyle(ui.ColorCyan)
	al.Title = "State Associations"
	al.WrapText = true
	if err := al.Reload(); err != nil {
		return nil, err
	}
	return al, nil
}

func (al *AssociationList) Reload() error {
	associations, err := al.ssmClient.ListAssociations(context.Background(), &ssm.ListAssociationsInput{})
	if err != nil {
		return err
	}
	al.Rows, err = prepareAssociationList(associations)
	if err != nil {
		return err
	}
	return nil
}
