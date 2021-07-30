package data

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/gizak/termui/v3/widgets"
)

type TargetList struct {
	*widgets.List
	ssmClient *ssm.Client
}

func NewTargetList(ssmClient *ssm.Client, initialAssociation string) (*TargetList, error) {
	tl := &TargetList{widgets.NewList(), ssmClient}
	tl.Title = "Targets"
	if err := tl.Reload(initialAssociation); err != nil {
		return nil, err
	}
	return tl, nil
}

func (tl *TargetList) Reload(association string) error {
	associationID := strings.Split(association, " ")[0]
	executions, err := tl.ssmClient.DescribeAssociationExecutions(context.Background(), &ssm.DescribeAssociationExecutionsInput{
		AssociationId: aws.String(associationID),
		MaxResults:    1,
	})
	if err != nil {
		tl.Rows = []string{"Failed to get executions."}
		log.Fatal(err)
		return nil
	}
	if len(executions.AssociationExecutions) != 1 {
		tl.Rows = nil
		return nil
	}
	executionTargets, err := tl.ssmClient.DescribeAssociationExecutionTargets(context.Background(), &ssm.DescribeAssociationExecutionTargetsInput{
		AssociationId: aws.String(associationID),
		ExecutionId:   executions.AssociationExecutions[0].ExecutionId,
	})
	if err != nil {
		return err
	}
	tl.Rows = nil
	for _, executionTarget := range executionTargets.AssociationExecutionTargets {
		row := fmt.Sprintf("%s: %s: %s", *executionTarget.ResourceType, *executionTarget.ResourceId, *executionTarget.Status)
		tl.Rows = append(tl.Rows, row)
	}

	return nil
}
