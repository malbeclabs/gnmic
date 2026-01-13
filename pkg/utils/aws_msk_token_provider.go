// © 2024 Nokia.
//
// This code is a Contribution to the gNMIc project ("Work") made under the Google Software Grant and Corporate Contributor License Agreement ("CLA") and governed by the Apache License 2.0.
// No other rights or licenses in or to any of Nokia's intellectual property are granted for any other purpose.
// This code is provided on an "as is" basis without any warranties of any kind.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"os"

	"github.com/IBM/sarama"
	"github.com/aws/aws-msk-iam-sasl-signer-go/signer"
)

// MSKTokenProvider implements sarama.AccessTokenProvider for AWS MSK IAM authentication.
// It uses the AWS SDK default credential chain, supporting:
// - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
// - IAM roles for service accounts (IRSA) in Kubernetes
// - EC2 instance profiles
// - ECS task roles
// - Shared credentials file (~/.aws/credentials)
type MSKTokenProvider struct {
	region string
}

// NewMSKTokenProvider creates a new MSKTokenProvider for the specified AWS region.
// If region is empty, it will attempt to use the AWS_REGION or AWS_DEFAULT_REGION
// environment variables.
func NewMSKTokenProvider(region string) sarama.AccessTokenProvider {
	if region == "" {
		region = os.Getenv("AWS_REGION")
		if region == "" {
			region = os.Getenv("AWS_DEFAULT_REGION")
		}
	}
	return &MSKTokenProvider{
		region: region,
	}
}

// Token generates an AWS MSK IAM authentication token using the default credential chain.
func (m *MSKTokenProvider) Token() (*sarama.AccessToken, error) {
	token, _, err := signer.GenerateAuthToken(context.Background(), m.region)
	if err != nil {
		return nil, err
	}
	return &sarama.AccessToken{Token: token}, nil
}
