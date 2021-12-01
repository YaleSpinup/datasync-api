package api

import (
	"context"

	"github.com/YaleSpinup/datasync-api/datasync"
	"github.com/YaleSpinup/datasync-api/resourcegroupstaggingapi"
	log "github.com/sirupsen/logrus"
)

type datasyncOrchestrator struct {
	account        string
	server         *server
	sp             *sessionParams
	datasyncClient datasync.Datasync
	rgClient       *resourcegroupstaggingapi.ResourceGroupsTaggingAPI
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
