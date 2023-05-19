package provider

import (
	"context"
	"fmt"

	"github.com/openshift/osde2e-framework/pkg/clients/ocm"
	"github.com/openshift/osde2e-framework/pkg/providers/rosa"
)

type Cluster struct {
	ChannelGroup string
	Name         string
	Version      string

	id string
}

type Provider struct {
	*rosa.Provider
	cluster *Cluster
}

func New(ctx context.Context, token string, environment ocm.Environment, cluster *Cluster) (*Provider, error) {
	provider, err := rosa.New(ctx, token, environment)
	if err != nil {
		return nil, err
	}
	return &Provider{
		Provider: provider,
		cluster:  cluster,
	}, nil
}

func (p *Provider) CreateHCPClusters(ctx context.Context) error {
	fmt.Printf("Provision cluster w/options: %+v\n", p.cluster)

	clusterID, err := p.CreateCluster(ctx, &rosa.CreateClusterOptions{
		ClusterName:  p.cluster.Name,
		Version:      p.cluster.Version,
		ChannelGroup: p.cluster.ChannelGroup,
		HostedCP:     true,
	})

	p.cluster.id = clusterID
	return err
}

func (p *Provider) DeleteHCPClusters(ctx context.Context) error {
	fmt.Printf("Delete cluster w/options: %+v", p.cluster)

	return p.DeleteCluster(ctx, &rosa.DeleteClusterOptions{
		ClusterName: p.cluster.Name,
		ClusterID:   p.cluster.id,
		HostedCP:    true,
	})
}
