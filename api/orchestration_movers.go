package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/aws-go/services/resourcegroupstaggingapi"
	"github.com/YaleSpinup/flywheel"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/datasync"
	log "github.com/sirupsen/logrus"
)

// datamoverCreate creates a data mover and returns the task id of the async Flywheel task
// consisting of a task, source and destination locations
func (o *datasyncOrchestrator) datamoverCreate(ctx context.Context, group string, req *DatamoverCreateRequest) (*flywheel.Task, error) {
	log.Infof("creating data mover %s with source %s and destination %s", aws.StringValue(req.Name), req.Source.Type, req.Destination.Type)

	req.Tags = req.Tags.normalize(o.server.org, group)

	task := flywheel.NewTask()

	// start async orchestration to create all components of the data mover
	go func() {
		taskCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		msgChan, errChan := o.startTask(taskCtx, task)

		// setup err var, rollback function list and defer execution
		// do not shadow err below for rollback to work properly
		var err error
		var rollBackTasks []rollbackFunc
		defer func() {
			if err != nil {
				log.Errorf("recovering from error: %s, executing %d rollback tasks", err, len(rollBackTasks))
				rollBack(&rollBackTasks)
			}
		}()

		var srcLocationArn, dstLocationArn string

		msgChan <- "requested creation of source location"
		srcLocationArn, err = o.createDatasyncLocation(taskCtx, aws.StringValue(req.Name), group, req.Source, req.Tags)
		if err != nil {
			errChan <- fmt.Errorf("failed to create source location: %s", err.Error())
			return
		}

		rollBackTasks = append(rollBackTasks, func(ctx context.Context) error {
			log.Errorf("rollback: deleting source location: %s", srcLocationArn)

			if err := o.deleteDatasyncLocation(ctx, aws.StringValue(req.Name), srcLocationArn, req.Source.Type); err != nil {
				log.Warnf("rollback: error deleting location: %s", err)
				return err
			}

			return nil
		})

		msgChan <- "requested creation of destination location"
		dstLocationArn, err = o.createDatasyncLocation(taskCtx, aws.StringValue(req.Name), group, req.Destination, req.Tags)
		if err != nil {
			errChan <- fmt.Errorf("failed to create destination location: %s", err.Error())
			return
		}

		rollBackTasks = append(rollBackTasks, func(ctx context.Context) error {
			log.Errorf("rollback: deleting destination location: %s", dstLocationArn)

			if err := o.deleteDatasyncLocation(ctx, aws.StringValue(req.Name), dstLocationArn, req.Destination.Type); err != nil {
				log.Warnf("rollback: error deleting location: %s", err)
				return err
			}

			return nil
		})

		var t *datasync.CreateTaskOutput

		msgChan <- fmt.Sprintf("requested creation of datasync task %s", aws.StringValue(req.Name))
		t, err = o.datasyncClient.CreateDatasyncTask(taskCtx, &datasync.CreateTaskInput{
			DestinationLocationArn: aws.String(dstLocationArn),
			Name:                   req.Name,
			SourceLocationArn:      aws.String(srcLocationArn),
			Options: &datasync.Options{
				// eventually we may allow for customizing these
				PreserveDeletedFiles: aws.String("PRESERVE"),
				TransferMode:         aws.String("CHANGED"),
				VerifyMode:           aws.String("ONLY_FILES_TRANSFERRED"),
			},
			Tags: req.Tags.toDatasyncTags(),
		})
		if err != nil {
			errChan <- fmt.Errorf("failed to create datasync task: %s", err.Error())
			return
		}

		a, _ := arn.Parse(aws.StringValue(t.TaskArn))
		parts := strings.SplitN(a.Resource, "/", 2)
		if len(parts) < 2 {
			errChan <- fmt.Errorf("failed to parse datasync task id %s", aws.StringValue(t.TaskArn))
			return
		}
		id := parts[1]

		msgChan <- fmt.Sprintf("created data mover '%s': %s", aws.StringValue(req.Name), id)
	}()

	return task, nil
}

// datamoverDelete deletes a data mover and all of its associated components
func (o *datasyncOrchestrator) datamoverDelete(ctx context.Context, group, name string) error {
	log.Infof("deleting data mover %s", name)

	// get information about the datasync task
	mover, err := o.datamoverDescribe(ctx, group, name)
	if err != nil {
		return err
	}

	// delete task
	if _, err := o.datasyncClient.DeleteDatasyncTask(ctx, &datasync.DeleteTaskInput{
		TaskArn: mover.Task.TaskArn,
	}); err != nil {
		return err
	}

	// delete source and destination locations
	if err := o.deleteDatasyncLocation(ctx, name, aws.StringValue(mover.Task.SourceLocationArn), mover.Source.Type); err != nil {
		return err
	}

	if err := o.deleteDatasyncLocation(ctx, name, aws.StringValue(mover.Task.DestinationLocationArn), mover.Destination.Type); err != nil {
		return err
	}

	return nil
}

