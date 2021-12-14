package api

import (
	"context"
	"encoding/json"

	"github.com/YaleSpinup/apierror"
	yiam "github.com/YaleSpinup/aws-go/services/iam"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	log "github.com/sirupsen/logrus"
)

var assumeRolePolicyDoc []byte

// BucketAccessRole generates the role (if it doesn't exist) for DataSync access to S3 bucket and returns the ARN
func (o *datasyncOrchestrator) BucketAccessRole(ctx context.Context, path, role, bucketArn string, tags []Tag) (string, error) {
	if path == "" || role == "" {
		return "", apierror.New(apierror.ErrBadRequest, "invalid path", nil)
	}

	log.Infof("generating bucket access role %s%s if it doesn't exist ", path, role)

	defaultPolicy := bucketAccessPolicy(bucketArn)

	var roleArn string
	if out, err := o.iamClient.GetRole(ctx, role); err != nil {
		if aerr, ok := err.(apierror.Error); !ok || aerr.Code != apierror.ErrNotFound {
			return "", err
		}

		log.Debugf("unable to find role %s%s, creating", path, role)

		output, err := o.createBucketAccessRole(ctx, path, role)
		if err != nil {
			return "", err
		}

		roleArn = output

		log.Infof("created role %s/%s with ARN: %s", path, role, roleArn)
	} else {
		roleArn = aws.StringValue(out.Arn)

		log.Infof("role %s exists with ARN: %s", role, roleArn)

		currentDoc, err := o.iamClient.GetRolePolicy(ctx, role, "DataSyncBucketAccessPolicy")
		if err != nil {
			if aerr, ok := err.(apierror.Error); !ok || aerr.Code != apierror.ErrNotFound {
				return "", err
			}

			log.Infof("inline policy for role %s/%s is not found, updating", path, role)
		} else {
			var currentPolicy yiam.PolicyDocument
			if err := json.Unmarshal([]byte(currentDoc), &currentPolicy); err != nil {
				log.Errorf("failed to unmarhsall policy from document: %s", err)
				return "", err
			}

			// if the current policy matches the generated (default) policy, return
			// the role ARN, otherwise, keep going and update the policy doc
			if yiam.PolicyDeepEqual(defaultPolicy, currentPolicy) {
				log.Debugf("inline policy for role %s%s is up to date", path, role)
				return roleArn, nil
			}

			log.Infof("inline policy for role %s%s is out of date, updating", path, role)
		}
	}

	defaultPolicyDoc, err := json.Marshal(defaultPolicy)
	if err != nil {
		log.Errorf("failed creating default bucket access policy for %s: %s", path, err.Error())
		return "", err
	}

	// attach default role policy to the role
	err = o.iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(string(defaultPolicyDoc)),
		PolicyName:     aws.String("DataSyncBucketAccessPolicy"),
		RoleName:       aws.String(role),
	})
	if err != nil {
		return "", err
	}

	// apply tags if any were passed
	if len(tags) > 0 {
		iamTags := make([]*iam.Tag, len(tags))
		for i, t := range tags {
			iamTags[i] = &iam.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
		}

		if err := o.iamClient.TagRole(ctx, role, iamTags); err != nil {
			return "", err
		}
	}

	return roleArn, nil
}

// createBucketAccessRole handles creating the bucket access role
func (o *datasyncOrchestrator) createBucketAccessRole(ctx context.Context, path, role string) (string, error) {
	if role == "" {
		return "", apierror.New(apierror.ErrBadRequest, "invalid role", nil)
	}

	log.Debugf("creating bucket access role %s", role)

	assumeRolePolicyDoc, err := assumeRolePolicy()
	if err != nil {
		log.Errorf("failed to generate assume role policy for %s: %s", role, err)
		return "", err
	}

	log.Debugf("generated assume role policy document: %s", assumeRolePolicyDoc)

	roleOutput, err := o.iamClient.CreateRole(ctx, &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicyDoc),
		Description:              aws.String("DataSync bucket access role"),
		Path:                     aws.String(path),
		RoleName:                 aws.String(role),
	})
	if err != nil {
		return "", err
	}

	return aws.StringValue(roleOutput.Arn), nil
}

// assumeRolePolicy generates the policy document to allow the datasync service to assume a role
func assumeRolePolicy() (string, error) {
	if assumeRolePolicyDoc != nil {
		return string(assumeRolePolicyDoc), nil
	}

	policyDoc, err := json.Marshal(yiam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []yiam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"sts:AssumeRole",
				},
				Principal: yiam.Principal{
					"Service": {"datasync.amazonaws.com"},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	// cache result since it doesn't change
	assumeRolePolicyDoc = policyDoc

	return string(policyDoc), nil
}

// bucketAccessPolicy generates the policy for DataSync bucket access
func bucketAccessPolicy(bucketArn string) yiam.PolicyDocument {
	log.Debugf("generating bucket access policy for %s", bucketArn)

	return yiam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []yiam.StatementEntry{
			{
				Sid:    "ListBucket",
				Effect: "Allow",
				Action: []string{
					"s3:GetBucketLocation",
					"s3:ListBucket",
					"s3:ListBucketMultipartUploads",
				},
				Resource: []string{bucketArn},
			},
			{
				Sid:    "GetBucketObjects",
				Effect: "Allow",
				Action: []string{
					"s3:AbortMultipartUpload",
					"s3:DeleteObject",
					"s3:GetObject",
					"s3:ListMultipartUploadParts",
					"s3:GetObjectTagging",
					"s3:PutObjectTagging",
					"s3:PutObject",
				},
				Resource: []string{bucketArn + "/*"},
			},
		},
	}
}
