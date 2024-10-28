package provider

import (
	"context"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"sigs.k8s.io/aws-load-balancer-controller/pkg/aws/endpoints"

	"k8s.io/apimachinery/pkg/util/sets"
)

type splitRoleAWSClientsProvider struct {
	// only ec2, elbv2 and shield needs both roles
	// other services only need cluster role
	ec2ClientWithClusterRole          *ec2.Client
	ec2ClientWithServiceLinkedRole    *ec2.Client
	elbv2ClientWithClusterRole        *elasticloadbalancingv2.Client
	elbv2ClientWithServiceLinkedRole  *elasticloadbalancingv2.Client
	shieldClientWithClusterRole       *shield.Client
	shieldClientWithServiceLinkedRole *shield.Client

	acmClientWithClusterRole       *acm.Client
	wafv2ClientWithClusterRole     *wafv2.Client
	wafRegionClientWithClusterRole *wafregional.Client

	ec2OperationsWithServiceLinkedRole    sets.Set[string]
	elbv2OperationsWithServiceLinkedRole  sets.Set[string]
	shieldOperationsWithServiceLinkedRole sets.Set[string]
	rgtClientWithServiceLinkedRole        *resourcegroupstaggingapi.Client
}

// NewSplitRoleAWSClientsProvider creates a new provider using cluster and service-linked roles.
func NewDefaultSplitRoleAWSClientsProvider(cfg aws.Config, endpointsResolver *endpoints.Resolver, serviceLinkedRoleArn string, clusterRoleArn string) (*splitRoleAWSClientsProvider, error) {
	if serviceLinkedRoleArn == "" || clusterRoleArn == "" {
		return nil, errors.New("both serviceLinkedRoleArn and clusterRoleArn must be specified")
	}
	// Create an STS client for role assumption
	stsClient := sts.NewFromConfig(cfg)

	// Create credentials for the service-linked role
	serviceLinkedRoleCreds := stscreds.NewAssumeRoleProvider(stsClient, serviceLinkedRoleArn)

	// Create credentials for the cluster role
	clusterRoleCreds := stscreds.NewAssumeRoleProvider(stsClient, clusterRoleArn)

	// Set up custom endpoints for services
	ec2CustomEndpoint := endpointsResolver.EndpointFor(ec2.ServiceID)
	elbv2CustomEndpoint := endpointsResolver.EndpointFor(elasticloadbalancingv2.ServiceID)
	acmCustomEndpoint := endpointsResolver.EndpointFor(acm.ServiceID)
	wafv2CustomEndpoint := endpointsResolver.EndpointFor(wafv2.ServiceID)
	wafregionalCustomEndpoint := endpointsResolver.EndpointFor(wafregional.ServiceID)
	shieldCustomEndpoint := endpointsResolver.EndpointFor(shield.ServiceID)
	rgtCustomEndpoint := endpointsResolver.EndpointFor(resourcegroupstaggingapi.ServiceID)

	// Create clients using the cluster role
	ec2ClientWithClusterRole := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		o.Credentials = aws.NewCredentialsCache(clusterRoleCreds)
		if ec2CustomEndpoint != nil {
			o.BaseEndpoint = ec2CustomEndpoint
		}
	})
	elbv2ClientWithClusterRole := elasticloadbalancingv2.NewFromConfig(cfg, func(o *elasticloadbalancingv2.Options) {
		o.Credentials = aws.NewCredentialsCache(clusterRoleCreds)
		if elbv2CustomEndpoint != nil {
			o.BaseEndpoint = elbv2CustomEndpoint
		}
	})
	acmClientWithClusterRole := acm.NewFromConfig(cfg, func(o *acm.Options) {
		o.Credentials = aws.NewCredentialsCache(clusterRoleCreds)
		if acmCustomEndpoint != nil {
			o.BaseEndpoint = acmCustomEndpoint
		}
	})
	wafv2ClientWithClusterRole := wafv2.NewFromConfig(cfg, func(o *wafv2.Options) {
		o.Credentials = aws.NewCredentialsCache(clusterRoleCreds)
		if wafv2CustomEndpoint != nil {
			o.BaseEndpoint = wafv2CustomEndpoint
		}
	})
	wafregionalClientWithClusterRole := wafregional.NewFromConfig(cfg, func(o *wafregional.Options) {
		o.Credentials = aws.NewCredentialsCache(clusterRoleCreds)
		o.Region = cfg.Region
		o.BaseEndpoint = wafregionalCustomEndpoint
	})
	shieldClientWithClusterRole := shield.NewFromConfig(cfg, func(o *shield.Options) {
		o.Credentials = aws.NewCredentialsCache(clusterRoleCreds)
		o.Region = "us-east-1"
		o.BaseEndpoint = shieldCustomEndpoint
	})

	// Create clients using the service-linked role
	ec2ClientWithServiceLinkedRole := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		o.Credentials = aws.NewCredentialsCache(serviceLinkedRoleCreds)
		if ec2CustomEndpoint != nil {
			o.BaseEndpoint = ec2CustomEndpoint
		}
	})
	elbv2ClientWithServiceLinkedRole := elasticloadbalancingv2.NewFromConfig(cfg, func(o *elasticloadbalancingv2.Options) {
		o.Credentials = aws.NewCredentialsCache(serviceLinkedRoleCreds)
		if elbv2CustomEndpoint != nil {
			o.BaseEndpoint = elbv2CustomEndpoint
		}
	})
	shieldClientWithServiceLinkedRole := shield.NewFromConfig(cfg, func(o *shield.Options) {
		o.Credentials = aws.NewCredentialsCache(serviceLinkedRoleCreds)
		o.Region = "us-east-1"
		o.BaseEndpoint = shieldCustomEndpoint
	})

	rgtClientWithServiceLinkedRole := resourcegroupstaggingapi.NewFromConfig(cfg, func(o *resourcegroupstaggingapi.Options) {
		o.Credentials = aws.NewCredentialsCache(serviceLinkedRoleCreds)
		if rgtCustomEndpoint != nil {
			o.BaseEndpoint = rgtCustomEndpoint
		}
	})

	// Initialize operations sets
	// source: https://code.amazon.com/packages/EKSTachyonIAMProposal/commits/30360c89afaf3fc7d3d74374cbdd6f3143aedbe4#AmazonEKSLoadBalancingServiceRolePolicy.json
	ec2OperationsWithServiceLinkedRole := sets.New[string](
		"DeleteSecurityGroup",
		"DescribeAccountAttributes",
		"DescribeAddresses",
		"DescribeAvailabilityZones",
		"DescribeInternetGateways",
		"DescribeVpcs",
		"DescribeVpcPeeringConnections",
		"DescribeSubnets",
		"DescribeSecurityGroups",
		"DescribeInstances",
		"DescribeNetworkInterfaces",
		"DescribeTags",
		"GetCoipPoolUsage",
		"DescribeCoipPools",
	)
	elbv2OperationsWithServiceLinkedRole := sets.New[string](
		"DeleteListener",
		"DeleteRule",
		"DeregisterTargets",
		"DeleteLoadBalancer",
		"DeleteTargetGroup",
		"DescribeLoadBalancers",
		"DescribeLoadBalancerAttributes",
		"DescribeListeners",
		"DescribeListenerCertificates",
		"DescribeSSLPolicies",
		"DescribeRules",
		"DescribeTargetGroups",
		"DescribeTargetGroupAttributes",
		"DescribeTargetHealth",
		"DescribeTags",
		"DescribeTrustStores",
		"DescribeListenerAttributes",
	)
	shieldOperationsWithServiceLinkedRole := sets.New[string](
		"DescribeProtection",
		"GetSubscriptionState",
	)

	// Return the constructed splitRoleAWSClientsProvider
	return &splitRoleAWSClientsProvider{
		ec2ClientWithClusterRole:              ec2ClientWithClusterRole,
		ec2ClientWithServiceLinkedRole:        ec2ClientWithServiceLinkedRole,
		elbv2ClientWithClusterRole:            elbv2ClientWithClusterRole,
		elbv2ClientWithServiceLinkedRole:      elbv2ClientWithServiceLinkedRole,
		shieldClientWithClusterRole:           shieldClientWithClusterRole,
		shieldClientWithServiceLinkedRole:     shieldClientWithServiceLinkedRole,
		acmClientWithClusterRole:              acmClientWithClusterRole,
		wafv2ClientWithClusterRole:            wafv2ClientWithClusterRole,
		wafRegionClientWithClusterRole:        wafregionalClientWithClusterRole,
		rgtClientWithServiceLinkedRole:        rgtClientWithServiceLinkedRole,
		ec2OperationsWithServiceLinkedRole:    ec2OperationsWithServiceLinkedRole,
		elbv2OperationsWithServiceLinkedRole:  elbv2OperationsWithServiceLinkedRole,
		shieldOperationsWithServiceLinkedRole: shieldOperationsWithServiceLinkedRole,
	}, nil
}

