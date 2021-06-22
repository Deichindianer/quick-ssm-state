package data

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func getAssociationTargets(associationString string) ([]string, error) {
	associationID := strings.Split(associationString, " ")[0]
	a, err := getAssociation(associationID)
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

func listAssociations() (*ssm.ListAssociationsOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	awsClient := ssm.NewFromConfig(cfg)
	associations, err := awsClient.ListAssociations(context.Background(), &ssm.ListAssociationsInput{})
	if err != nil {
		return nil, err
	}
	return associations, nil
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

func getAssociation(associationID string) (*ssm.DescribeAssociationOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	awsClient := ssm.NewFromConfig(cfg)
	association, err := awsClient.DescribeAssociation(context.Background(), &ssm.DescribeAssociationInput{
		AssociationId: aws.String(associationID),
	})
	if err != nil {
		return nil, err
	}
	return association, nil
}

func getAssociationPendingTargets(association *ssm.DescribeAssociationOutput) float64 {
	return float64(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Pending"])
}

func getAssociationSuccessTargets(association *ssm.DescribeAssociationOutput) float64 {
	return float64(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Success"])
}

func getAssociationFailedTargets(association *ssm.DescribeAssociationOutput) float64 {
	return float64(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Failed"])
}

func getAssociationSkippedTargets(association *ssm.DescribeAssociationOutput) float64 {
	return float64(association.AssociationDescription.Overview.AssociationStatusAggregatedCount["Skipped"])
}
