package api

import (
	"encoding/json"
	"fmt"

	"github.com/YaleSpinup/aws-go/services/iam"
)

// moverCreatePolicy returns the IAM inline policy for creating a mover
func (s *server) moverCreatePolicy() (string, error) {
	policy := &iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Sid:    "CreateRole",
				Effect: "Allow",
				Action: []string{
					"iam:CreateRole",
					"iam:GetRole",
					"iam:GetRolePolicy",
					"iam:ListAttachedRolePolicies",
					"iam:ListRolePolicies",
					"iam:AttachRolePolicy",
					"iam:PutRolePolicy",
					"iam:TagRole",
					"iam:UntagRole",
				},
				Resource: []string{
					fmt.Sprintf("arn:aws:iam::*:role/spinup/%s/*", s.org),
				},
			},
			{
				Sid:    "DeleteRole",
				Effect: "Allow",
				Action: []string{
					"iam:DeleteRole",
					"iam:DetachRolePolicy",
					"iam:DeleteRolePolicy",
				},
				Resource: []string{
					fmt.Sprintf("arn:aws:iam::*:role/spinup/%s/*", s.org),
				},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

// moverDeletePolicy returns the IAM inline policy for deleting a mover
func (s *server) moverDeletePolicy() (string, error) {
	policy := &iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Sid:    "DeleteRole",
				Effect: "Allow",
				Action: []string{
					"iam:DeleteRole",
					"iam:GetRole",
					"iam:GetRolePolicy",
					"iam:ListAttachedRolePolicies",
					"iam:ListRolePolicies",
					"iam:DetachRolePolicy",
					"iam:DeleteRolePolicy",
				},
				Resource: []string{
					fmt.Sprintf("arn:aws:iam::*:role/spinup/%s/*", s.org),
				},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}
