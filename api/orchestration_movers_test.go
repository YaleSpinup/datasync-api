package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/YaleSpinup/aws-go/services/iam"
	yresource "github.com/YaleSpinup/aws-go/services/resourcegroupstaggingapi"
	ydataSync "github.com/YaleSpinup/datasync-api/datasync"
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

func (d *mockDatasync) StartTaskExecutionWithContext(ctx context.Context, input *datasync.StartTaskExecutionInput, opts ...request.Option) (*datasync.StartTaskExecutionOutput, error) {
	fmt.Println("mocked StartTaskExecutionWithContext")
	if d.err != nil {
		return nil, d.err
	}
	fmt.Print(ctx, &input, opts)
	out := &datasync.StartTaskExecutionOutput{
		TaskExecutionArn: aws.String("arn:aws:datasync:us-east-1:516855177326:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
	}

	return out, nil
}

func (d *mockDatasync) DescribeTaskWithContext(ctx context.Context, input *datasync.DescribeTaskInput, opts ...request.Option) (*datasync.DescribeTaskOutput, error) {
	fmt.Println("mocked DescribeTaskWithContext")
	return &datasync.DescribeTaskOutput{
		Name:    aws.String("name1"),
		Status:  aws.String("AVAILABLE"),
		TaskArn: aws.String("arn:aws:datasync:us-east-1:516855177326:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
	}, nil
}

func (r *mockrgClient) GetResourcesWithContext(ctx context.Context, input *resourcegroupstaggingapi.GetResourcesInput, opts ...request.Option) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	fmt.Println("mocked GetResourcesWithContext")
	fmt.Print(ctx, input, opts)
	return &resourcegroupstaggingapi.GetResourcesOutput{
		PaginationToken: new(string),
		ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{
			{ResourceARN: aws.String("arn:aws:datasync:us-east-1:516855177326:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"),
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

func Test_StartRun(T *testing.T) {

	o := &datasyncOrchestrator{
		account: "",
		server:  &server{},
		sp:      &sessionParams{},
		datasyncClient: ydataSync.Datasync{
			Service:         newmockDatasync(T, nil),
			DefaultKMSKeyId: "",
		},
		iamClient: iam.IAM{},
		rgClient: yresource.ResourceGroupsTaggingAPI{
			Service: newmockrgClient(T, nil),
		},
	}
	resp, err := o.startTaskRun(nil, "group1", "name1")
	assert.NoError(T, err, "no error")
	assert.Equal(T, "exec-086d6c629a6bf3581", resp)
}

// func Tesst_startTaskRun(T *testing.T) {
// 	type TaskRunRes struct {
// 		res string
// 		err error
// 	}

// 	cases := []struct {
// 		ctx      context.Context
// 		group    string
// 		name     string
// 		expected TaskRunRes
// 		message  string
// 	}{
// 		{nil, "Gruup1", "AvialName",
// 			TaskRunRes{"arn:aws:datasync:us-east-1:516855177326:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581", nil}, "Should not return full arn"},
// 		//	{nil, "Gruup1", "RunningName", TaskRunRes{nil, apierror.New(apierror.ErrConflict, "")}, "y greater than x"},
// 		{nil, "Gruup1", "AvialSuccess", TaskRunRes{"exec-086d6c629a6bf3581", nil}, "Should return a trimmed error"},
// 		//	{nil, "Gruup1", "Running", TaskRunRes{"", nil}, "negative values, y greater than x"},
// 	}
// 	fmt.Println(cases, "cases")
// 	o := &MockdatasyncOrchestrator{}
// 	for _, test := range cases {
// 		resp, err := o.datasyncOrchestrator.startTaskRun(test.ctx, test.group, test.name)
// 		assert.NotEqual(T, err, nil, "")
// 		assert.Equal(T, test.expected, resp, test.message)
// 	}
// }
