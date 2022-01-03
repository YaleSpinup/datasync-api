package api

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
)

type Tag struct {
	Key   string
	Value string
}

type Tags []Tag

// inGroup returns true if there is a spinup:spaceid tag that matches our group
func (tags *Tags) inGroup(group string) bool {
	for _, t := range *tags {
		if t.Key == "spinup:spaceid" && t.Value == group {
			return true
		}
	}
	return false
}

// inOrg returns true if there is a spinup:org tag that matches our org
func (tags *Tags) inOrg(org string) bool {
	for _, t := range *tags {
		if t.Key == "spinup:org" && t.Value == org {
			return true
		}
	}
	return false
}

// normalize sets required tags
func (tags *Tags) normalize(org, group string) Tags {
	normalizedTags := Tags{
		{
			Key:   "spinup:org",
			Value: org,
		},
		{
			Key:   "spinup:spaceid",
			Value: group,
		},
		{
			Key:   "spinup:type",
			Value: "storage",
		},
		{
			Key:   "spinup:flavor",
			Value: "datamover",
		},
	}

	for _, t := range *tags {
		if strings.HasPrefix(t.Key, "aws:") {
			continue
		}

		if t.Key == "spinup:org" || t.Key == "spinup:spaceid" || t.Key == "spinup:type" || t.Key == "spinup:flavor" {
			continue
		}

		normalizedTags = append(normalizedTags, t)
	}

	return normalizedTags
}

// toDatasyncTags converts from api Tags to DataSync tags
func (tags *Tags) toDatasyncTags() []*datasync.TagListEntry {
	datasyncTags := make([]*datasync.TagListEntry, 0, len(*tags))
	for _, t := range *tags {
		datasyncTags = append(datasyncTags, &datasync.TagListEntry{
			Key:   aws.String(t.Key),
			Value: aws.String(t.Value),
		})
	}
	return datasyncTags
}

// fromDatasyncTags converts from DocDB tags to api Tags
func fromDatasyncTags(datasyncTags []*datasync.TagListEntry) Tags {
	tags := make(Tags, 0, len(datasyncTags))
	for _, t := range datasyncTags {
		tags = append(tags, Tag{
			Key:   aws.StringValue(t.Key),
			Value: aws.StringValue(t.Value),
		})
	}
	return tags
}

// fromResourcegroupstaggingapiTags converts from Resourcegroupstaggingapi tags to api Tags
func fromResourcegroupstaggingapiTags(resourcegroupstaggingapiTags []*resourcegroupstaggingapi.Tag) Tags {
	tags := make(Tags, 0, len(resourcegroupstaggingapiTags))
	for _, t := range resourcegroupstaggingapiTags {
		tags = append(tags, Tag{
			Key:   aws.StringValue(t.Key),
			Value: aws.StringValue(t.Value),
		})
	}
	return tags
}
