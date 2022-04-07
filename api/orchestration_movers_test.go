package api

import (
	"context"
	"testing"
	"time"

	"github.com/YaleSpinup/apierror"
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

type mockDataSync struct {
	datasynciface.DataSyncAPI
	t   *testing.T
	err error
}

type mockRGClient struct {
	resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI
	t   *testing.T
	err error
}

// We dont want to start task so we mocked it
func (d *mockDataSync) StartTaskExecutionWithContext(ctx context.Context, input *datasync.StartTaskExecutionInput, opts ...request.Option) (*datasync.StartTaskExecutionOutput, error) {
	if d.err != nil {
		return nil, d.err
	}
	out := &datasync.StartTaskExecutionOutput{
		TaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
	}

	return out, nil
}

// We dont want to start task so we mocked it
func (d *mockDataSync) CancelTaskExecutionWithContext(ctx context.Context, input *datasync.CancelTaskExecutionInput, opts ...request.Option) (*datasync.CancelTaskExecutionOutput, error) {
	if d.err != nil {
		return nil, d.err
	}
	out := &datasync.CancelTaskExecutionOutput{}

	return out, nil
}

var is_running = false

func (d *mockDataSync) ListTaskExecutionsPagesWithContext(ctx context.Context, input *datasync.ListTaskExecutionsInput, callback func(*datasync.ListTaskExecutionsOutput, bool) bool, opts ...request.Option) error {

	out := &datasync.ListTaskExecutionsOutput{
		TaskExecutions: []*datasync.TaskExecutionListEntry{
			{TaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581")},
			{TaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3582")},
			{TaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3583")},
			{TaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3584")},
			{TaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3585")},
		},
	}

	callback(out, false)
	return nil

}

var testTime = time.Date(2020, 10, 10, 10, 10, 10, 0, time.UTC)

func (d *mockDataSync) DescribeTaskExecutionWithContext(ctx context.Context, input *datasync.DescribeTaskExecutionInput, opts ...request.Option) (*datasync.DescribeTaskExecutionOutput, error) {

	var ret = &datasync.DescribeTaskExecutionOutput{
		BytesTransferred:         aws.Int64(100),
		BytesWritten:             aws.Int64(100),
		EstimatedBytesToTransfer: aws.Int64(100),
		EstimatedFilesToTransfer: aws.Int64(100),
		Excludes:                 []*datasync.FilterRule{},
		FilesTransferred:         aws.Int64(100),
		Includes:                 []*datasync.FilterRule{},
		Options:                  &datasync.Options{},
		Result:                   &datasync.TaskExecutionResultDetail{},
		StartTime:                aws.Time(testTime),
		Status:                   aws.String("RUNNING"),
		TaskExecutionArn:         aws.String("arn:aws:datasync:us-east-1:012345678901:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3585"),
	}
	return ret, nil

}

// It returns a task with RUNNING / AVAIABLE bases on is_running
func (d *mockDataSync) DescribeTaskWithContext(ctx context.Context, input *datasync.DescribeTaskInput, opts ...request.Option) (*datasync.DescribeTaskOutput, error) {
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
func (r *mockRGClient) GetResourcesWithContext(ctx context.Context, input *resourcegroupstaggingapi.GetResourcesInput, opts ...request.Option) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
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

func newMockDataSync(t *testing.T, err error) datasynciface.DataSyncAPI {
	return &mockDataSync{
		t:   t,
		err: err,
	}
}

func newMockRGClient(t *testing.T, err error) resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI {
	return &mockRGClient{
		t:   t,
		err: err,
	}
}

func newMockDataSyncOrchestrator(t *testing.T) *datasyncOrchestrator {
	return &datasyncOrchestrator{
		server: &server{},
		sp:     &sessionParams{},
		datasyncClient: ydatasync.Datasync{
			Service: newMockDataSync(t, nil),
		},
		rgClient: yresourcegroupstaggingapi.ResourceGroupsTaggingAPI{
			Service: newMockRGClient(t, nil),
		},
	}
}

func TestStartTaskRun(t *testing.T) {
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

	o := newMockDataSyncOrchestrator(t)
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

	o := newMockDataSyncOrchestrator(t)

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

func Test_datamoverRunDescribe(t *testing.T) {

	run := &DatamoverRun{
		BytesTransferred:         aws.Int64(100),
		BytesWritten:             aws.Int64(100),
		EstimatedBytesToTransfer: aws.Int64(100),
		EstimatedFilesToTransfer: aws.Int64(100),
		FilesTransferred:         aws.Int64(100),
		Status:                   aws.String("RUNNING"),
		Result:                   &datasync.TaskExecutionResultDetail{},
		StartTime:                aws.Time(testTime),
	}

	type input struct {
		group string
		name  string
		id    string
	}
	type output struct {
		isErr bool
		run   *DatamoverRun
	}
	tests := []struct {
		name  string
		input input
		exp   output
	}{
		{"group empty", input{"", "name1", "1234"}, output{true, nil}},
		{"name empty", input{"group1", "", "1234"}, output{true, nil}},
		{"id empty", input{"group1", "name1", ""}, output{true, nil}},
		{"valid test", input{"group1", "name1", "1234"}, output{false, run}},
	}

	o := newMockDataSyncOrchestrator(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := o.datamoverRunDescribe(ctx, tt.input.group, tt.input.name, tt.input.id)
			if tt.exp.isErr && err == nil {
				t.Error("expected error but did not receive error")
			} else if !tt.exp.isErr && err != nil {
				t.Errorf("received unexpected error %v", err)
			}
			assert.Equal(t, tt.exp.run, got)
		})
	}
}

func Test_datamoverRunList(t *testing.T) {
	validCaseRes := []string{"exec-086d6c629a6bf3581", "exec-086d6c629a6bf3582", "exec-086d6c629a6bf3583", "exec-086d6c629a6bf3584", "exec-086d6c629a6bf3585"}
	type input struct {
		group string
		name  string
	}
	type output struct {
		isErr bool
		res   []string
	}
	tests := []struct {
		name  string
		input input
		exp   output
	}{
		{"group empty", input{"", "name1"}, output{true, nil}},
		{"name empty", input{"group1", ""}, output{true, nil}},
		{"valid test", input{"group1", "name1"}, output{false, validCaseRes}},
	}

	o := newMockDataSyncOrchestrator(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := o.datamoverRunList(ctx, tt.input.group, tt.input.name)
			if tt.exp.isErr && err == nil {
				t.Error("expected error but did not receive error")
			} else if !tt.exp.isErr && err != nil {
				t.Errorf("received unexpected error %v", err)
			}
			assert.Equal(t, tt.exp.res, got)
		})
	}
}

func Test_datamoverList(t *testing.T) {
	type input struct {
		group string
	}
	type output struct {
		isErr bool
		res   []string
	}
	tests := []struct {
		name  string
		input input
		exp   output
	}{
		{"empty group", input{""}, output{false, []string{"name1"}}},
		{"non empty group", input{"group1"}, output{false, []string{"name1"}}},
	}

	o := newMockDataSyncOrchestrator(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := o.datamoverList(ctx, tt.input.group)
			if tt.exp.isErr && err == nil {
				t.Error("expected error but did not receive error")
			} else if !tt.exp.isErr && err != nil {
				t.Errorf("received unexpected error %v", err)
			}
			assert.Equal(t, tt.exp.res, got)
		})
	}
}
