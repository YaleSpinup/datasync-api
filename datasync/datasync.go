package datasync

import (
	"context"
	"strings"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/datasync/datasynciface"
	log "github.com/sirupsen/logrus"
)

// Datasync is a wrapper around the aws docdb service
type Datasync struct {
	session         *session.Session
	Service         datasynciface.DataSyncAPI
	DefaultKMSKeyId string
}

type DatasyncOption func(*Datasync)

func New(opts ...DatasyncOption) Datasync {
	e := Datasync{}

	for _, opt := range opts {
		opt(&e)
	}

	if e.session != nil {
		e.Service = datasync.New(e.session)
	}

	return e
}

func WithSession(sess *session.Session) DatasyncOption {
	return func(e *Datasync) {
		log.Debug("using aws session")
		e.session = sess
	}
}

func WithCredentials(key, secret, token, region string) DatasyncOption {
	return func(e *Datasync) {
		log.Debugf("creating new session with key id %s in region %s", key, region)
		sess := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(key, secret, token),
			Region:      aws.String(region),
		}))
		e.session = sess
	}
}

func WithDefaultKMSKeyId(keyId string) DatasyncOption {
	return func(e *Datasync) {
		log.Debugf("using default kms keyid %s", keyId)
		e.DefaultKMSKeyId = keyId
	}
}

// ListDatasyncLocations lists all datasync locations
// returns a map of Location ARNs to Location Types (s3, efs, smb, nfs)
func (d *Datasync) ListDatasyncLocations(ctx context.Context) (map[string]string, error) {
	log.Debug("listing datasync locations")

	filters := []*datasync.LocationFilter{}

	locations := map[string]string{}
	if err := d.Service.ListLocationsPagesWithContext(ctx,
		&datasync.ListLocationsInput{Filters: filters},
		func(page *datasync.ListLocationsOutput, lastPage bool) bool {
			for _, l := range page.Locations {
				lType := strings.SplitN(aws.StringValue(l.LocationUri), ":", 2)
				locations[aws.StringValue(l.LocationArn)] = lType[0]
			}

			return true
		}); err != nil {
		return nil, ErrCode("failed to list locations", err)
	}

	log.Debugf("listing datasync locations output: %+v", locations)

	return locations, nil
}

// ListDatasyncTasks lists all datasync tasks
func (d *Datasync) ListDatasyncTasks(ctx context.Context) ([]string, error) {
	log.Debug("listing datasync tasks")

	filters := []*datasync.TaskFilter{}

	tasks := []string{}
	if err := d.Service.ListTasksPagesWithContext(ctx,
		&datasync.ListTasksInput{Filters: filters},
		func(page *datasync.ListTasksOutput, lastPage bool) bool {
			for _, c := range page.Tasks {
				tasks = append(tasks, aws.StringValue(c.TaskArn))
			}

			return true
		}); err != nil {
		return nil, ErrCode("failed to list tasks", err)
	}

	log.Debugf("listing datasync tasks output: %+v", tasks)

	return tasks, nil
}

// DescribeDatasyncTask return details about a datasync task
func (d *Datasync) DescribeDatasyncTask(ctx context.Context, tArn string) (*datasync.DescribeTaskOutput, error) {
	if !arn.IsARN(tArn) {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid task arn", nil)
	}

	log.Debugf("describing datasync task %s", tArn)

	out, err := d.Service.DescribeTaskWithContext(ctx, &datasync.DescribeTaskInput{
		TaskArn: aws.String(tArn),
	})
	if err != nil {
		return nil, ErrCode("failed to describe task", err)
	}

	return out, nil
}

// DescribeDatasyncLocationS3 return details about an S3 datasync location
func (d *Datasync) DescribeDatasyncLocationS3(ctx context.Context, lArn string) (*datasync.DescribeLocationS3Output, error) {
	if !arn.IsARN(lArn) {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid location arn", nil)
	}

	log.Debugf("describing datasync location (S3) %s", lArn)

	out, err := d.Service.DescribeLocationS3WithContext(ctx, &datasync.DescribeLocationS3Input{
		LocationArn: aws.String(lArn),
	})
	if err != nil {
		return nil, ErrCode("failed to describe location", err)
	}

	return out, nil
}

// DescribeDatasyncLocationEfs returns details about an EFS datasync location
func (d *Datasync) DescribeDatasyncLocationEfs(ctx context.Context, lArn string) (*datasync.DescribeLocationEfsOutput, error) {
	if !arn.IsARN(lArn) {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid location arn", nil)
	}

	log.Debugf("describing datasync location (EFS) %s", lArn)

	out, err := d.Service.DescribeLocationEfsWithContext(ctx, &datasync.DescribeLocationEfsInput{
		LocationArn: aws.String(lArn),
	})
	if err != nil {
		return nil, ErrCode("failed to describe location", err)
	}

	return out, nil
}

// DescribeDatasyncLocationNfs returns details about an NFS datasync location
func (d *Datasync) DescribeDatasyncLocationNfs(ctx context.Context, lArn string) (*datasync.DescribeLocationNfsOutput, error) {
	if !arn.IsARN(lArn) {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid location arn", nil)
	}

	log.Debugf("describing datasync location (NFS) %s", lArn)

	out, err := d.Service.DescribeLocationNfsWithContext(ctx, &datasync.DescribeLocationNfsInput{
		LocationArn: aws.String(lArn),
	})
	if err != nil {
		return nil, ErrCode("failed to describe location", err)
	}

	return out, nil
}

// DescribeDatasyncLocationSmb returns details about an SMB datasync location
func (d *Datasync) DescribeDatasyncLocationSmb(ctx context.Context, lArn string) (*datasync.DescribeLocationSmbOutput, error) {
	if !arn.IsARN(lArn) {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid location arn", nil)
	}

	log.Debugf("describing datasync location (SMB) %s", lArn)

	out, err := d.Service.DescribeLocationSmbWithContext(ctx, &datasync.DescribeLocationSmbInput{
		LocationArn: aws.String(lArn),
	})
	if err != nil {
		return nil, ErrCode("failed to describe location", err)
	}

	return out, nil
}

// GetDatasyncTags gets the tags for a documentDB cluster
func (d *Datasync) GetDatasyncTags(ctx context.Context, tArn string) ([]*datasync.TagListEntry, error) {
	log.Debugf("getting tags for datasync task %s", tArn)

	out, err := d.Service.ListTagsForResourceWithContext(ctx, &datasync.ListTagsForResourceInput{
		ResourceArn: aws.String(tArn),
	})
	if err != nil {
		return nil, ErrCode("failed to get tags", err)
	}

	log.Debugf("getting datasync task tags output: %+v", out)

	return out.Tags, err
}