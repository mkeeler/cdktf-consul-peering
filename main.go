package main

import (
	"fmt"

	"cdk.tf/go/stack/generated/hashicorp/consul"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ClusterName string

type Config struct {
	Providers map[ClusterName]consul.ConsulProviderConfig
	Peerings  []Peering
}

type Peering struct {
	Dialer   ClusterID
	Acceptor ClusterID
}

type ClusterID struct {
	Name      ClusterName
	Partition string
}

func strPointer(s string) *string {
	return &s
}

var (
	conf = Config{
		Providers: map[ClusterName]consul.ConsulProviderConfig{
			"alpha": {
				Address: strPointer("https://localhost:55000"),
				Alias:   strPointer("alpha"),
				CaFile:  strPointer("/Users/mkeeler/code/mkeeler/consul-docker-test/peering/servers/alpha/cacert.pem"),
				Token:   strPointer("c181ddfd-8a1b-8aa2-7d61-a04e179400cd"),
			},
			"beta": {
				Address: strPointer("https://localhost:55001"),
				Alias:   strPointer("beta"),
				CaFile:  strPointer("/Users/mkeeler/code/mkeeler/consul-docker-test/peering/servers/beta/cacert.pem"),
				Token:   strPointer("65e5c4f1-1428-c41b-71b0-0991f3797210"),
			},
		},
		Peerings: []Peering{
			{
				Dialer:   ClusterID{Name: "alpha", Partition: "foo"},
				Acceptor: ClusterID{Name: "beta", Partition: "bar"},
			},
			{
				Dialer:   ClusterID{Name: "beta", Partition: "default"},
				Acceptor: ClusterID{Name: "alpha", Partition: "default"},
			},
		},
	}
)

func NewMyStack(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)

	// Instantiate all the providers
	providers := make(map[ClusterName]consul.ConsulProvider)
	for clusterName, providerConfig := range conf.Providers {
		providers[clusterName] = consul.NewConsulProvider(stack, jsii.String(string(clusterName)), &providerConfig)
	}

	for _, peering := range conf.Peerings {
		dialingPeerName := fmt.Sprintf("%s-%s", peering.Acceptor.Name, peering.Acceptor.Partition)
		acceptingPeerName := fmt.Sprintf("%s-%s", peering.Dialer.Name, peering.Dialer.Partition)
		id := fmt.Sprintf("%s:%s", dialingPeerName, acceptingPeerName)
		token := consul.NewPeeringToken(stack, jsii.String("peering-token:"+id), &consul.PeeringTokenConfig{
			Provider:  providers[peering.Acceptor.Name],
			PeerName:  &acceptingPeerName,
			Partition: &peering.Acceptor.Partition,
		})

		consul.NewPeering(stack, jsii.String("peering:"+id), &consul.PeeringConfig{
			Provider:     providers[peering.Dialer.Name],
			PeerName:     &dialingPeerName,
			Partition:    &peering.Dialer.Partition,
			PeeringToken: token.PeeringToken(),
		})
	}

	return stack
}

func main() {
	app := cdktf.NewApp(nil)

	NewMyStack(app, "cdktf-consul-peering")

	app.Synth()
}