func (p *splitRoleAWSClientsProvider) GetEC2Client(ctx context.Context, operationName string) (*ec2.Client, error) {
	if p.ec2OperationsWithServiceLinkedRole.Has(operationName) {
		return p.ec2ClientWithServiceLinkedRole, nil
	} else {
		return p.ec2ClientWithClusterRole, nil
	}
}

func (p *splitRoleAWSClientsProvider) GetELBv2Client(ctx context.Context, operationName string) (*elasticloadbalancingv2.Client, error) {
	if p.elbv2OperationsWithServiceLinkedRole.Has(operationName) {
		return p.elbv2ClientWithServiceLinkedRole, nil
	} else {
		return p.elbv2ClientWithClusterRole, nil
	}
}

func (p *splitRoleAWSClientsProvider) GetACMClient(ctx context.Context, operationName string) (*acm.Client, error) {
	return p.acmClientWithClusterRole, nil
}

func (p *splitRoleAWSClientsProvider) GetWAFv2Client(ctx context.Context, operationName string) (*wafv2.Client, error) {
	return p.wafv2ClientWithClusterRole, nil
}

func (p *splitRoleAWSClientsProvider) GetWAFRegionClient(ctx context.Context, operationName string) (*wafregional.Client, error) {
	return p.wafRegionClientWithClusterRole, nil
}

func (p *splitRoleAWSClientsProvider) GetShieldClient(ctx context.Context, operationName string) (*shield.Client, error) {
	if p.shieldOperationsWithServiceLinkedRole.Has(operationName) {
		return p.shieldClientWithServiceLinkedRole, nil
	} else {
		return p.shieldClientWithClusterRole, nil
	}
}

func (p *splitRoleAWSClientsProvider) GetRGTClient(ctx context.Context, operationName string) (*resourcegroupstaggingapi.Client, error) {
	return p.rgtClientWithServiceLinkedRole, nil
}
