package resourcegroupstaggingapi

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	log "github.com/sirupsen/logrus"
)

// ResourceGroupsTaggingAPI is a wrapper around the aws resourcegroupstaggingapi service with some default config info
type ResourceGroupsTaggingAPI struct {
	session *session.Session
	Service resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI
}

type ResourceGroupsTaggingAPIOption func(*ResourceGroupsTaggingAPI)

// Tag Filter is used to filter resources based on tags.  The Value portion is optional.
type TagFilter struct {
	Key   string
	Value []string
}

func New(opts ...ResourceGroupsTaggingAPIOption) *ResourceGroupsTaggingAPI {
	client := ResourceGroupsTaggingAPI{}

	for _, opt := range opts {
		opt(&client)
	}

	if client.session != nil {
		client.Service = resourcegroupstaggingapi.New(client.session)
	}

	return &client
}

func WithSession(sess *session.Session) ResourceGroupsTaggingAPIOption {
	return func(client *ResourceGroupsTaggingAPI) {
		log.Debug("using aws session")
		client.session = sess
	}
}

func WithCredentials(key, secret, token, region string) ResourceGroupsTaggingAPIOption {
	return func(client *ResourceGroupsTaggingAPI) {
		log.Debugf("creating new session with key id %s in region %s", key, region)
		sess := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(key, secret, token),
			Region:      aws.String(region),
		}))
		client.session = sess
	}
}

func (r *ResourceGroupsTaggingAPI) ListResourcesWithTags(ctx context.Context, input *resourcegroupstaggingapi.GetResourcesInput) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Info("listing tagged resources")

	out, err := r.Service.GetResourcesWithContext(ctx, input)
	if err != nil {
		return nil, ErrCode("listing resource with tags", err)
	}

	log.Debugf("got output from get resources: %+v", out)

	return out, nil
}

// GetResourcesInOrg returns all of the resources in the specified org, with a given resource type and flavor
// More details about which services support the resourgroup tagging api here: https://docs.aws.amazon.com/ARG/latest/userguide/supported-resources.html
func (r *ResourceGroupsTaggingAPI) GetResourcesInOrg(ctx context.Context, org, rtype, rflavor, group string) ([]*resourcegroupstaggingapi.ResourceTagMapping, error) {
	if org == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting resources with type '%s' flavor '%s' in org %s (group '%s')", rtype, rflavor, org, group)

	filters := []*resourcegroupstaggingapi.TagFilter{
		{
			Key:    aws.String("spinup:org"),
			Values: []*string{aws.String(org)},
		},
	}

	if rtype != "" {
		filters = append(filters, &resourcegroupstaggingapi.TagFilter{
			Key:    aws.String("spinup:type"),
			Values: []*string{aws.String(rtype)},
		})
	}

	if rflavor != "" {
		filters = append(filters, &resourcegroupstaggingapi.TagFilter{
			Key:    aws.String("spinup:flavor"),
			Values: []*string{aws.String(rflavor)},
		})
	}

	if group != "" {
		filters = append(filters, &resourcegroupstaggingapi.TagFilter{
			Key:    aws.String("spinup:spaceid"),
			Values: []*string{aws.String(group)},
		})
	}

	out, err := r.Service.GetResourcesWithContext(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		ResourcesPerPage:    aws.Int64(100),
		ResourceTypeFilters: aws.StringSlice([]string{"datasync"}),
		TagFilters:          filters,
	})
	if err != nil {
		return nil, ErrCode("getting resources with tags", err)
	}

	log.Debugf("got output from get resources: %+v", out)

	return out.ResourceTagMappingList, nil
}
