package api

import "github.com/aws/aws-sdk-go/service/datasync"

// DatamoverCreateRequest is data used to create a DataSync mover
type DatamoverCreateRequest struct {
	Name        *string
	Source      *DatamoverLocationInput
	Destination *DatamoverLocationInput
	Tags        Tags
}

// DatamoverLocationInput is an abstraction for the different location type inputs
// currently only S3 is supported
type DatamoverLocationInput struct {
	Type *LocationType
	S3   *DescribeLocationS3Input
}

type LocationType string

const (
	S3  LocationType = "S3"
	EFS LocationType = "EFS"
	SMB LocationType = "SMB"
	NFS LocationType = "NFS"
)

func (lt LocationType) String() string {
	switch lt {
	case S3:
		return "S3"
	case EFS:
		return "EFS"
	case SMB:
		return "SMB"
	case NFS:
		return "NFS"
	}
	return ""
}

type DescribeLocationS3Input struct {
	S3BucketArn *string
	// S3StorageClass is one of the following:
	// OUTPOSTS, ONEZONE_IA, DEEP_ARCHIVE, GLACIER, INTELLIGENT_TIERING, STANDARD_IA, STANDARD
	S3StorageClass *string
	Subdirectory   *string
}

// DatamoverResponse is the output from DataSync mover operations
type DatamoverResponse struct {
	// https://docs.aws.amazon.com/sdk-for-go/api/service/datasync/#DescribeTaskOutput
	Task        *datasync.DescribeTaskOutput
	Source      *DatamoverLocationOutput
	Destination *DatamoverLocationOutput
	Tags        Tags `json:",omitempty"`
}

// DatamoverLocationOutput is an abstraction for the different location type outputs
type DatamoverLocationOutput struct {
	S3  *datasync.DescribeLocationS3Output  `json:",omitempty"`
	EFS *datasync.DescribeLocationEfsOutput `json:",omitempty"`
	SMB *datasync.DescribeLocationSmbOutput `json:",omitempty"`
	NFS *datasync.DescribeLocationNfsOutput `json:",omitempty"`
}
