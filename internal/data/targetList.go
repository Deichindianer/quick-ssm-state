package data

import "github.com/gizak/termui/v3/widgets"

type TargetList struct {
	*widgets.List
}

func NewTargetList(initialAssociation string) (*TargetList, error) {
	tl := &TargetList{widgets.NewList()}
	tl.Title = "Target Selector"
	if err := tl.Reload(initialAssociation); err != nil {
		return nil, err
	}
	return tl, nil
}

func (tl *TargetList) Reload(association string) error {
	targets, err := getAssociationTargets(association)
	if err != nil {
		return err
	}
	tl.Rows = targets
	return nil
}
