package data

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/gizak/termui/v3/widgets"
)

type OutputParagraph struct {
	*widgets.Paragraph
	ssmClient *ssm.Client
}

func NewOutputParagraph(ssmClient *ssm.Client, initialAssociation string) (*OutputParagraph, error) {
	o := &OutputParagraph{
		Paragraph: widgets.NewParagraph(),
		ssmClient:    ssmClient,
	}
	if err := o.Reload(initialAssociation); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *OutputParagraph) Reload(association string) error {
	associationID := strings.Split(association, " ")[0]
	executionTargets, err := getExecutionTargetsFromExecution(o, associationID)
	if err != nil {
		return err
	}
	output, err := GetTargetOutput(o.ssmClient, executionTargets.AssociationExecutionTargets[0])
	if err != nil {
		return err
	}
	o.Paragraph.Text = output
	return nil
}
