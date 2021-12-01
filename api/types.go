package api

import "github.com/aws/aws-sdk-go/service/datasync"

// DatamoverResponse is the output from DataSync mover operations
type DatamoverResponse struct {
	// https://docs.aws.amazon.com/sdk-for-go/api/service/datasync/#DescribeTaskOutput
	Task        *datasync.DescribeTaskOutput
	Source      *DatamoverLocation
	Destination *DatamoverLocation
	Tags        Tags `json:",omitempty"`
}

// DatamoverLocation is an abstraction for the different location types
type DatamoverLocation struct {
	S3  *datasync.DescribeLocationS3Output  `json:",omitempty"`
	EFS *datasync.DescribeLocationEfsOutput `json:",omitempty"`
	SMB *datasync.DescribeLocationSmbOutput `json:",omitempty"`
	NFS *datasync.DescribeLocationNfsOutput `json:",omitempty"`
}