// datamoverDescribe gets details about a specific data mover (task and locations)
func (o *datasyncOrchestrator) datamoverDescribe(ctx context.Context, group, name string) (*DatamoverResponse, error) {
	// get information about the task, including tags
	task, tags, err := o.taskDetailsFromName(ctx, group, name)
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

	srcLocation, err := o.describeDatasyncLocation(ctx, srcLocationType, aws.StringValue(task.SourceLocationArn))
	if err != nil {
		return nil, err
	}

	dstLocation, err := o.describeDatasyncLocation(ctx, dstLocationType, aws.StringValue(task.DestinationLocationArn))
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

// datamoverList lists all data movers (tasks) in a group by querying the Resourcegroupstaggingapi
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
		if len(parts) < 2 {
			return nil, apierror.New(apierror.ErrInternalError, "failed to parse ARN "+aws.StringValue(r.ResourceARN), err)
		}

		// we get both tasks and locations back, and we only care about tasks
		if parts[0] == "task" {
			name, err := o.datamoverNameFromArn(ctx, aws.StringValue(r.ResourceARN))
			if err != nil {
				return nil, err
			}
			resources = append(resources, name)
		}
	}

	return resources, nil
}

// describeDatasyncLocation returns information for the specific location type
func (o *datasyncOrchestrator) describeDatasyncLocation(ctx context.Context, lType, lArn string) (*DatamoverLocationOutput, error) {
	if lType == "" || lArn == "" {
		return nil, nil
	}

	log.Debugf("location %s is type %s", lArn, lType)

	switch strings.ToUpper(lType) {
	case S3.String():
		dstLocationS3, err := o.datasyncClient.DescribeDatasyncLocationS3(ctx, lArn)
		if err != nil {
			return nil, err
		}
		return &DatamoverLocationOutput{Type: S3, S3: dstLocationS3}, nil
	case EFS.String():
		dstLocationEfs, err := o.datasyncClient.DescribeDatasyncLocationEfs(ctx, lArn)
		if err != nil {
			return nil, err
		}
		return &DatamoverLocationOutput{Type: EFS, EFS: dstLocationEfs}, nil
	case SMB.String():
		dstLocationSmb, err := o.datasyncClient.DescribeDatasyncLocationSmb(ctx, lArn)
		if err != nil {
			return nil, err
		}
		return &DatamoverLocationOutput{Type: SMB, SMB: dstLocationSmb}, nil
	case NFS.String():
		dstLocationNfs, err := o.datasyncClient.DescribeDatasyncLocationNfs(ctx, lArn)
		if err != nil {
			return nil, err
		}
		return &DatamoverLocationOutput{Type: NFS, NFS: dstLocationNfs}, nil
	default:
		log.Warnf("type %s didn't match any supported location types", lType)
		return nil, apierror.New(apierror.ErrInternalError, "unknown datasync location type "+lType, nil)
	}
}

// createDatasyncLocation creates the specific location type and returns the ARN
func (o *datasyncOrchestrator) createDatasyncLocation(ctx context.Context, mover, group string, input *DatamoverLocationInput, tags Tags) (string, error) {
	if input == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("creating data mover %s location type %s", mover, input.Type)

	switch input.Type {
	case S3:
		s3Arn, err := arn.Parse(aws.StringValue(input.S3.S3BucketArn))
		if err != nil {
			return "", apierror.New(apierror.ErrInternalError, "failed to parse ARN "+aws.StringValue(input.S3.S3BucketArn), err)
		}

		// access to S3 locations is managed via a BucketAccessRole, so
		// we need to generate that first, before creating the location

		path := fmt.Sprintf("/spinup/%s/%s/", o.server.org, group)
		roleName := fmt.Sprintf("%s-%s", mover, s3Arn.Resource)

		roleARN, err := o.bucketAccessRole(ctx, path, roleName, s3Arn.String(), tags)
		if err != nil {
			return "", err
		}

		// if we just created a new role above, it may take some time to propagate across AWS,
		// so we need to retry when creating the location
		var l *datasync.CreateLocationS3Output
		if err = retry(6, 0, 5*time.Second, func() error {
			log.Info("retrying to create datasync location ...")

			var err error
			l, err = o.datasyncClient.CreateDatasyncLocationS3(ctx, &datasync.CreateLocationS3Input{
				S3BucketArn:    input.S3.S3BucketArn,
				S3Config:       &datasync.S3Config{BucketAccessRoleArn: aws.String(roleARN)},
				S3StorageClass: input.S3.S3StorageClass,
				Subdirectory:   input.S3.Subdirectory,
				Tags:           tags.toDatasyncTags(),
			})
			if err != nil {
				log.Debugf("got an error creating location: %s", err)
				return err
			}

			log.Infof("created location successfully: %s", aws.StringValue(l.LocationArn))

			return nil
		}); err != nil {
			log.Infof("failed to create location, timeout retrying: %s", err.Error())

			// clean up the role we created earlier
			if err := o.deleteBucketAccessRole(ctx, aws.String(roleARN)); err != nil {
				log.Warnf("failed deleting role %s: %s", roleARN, err)
			}

			return "", err
		}

		return aws.StringValue(l.LocationArn), nil
	default:
		log.Warnf("type %s didn't match any supported location types", input.Type)
		return "", apierror.New(apierror.ErrBadRequest, "invalid location type", nil)
	}
}

