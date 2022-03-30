package api

import (
	"context"
	"fmt"
	"testing"

	yiam "github.com/YaleSpinup/aws-go/services/iam"
	yresourcegroupstaggingapi "github.com/YaleSpinup/aws-go/services/resourcegroupstaggingapi"
	ydatasync "github.com/YaleSpinup/datasync-api/datasync"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/datasync/datasynciface"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"github.com/stretchr/testify/assert"
)

type mockDatasync struct {
	datasynciface.DataSyncAPI
	t   *testing.T
	err error
}

type mockrgClient struct {
	resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI
	t   *testing.T
	err error
}

// We dont want to start task so we mocked it
func (d *mockDatasync) StartTaskExecutionWithContext(ctx context.Context, input *datasync.StartTaskExecutionInput, opts ...request.Option) (*datasync.StartTaskExecutionOutput, error) {
	if d.err != nil {
		return nil, d.err
	}
	out := &datasync.StartTaskExecutionOutput{
		TaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
	}

	return out, nil
}

var is_running = false

// It returns a task with RUNNING / AVAIABLE bases on is_running
func (d *mockDatasync) DescribeTaskWithContext(ctx context.Context, input *datasync.DescribeTaskInput, opts ...request.Option) (*datasync.DescribeTaskOutput, error) {
	if is_running {

		return &datasync.DescribeTaskOutput{
			Name:    aws.String("name1"),
			Status:  aws.String("RUNNING"),
			TaskArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
		}, nil
	}
	return &datasync.DescribeTaskOutput{
		Name:    aws.String("name1"),
		Status:  aws.String("AVAILABLE"),
		TaskArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
	}, nil
}

//Get tags
func (r *mockrgClient) GetResourcesWithContext(ctx context.Context, input *resourcegroupstaggingapi.GetResourcesInput, opts ...request.Option) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	fmt.Print(ctx, input, opts)
	return &resourcegroupstaggingapi.GetResourcesOutput{
		PaginationToken: new(string),
		ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{
			{ResourceARN: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
				Tags: []*resourcegroupstaggingapi.Tag{
					{Key: aws.String("mockkey"),
						Value: aws.String("mockvalue")},
				}},
		},
	}, nil
}

func newmockDatasync(t *testing.T, err error) datasynciface.DataSyncAPI {
	return &mockDatasync{
		t:   t,
		err: err,
	}
}

func newmockrgClient(t *testing.T, err error) resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI {
	return &mockrgClient{
		t:   t,
		err: err,
	}
}

func Test_StartRunRunsWhenStatusAvailable(t *testing.T) {
	is_running = false
	o := &datasyncOrchestrator{
		account: "",
		server:  &server{},
		sp:      &sessionParams{},
		datasyncClient: ydatasync.Datasync{
			Service:         newmockDatasync(t, nil),
			DefaultKMSKeyId: "",
		},
		iamClient: yiam.IAM{},
		rgClient: yresourcegroupstaggingapi.ResourceGroupsTaggingAPI{
			Service: newmockrgClient(t, nil),
		},
	}
	resp, err := o.startTaskRun(nil, "group1", "name1")
	assert.NoError(t, err, "no error")
	assert.Equal(t, "exec-086d6c629a6bf3581", resp)
}

func Test_StartRunDoseNotRunIfitsrunning(t *testing.T) {
	is_running = true
	o := &datasyncOrchestrator{
		account: "",
		server:  &server{},
		sp:      &sessionParams{},
		datasyncClient: ydatasync.Datasync{
			Service:         newmockDatasync(t, nil),
			DefaultKMSKeyId: "",
		},
		iamClient: yiam.IAM{},
		rgClient: yresourcegroupstaggingapi.ResourceGroupsTaggingAPI{
			Service: newmockrgClient(t, nil),
		},
	}
	resp, err := o.startTaskRun(nil, "group1", "name1")
	assert.Error(t, err, "Conflict: another datasync mover task is already running")
	assert.Equal(t, "", resp)
}
func Test_StartRunFailsWithoutgroupAndName(t *testing.T) {
	is_running = false
	o := &datasyncOrchestrator{
		account: "",
		server:  &server{},
		sp:      &sessionParams{},
		datasyncClient: ydatasync.Datasync{
			Service:         newmockDatasync(t, nil),
			DefaultKMSKeyId: "",
		},
		iamClient: yiam.IAM{},
		rgClient: yresourcegroupstaggingapi.ResourceGroupsTaggingAPI{
			Service: newmockrgClient(t, nil),
		},
	}
	resp, err := o.startTaskRun(nil, "", "")
	assert.Error(t, err, "BadRequest: invalid input")
	assert.Equal(t, "", resp)
}
