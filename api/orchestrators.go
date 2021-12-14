package api

import (
	"context"
	"strconv"
	"time"

	"github.com/YaleSpinup/aws-go/services/iam"
	"github.com/YaleSpinup/aws-go/services/resourcegroupstaggingapi"
	"github.com/YaleSpinup/datasync-api/common"
	"github.com/YaleSpinup/datasync-api/datasync"
	"github.com/YaleSpinup/flywheel"
	log "github.com/sirupsen/logrus"
)

type datasyncOrchestrator struct {
	account        string
	server         *server
	sp             *sessionParams
	datasyncClient datasync.Datasync
	iamClient      iam.IAM
	rgClient       resourcegroupstaggingapi.ResourceGroupsTaggingAPI
}

// sessionParams stores all required parameters to initialize the connection session
type sessionParams struct {
	role         string
	inlinePolicy string
	policyArns   []string
}

// newDatasyncOrchestrator creates a new session and initializes all clients
func (s *server) newDatasyncOrchestrator(ctx context.Context, account string, sp *sessionParams) (*datasyncOrchestrator, error) {
	log.Debug("initializing datasyncOrchestrator")

	sess, err := s.assumeRole(
		ctx,
		s.session.ExternalID,
		sp.role,
		sp.inlinePolicy,
		sp.policyArns...,
	)
	if err != nil {
		return nil, err
	}

	return &datasyncOrchestrator{
		account:        account,
		server:         s,
		sp:             sp,
		datasyncClient: datasync.New(datasync.WithSession(sess.Session)),
		iamClient:      iam.New(iam.WithSession(sess.Session)),
		rgClient:       resourcegroupstaggingapi.New(resourcegroupstaggingapi.WithSession(sess.Session)),
	}, nil
}

// refreshSession refreshes the session for all client connections
func (o *datasyncOrchestrator) refreshSession(ctx context.Context) error {
	log.Debug("refreshing datasyncOrchestrator session")

	sess, err := o.server.assumeRole(
		ctx,
		o.server.session.ExternalID,
		o.sp.role,
		o.sp.inlinePolicy,
		o.sp.policyArns...,
	)
	if err != nil {
		return err
	}

	o.datasyncClient = datasync.New(datasync.WithSession(sess.Session))
	o.rgClient = resourcegroupstaggingapi.New(resourcegroupstaggingapi.WithSession(sess.Session))

	return nil
}

// startTask starts the flywheel task and receives messages on the channels
func (o *datasyncOrchestrator) startTask(ctx context.Context, task *flywheel.Task) (chan<- string, chan<- error) {
	msgChan := make(chan string)
	errChan := make(chan error)

	// track the task
	go func() {
		taskCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := o.server.flywheel.Start(taskCtx, task); err != nil {
			log.Errorf("failed to start flywheel task, won't be tracked: %s", err)
		}

		for {
			select {
			case msg := <-msgChan:
				log.Infof("task %s: %s", task.ID, msg)

				if ferr := o.server.flywheel.CheckIn(taskCtx, task.ID); ferr != nil {
					log.Errorf("failed to checkin task %s: %s", task.ID, ferr)
				}

				if ferr := o.server.flywheel.Log(taskCtx, task.ID, msg); ferr != nil {
					log.Errorf("failed to log flywheel message for %s: %s", task.ID, ferr)
				}
			case err := <-errChan:
				log.Error(err)

				if ferr := o.server.flywheel.Fail(taskCtx, task.ID, err.Error()); ferr != nil {
					log.Errorf("failed to fail flywheel task %s: %s", task.ID, ferr)
				}

				return
			case <-ctx.Done():
				log.Infof("marking task %s complete", task.ID)

				if ferr := o.server.flywheel.Complete(taskCtx, task.ID); ferr != nil {
					log.Errorf("failed to complete flywheel task %s: %s", task.ID, ferr)
				}

				return
			}
		}
	}()

	return msgChan, errChan
}

func newFlywheelManager(config common.Flywheel) (*flywheel.Manager, error) {
	opts := []flywheel.ManagerOption{}

	if config.RedisAddress != "" {
		opts = append(opts, flywheel.WithRedisAddress(config.RedisAddress))
	}

	if config.RedisUsername != "" {
		opts = append(opts, flywheel.WithRedisAddress(config.RedisUsername))
	}

	if config.RedisPassword != "" {
		opts = append(opts, flywheel.WithRedisAddress(config.RedisPassword))
	}

	if config.RedisDatabase != "" {
		db, err := strconv.Atoi(config.RedisDatabase)
		if err != nil {
			return nil, err
		}
		opts = append(opts, flywheel.WithRedisDatabase(db))
	}

	if config.TTL != "" {
		ttl, err := time.ParseDuration(config.TTL)
		if err != nil {
			return nil, err
		}
		opts = append(opts, flywheel.WithTTL(ttl))
	}

	manager, err := flywheel.NewManager(config.Namespace, opts...)
	if err != nil {
		return nil, err
	}

	return manager, nil
}
