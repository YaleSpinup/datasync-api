package api

import (
	"time"

	"github.com/aws/aws-sdk-go/service/datasync"
)

// DatamoverCreateRequest is data used to create a DataSync mover
type DatamoverCreateRequest struct {
	Name        *string
	Source      *DatamoverLocationInput
	Destination *DatamoverLocationInput
	Tags        Tags
}

// DatamoverLocationInput is an abstraction for the different location type inputs
// currently only S3 and EFS are supported
type DatamoverLocationInput struct {
	Type LocationType
	S3   *DatamoverLocationS3Input
	EFS  *DatamoverLocationEFSInput
}

type DatamoverLocationS3Input struct {
	S3BucketArn *string
	// S3StorageClass is one of the following:
	// OUTPOSTS, ONEZONE_IA, DEEP_ARCHIVE, GLACIER, INTELLIGENT_TIERING, STANDARD_IA, STANDARD
	S3StorageClass *string
	Subdirectory   *string
}

type DatamoverLocationEFSInput struct {
	EfsFilesystemArn  *string
	SecurityGroupArns []*string
	SubnetArn         *string
	Subdirectory      *string
}

type LocationType string

const (
	S3  LocationType = "S3"
	EFS LocationType = "EFS"
	SMB LocationType = "SMB"
	NFS LocationType = "NFS"
)

func (lt LocationType) String() string {
	return string(lt)
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
	Type LocationType
	S3   *datasync.DescribeLocationS3Output  `json:",omitempty"`
	EFS  *datasync.DescribeLocationEfsOutput `json:",omitempty"`
	SMB  *datasync.DescribeLocationSmbOutput `json:",omitempty"`
	NFS  *datasync.DescribeLocationNfsOutput `json:",omitempty"`
}

type DatamoverRun struct {
	BytesTransferred         *int64
	BytesWritten             *int64
	EstimatedBytesToTransfer *int64
	EstimatedFilesToTransfer *int64
	FilesTransferred         *int64
	StartTime                *time.Time
	Status                   *string
	Result                   *datasync.TaskExecutionResultDetail
}

type DatamoverAction struct {
	State *string
}
