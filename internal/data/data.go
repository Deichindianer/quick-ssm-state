package data

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func getAssociationTargets(ssmClient *ssm.Client, associationString string) ([]string, error) {
	associationID := strings.Split(associationString, " ")[0]
	a, err := getAssociation(ssmClient, associationID)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, t := range a.AssociationDescription.Targets {
		vals := strings.Join(t.Values, ", ")
		result = append(result, fmt.Sprintf("%s:%s", *t.Key, vals))
	}
	return result, nil
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