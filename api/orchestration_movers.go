package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/aws-go/services/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	log "github.com/sirupsen/logrus"
)

// datamoverDescribe gets details about a specific data mover (task and locations)
func (o *datasyncOrchestrator) datamoverDescribe(ctx context.Context, group, id string) (*DatamoverResponse, error) {
	// construct the task ARN from the id
	arn := fmt.Sprintf("arn:aws:datasync:%s:%s:task/%s", *o.server.session.Session.Config.Region, o.account, id)

	t, err := o.datasyncClient.GetDatasyncTags(ctx, arn)
	if err != nil {
		return nil, err
	}
	tags := fromDatasyncTags(t)

	if !tags.inOrg(o.server.org) {
		return nil, apierror.New(apierror.ErrNotFound, "datasync mover not found in org", nil)
	}

	if !tags.inGroup(group) {
		return nil, apierror.New(apierror.ErrNotFound, "datasync mover not found in group", nil)
	}

	task, err := o.datasyncClient.DescribeDatasyncTask(ctx, arn)
	if err != nil {
		return nil, err
	}

	// get a list of all locations with their type
	// there's no way to determine the type of a specific location :(
	locations, err := o.datasyncClient.ListDatasyncLocations(ctx)
	if err != nil {
		return nil, err
	}

	srcLocationType, ok := locations[*task.SourceLocationArn]
	if !ok {
		log.Warn("unable to determine source location type")
	}

	dstLocationType, ok := locations[*task.DestinationLocationArn]
	if !ok {
		log.Warn("unable to determine destination location type")
	}

	srcLocation, err := o.describeDatasyncLocation(ctx, srcLocationType, *task.SourceLocationArn)
	if err != nil {
		return nil, err
	}

	dstLocation, err := o.describeDatasyncLocation(ctx, dstLocationType, *task.DestinationLocationArn)
	if err != nil {
		return nil, err
	}

	return &DatamoverResponse{
		Task:        task,
		Source:      srcLocation,
		Destination: dstLocation,
		Tags:        tags,
	}, nil
}

// datamoverList lists all data movers (tasks) in a group
func (o *datasyncOrchestrator) datamoverList(ctx context.Context, group string) ([]string, error) {
	filters := []*resourcegroupstaggingapi.TagFilter{
		{
			Key:   "spinup:org",
			Value: []string{o.server.org},
		},
		{
			Key:   "spinup:type",
			Value: []string{"storage"},
		},
		{
			Key:   "spinup:flavor",
			Value: []string{"datamover"},
		},
	}

	if group == "" {
		log.Debug("listing all data movers")
	} else {
		log.Debugf("listing data movers in group %s", group)

		filters = append(filters, &resourcegroupstaggingapi.TagFilter{
			Key:   "spinup:spaceid",
			Value: []string{group},
		})
	}

	out, err := o.rgClient.GetResourcesWithTags(ctx, []string{"datasync"}, filters)
	if err != nil {
		return nil, err
	}

	resources := make([]string, 0, len(out))
	for _, r := range out {
		a, err := arn.Parse(aws.StringValue(r.ResourceARN))
		if err != nil {
			return nil, apierror.New(apierror.ErrInternalError, "failed to parse ARN "+aws.StringValue(r.ResourceARN), err)
		}

		parts := strings.SplitN(a.Resource, "/", 2)
		if len(parts) > 1 {
			resources = append(resources, parts[1])
		}
	}

	return resources, nil
}

// describeDatasyncLocation returns information for the spcific location type
func (o *datasyncOrchestrator) describeDatasyncLocation(ctx context.Context, lType, lArn string) (*DatamoverLocation, error) {
	if lType == "" || lArn == "" {
		return nil, nil
	}

	log.Debugf("location %s is type %s", lArn, lType)

	switch lType {
	case "s3":
		dstLocationS3, err := o.datasyncClient.DescribeDatasyncLocationS3(ctx, lArn)
		if err != nil {
			return nil, err
		}
		return &DatamoverLocation{S3: dstLocationS3}, nil
	case "efs":
		dstLocationEfs, err := o.datasyncClient.DescribeDatasyncLocationEfs(ctx, lArn)
		if err != nil {
			return nil, err
		}
		return &DatamoverLocation{EFS: dstLocationEfs}, nil
	case "smb":
		dstLocationSmb, err := o.datasyncClient.DescribeDatasyncLocationSmb(ctx, lArn)
		if err != nil {
			return nil, err
		}
		return &DatamoverLocation{SMB: dstLocationSmb}, nil
	case "nfs":
		dstLocationNfs, err := o.datasyncClient.DescribeDatasyncLocationNfs(ctx, lArn)
		if err != nil {
			return nil, err
		}
		return &DatamoverLocation{NFS: dstLocationNfs}, nil
	default:
		log.Warnf("type %s didn't match any supported location types", lType)
		return nil, apierror.New(apierror.ErrInternalError, "unknown datasync location type "+lType, nil)
	}
}
