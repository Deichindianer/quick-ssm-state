package data

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

func getAssociation(ssmClient *ssm.Client, associationID string) (*ssm.DescribeAssociationOutput, error) {
	association, err := ssmClient.DescribeAssociation(context.Background(), &ssm.DescribeAssociationInput{
		AssociationId: aws.String(associationID),
	})
	if err != nil {
		return nil, err
	}
	return association, nil
}

func prepareAssociationList(associations *ssm.ListAssociationsOutput) ([]string, error) {
	if associations == nil {
		return nil, fmt.Errorf("associations list is empty")
	}

	var associationNames []string
	for _, a := range associations.Associations {
		if a.AssociationName == nil {
			a.AssociationName = aws.String("None")
		}
		if a.AssociationId == nil {
			return nil, errors.New("AssociationID is nil, wtf man")
		}
		associationNames = append(associationNames, fmt.Sprintf("%s %s", *a.AssociationId, *a.AssociationName))
	}
	return associationNames, nil
}

func GetTargetOutput(ssmClient *ssm.Client, executionTarget types.AssociationExecutionTarget) (string, error) {
	if executionTarget.OutputSource.OutputSourceType == nil {
		return "", nil
	}
	switch *executionTarget.OutputSource.OutputSourceType {
	case "Amazon S3":
		return *executionTarget.OutputSource.OutputSourceId, nil
	case "RunCommand":
		command, err := ssmClient.GetCommandInvocation(context.Background(), &ssm.GetCommandInvocationInput{
			CommandId:  executionTarget.OutputSource.OutputSourceId,
			InstanceId: executionTarget.ResourceId,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get command output: %s", err)
		}
		if command.StandardOutputContent == nil {
			return "", nil
		}
		return *command.StandardOutputContent, nil
	default:
		return "", fmt.Errorf("unrecognised output executionTarget: %s; id: %s", *executionTarget.OutputSource.OutputSourceType, *executionTarget.OutputSource.OutputSourceId)
	}
}

func getExecutionTargetsFromExecution(o *OutputParagraph, associationID string) (*ssm.DescribeAssociationExecutionTargetsOutput, error) {
	executions, err := o.ssmClient.DescribeAssociationExecutions(context.Background(), &ssm.DescribeAssociationExecutionsInput{
		AssociationId: aws.String(associationID),
		MaxResults:    1,
	})
	executionTargets, err := o.ssmClient.DescribeAssociationExecutionTargets(context.Background(), &ssm.DescribeAssociationExecutionTargetsInput{
		AssociationId: aws.String(associationID),
		ExecutionId:   executions.AssociationExecutions[0].ExecutionId,
	})
	if err != nil {
		return nil, err
	}
	return executionTargets, nil
}