// deleteDatasyncLocation deletes the specific location type and associated resources
// func (o *datasyncOrchestrator) deleteDatasyncLocation(ctx context.Context, mover, group string, input *DatamoverLocationOutput) error {
func (o *datasyncOrchestrator) deleteDatasyncLocation(ctx context.Context, mover, lArn string, lType LocationType) error {
	if mover == "" || lType == "" || lArn == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("deleting data mover %s location type %s", mover, lType)

	l, err := o.describeDatasyncLocation(ctx, lType.String(), lArn)
	if err != nil {
		return err
	}

	switch lType {
	case S3:
		if _, err := o.datasyncClient.DeleteDatasyncLocation(ctx, &datasync.DeleteLocationInput{
			LocationArn: aws.String(lArn),
		}); err != nil {
			log.Warnf("error deleting source location %s: %s", lArn, err)
			return err
		}

		// clean up bucket access role
		if l.S3.S3Config != nil {
			if err := o.deleteBucketAccessRole(ctx, l.S3.S3Config.BucketAccessRoleArn); err != nil {
				return err
			}
		}

		return nil
	default:
		log.Warnf("type %s didn't match any supported location types", lType)
		return apierror.New(apierror.ErrBadRequest, "invalid location type", nil)
	}
}

// datamoverNameFromArn determines the name of a datamover (datasync task) from its ARN
func (o *datasyncOrchestrator) datamoverNameFromArn(ctx context.Context, tArn string) (string, error) {
	if tArn == "" {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	task, err := o.datasyncClient.DescribeDatasyncTask(ctx, tArn)
	if err != nil {
		return "", err
	}

	if task.Name == nil {
		return "", apierror.New(apierror.ErrInternalError, "unable to determine datamover name", nil)
	}

	return aws.StringValue(task.Name), nil
}

// taskDetailsFromName finds a datasync task based on its group/name and returns information about it
func (o *datasyncOrchestrator) taskDetailsFromName(ctx context.Context, group, name string) (*datasync.DescribeTaskOutput, Tags, error) {
	if group == "" || name == "" {
		return nil, nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

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
		{
			Key:   "spinup:spaceid",
			Value: []string{group},
		},
	}

	// get a list of all datasync resources in the group
	out, err := o.rgClient.GetResourcesWithTags(ctx, []string{"datasync:task"}, filters)
	if err != nil {
		return nil, nil, err
	}

	if len(out) == 0 {
		return nil, nil, apierror.New(apierror.ErrNotFound, "datasync mover not found", nil)
	}

	for _, r := range out {
		task, err := o.datasyncClient.DescribeDatasyncTask(ctx, aws.StringValue(r.ResourceARN))
		if err != nil {
			return nil, nil, err
		}

		if aws.StringValue(task.Name) == name {
			return task, fromResourcegroupstaggingapiTags(r.Tags), nil
		}
	}

	return nil, nil, apierror.New(apierror.ErrNotFound, "datasync mover not found", nil)
}

// datamoverRunList returns a list of executions for a given task
func (o *datasyncOrchestrator) datamoverRunList(ctx context.Context, group, name string) ([]string, error) {
	task, _, err := o.taskDetailsFromName(ctx, group, name)
	if err != nil {
		return nil, err
	}

	out, err := o.datasyncClient.ListDatasyncTaskExecutions(ctx, aws.StringValue(task.TaskArn))
	if err != nil {
		return nil, err
	}

	execs := []string{}
	for _, v := range out {
		parts := strings.Split(v, "/")
		if len(parts) > 0 {
			execs = append(execs, parts[len(parts)-1])
		}
	}

	return execs, nil
}

func (o *datasyncOrchestrator) datamoverRunDescribe(ctx context.Context, group, name, id string) (*DatamoverRun, error) {
	task, _, err := o.taskDetailsFromName(ctx, group, name)
	if err != nil {
		return nil, err
	}

	execArn := fmt.Sprintf("%s/execution/%s", aws.StringValue(task.TaskArn), id)

	exec, err := o.datasyncClient.DescribeTaskExecution(ctx, execArn)
	if err != nil {
		return nil, err
	}

	return &DatamoverRun{
		BytesTransferred:         exec.BytesTransferred,
		BytesWritten:             exec.BytesWritten,
		EstimatedBytesToTransfer: exec.EstimatedBytesToTransfer,
		EstimatedFilesToTransfer: exec.EstimatedFilesToTransfer,
		FilesTransferred:         exec.FilesTransferred,
		StartTime:                exec.StartTime,
		Status:                   exec.Status,
		Result:                   exec.Result,
	}, nil
}
