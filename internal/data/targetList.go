package data

import (
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/gizak/termui/v3/widgets"
)

type TargetList struct {
	*widgets.List
	ssmClient *ssm.Client
}

func NewTargetList(ssmClient *ssm.Client, initialAssociation string) (*TargetList, error) {
	tl := &TargetList{widgets.NewList(), ssmClient}
	tl.Title = "Target Selector"
	if err := tl.Reload(initialAssociation); err != nil {
		return nil, err
	}
	return tl, nil
}

func (tl *TargetList) Reload(association string) error {
	targets, err := getAssociationTargets(tl.ssmClient, association)
	if err != nil {
		return err
	}
	tl.Rows = targets
	return nil
}
