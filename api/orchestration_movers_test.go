package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/YaleSpinup/apierror"
	yiam "github.com/YaleSpinup/aws-go/services/iam"
	yresourcegroupstaggingapi "github.com/YaleSpinup/aws-go/services/resourcegroupstaggingapi"
	ydatasync "github.com/YaleSpinup/datasync-api/datasync"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/datasync/datasynciface"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
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

// We dont want to start task so we mocked it
func (d *mockDatasync) CancelTaskExecutionWithContext(ctx context.Context, input *datasync.CancelTaskExecutionInput, opts ...request.Option) (*datasync.CancelTaskExecutionOutput, error) {
	if d.err != nil {
		return nil, d.err
	}
	out := &datasync.CancelTaskExecutionOutput{}

	return out, nil
}

var is_running = false

// It returns a task with RUNNING / AVAIABLE bases on is_running
func (d *mockDatasync) DescribeTaskWithContext(ctx context.Context, input *datasync.DescribeTaskInput, opts ...request.Option) (*datasync.DescribeTaskOutput, error) {
	if is_running {

		return &datasync.DescribeTaskOutput{
			Name:                    aws.String("name1"),
			Status:                  aws.String("RUNNING"),
			TaskArn:                 aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
			CurrentTaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
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

func Test_startTaskRun(t *testing.T) {
	type TaskRunRes struct {
		res string
		err error
	}

	cases := []struct {
		isNegative bool
		ctx        context.Context
		group      string
		name       string
		isRunning  bool
		expected   TaskRunRes
		message    string
	}{
		{false, nil, "Group1", "name1", false,
			TaskRunRes{"exec-086d6c629a6bf3581", nil}, "StartTask Positive."},
		{true, nil, "Group1", "name1", true,
			TaskRunRes{"", apierror.New(apierror.ErrConflict, "", nil)}, "StartTask Negative"},
		{true, nil, "", "", true,
			TaskRunRes{"", apierror.New(apierror.ErrConflict, "", nil)}, "StartTask Negative Without Name and group"},
	}

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

	for _, test := range cases {
		t.Log(test.message)
		is_running = test.isRunning
		resp, err := o.startTaskRun(test.ctx, test.group, test.name)
		if test.isNegative {
			if err == nil {
				t.Error("expected error , got no error")
			}

		} else {
			if err != nil {
				t.Errorf("expected err nil, got: %v", err)
			}
		}
		if resp != test.expected.res {
			t.Errorf("expected resp %v, got: %v", test.expected.res, resp)
		}
	}
}

func Test_stopTaskRun(t *testing.T) {

	cases := []struct {
		isNegative bool
		ctx        context.Context
		group      string
		name       string
		isRunning  bool
		expected   error
		message    string
	}{
		{false, nil, "Group1", "name1", true, nil, "StopTask Positive."},
		{true, nil, "Group1", "name1", false, apierror.New(apierror.ErrConflict, "", nil), "StopTask Negative"},
		{true, nil, "", "", true, apierror.New(apierror.ErrConflict, "", nil), "StopTask Negative Without Name and group"},
	}

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

	for _, test := range cases {
		t.Log(test.message)
		is_running = test.isRunning
		err := o.stopTaskRun(test.ctx, test.group, test.name)
		if test.isNegative {
			if err == nil {
				t.Error("expected error , got no error")
			}

		} else {
			if err != nil {
				t.Errorf("expected err nil, got: %v", err)
			}
		}
	}
}
