// Package provider is the terraform provider
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	client "github.com/threefoldtech/terraform-provider-grid/internal/node"
	"github.com/threefoldtech/terraform-provider-grid/pkg/deployer"
	"github.com/threefoldtech/terraform-provider-grid/pkg/subi"
	"github.com/threefoldtech/terraform-provider-grid/pkg/workloads"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/gridtypes/zos"
)

type GatewayFQDNDeployer struct {
	Gw               workloads.GatewayFQDNProxy
	ID               string
	Description      string
	Node             uint32
	NodeDeploymentID map[uint32]uint64

	ThreefoldPluginClient *threefoldPluginClient
	ncPool                client.NodeClientGetter
	deployer              deployer.Deployer
}

// NewGatewayFQDNDeployer reads the gateway_fqdn_proxy resource configuration data from schema.ResourceData, converts them into a GatewayFQDNDeployer instance, then returns this instance.
func NewGatewayFQDNDeployer(ctx context.Context, d *schema.ResourceData, threefoldPluginClient *threefoldPluginClient) (GatewayFQDNDeployer, error) {
	backendsIf := d.Get("backends").([]interface{})
	backends := make([]zos.Backend, len(backendsIf))
	for idx, n := range backendsIf {
		backends[idx] = zos.Backend(n.(string))
	}
	nodeDeploymentIDIf := d.Get("node_deployment_id").(map[string]interface{})
	nodeDeploymentID := make(map[uint32]uint64)
	for node, id := range nodeDeploymentIDIf {
		nodeInt, err := strconv.ParseUint(node, 10, 32)
		if err != nil {
			return GatewayFQDNDeployer{}, errors.Wrap(err, "couldn't parse node id")
		}
		deploymentID := uint64(id.(int))
		nodeDeploymentID[uint32(nodeInt)] = deploymentID
	}
	ncPool := client.NewNodeClientPool(threefoldPluginClient.rmb, threefoldPluginClient.rmbTimeout)
	deploymentData := DeploymentData{
		Name:        d.Get("name").(string),
		Type:        "gateway",
		ProjectName: d.Get("solution_type").(string),
	}
	deploymentDataStr, err := json.Marshal(deploymentData)
	if err != nil {
		log.Printf("error parsing deploymentdata: %s", err.Error())
	}
	deployer := GatewayFQDNDeployer{
		Gw: workloads.GatewayFQDNProxy{
			Name:           d.Get("name").(string),
			Backends:       backends,
			FQDN:           d.Get("fqdn").(string),
			TLSPassthrough: d.Get("tls_passthrough").(bool),
		},
		ID:                    d.Id(),
		Description:           d.Get("description").(string),
		Node:                  uint32(d.Get("node").(int)),
		NodeDeploymentID:      nodeDeploymentID,
		ThreefoldPluginClient: threefoldPluginClient,
		ncPool:                ncPool,
		deployer:              deployer.NewDeployer(threefoldPluginClient.identity, threefoldPluginClient.twinID, threefoldPluginClient.gridProxyClient, ncPool, true, nil, string(deploymentDataStr)),
	}
	return deployer, nil
}

func (k *GatewayFQDNDeployer) Validate(ctx context.Context, sub subi.SubstrateExt, nodeID uint32) error {

	nodeClient, err := k.ncPool.GetNodeClient(sub, nodeID)
	if err != nil {
		return errors.Wrapf(err, "failed to get node client with ID %d", nodeID)
	}

	cfg, err := nodeClient.NetworkGetPublicConfig(ctx)

	if err != nil {
		return errors.Wrapf(err, "couldn't get node %d public config", nodeID)
	}

	if cfg.IPv4.IP == nil {
		return fmt.Errorf("node %d doesn't contain a public IP in its public config", nodeID)
	}
	return client.AreNodesUp(ctx, sub, []uint32{k.Node}, k.ncPool)
}

// SyncContractsDeployments updates the terraform local state with the resource's latest changes.
func (k *GatewayFQDNDeployer) SyncContractsDeployments(d *schema.ResourceData) (errors error) {

	nodeDeploymentID := make(map[string]interface{})
	for node, id := range k.NodeDeploymentID {
		nodeDeploymentID[fmt.Sprintf("%d", node)] = int(id)
	}

	err := d.Set("node", k.Node)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	err = d.Set("tls_passthrough", k.Gw.TLSPassthrough)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	err = d.Set("backends", k.Gw.Backends)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	err = d.Set("fqdn", k.Gw.FQDN)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	err = d.Set("node_deployment_id", nodeDeploymentID)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	d.SetId(k.ID)

	return
}

func (k *GatewayFQDNDeployer) GenerateVersionlessDeployments(ctx context.Context) (map[uint32]gridtypes.Deployment, error) {
	deployments := make(map[uint32]gridtypes.Deployment)
	dl := workloads.NewDeployment(k.ThreefoldPluginClient.twinID)
	dl.Workloads = append(dl.Workloads, k.Gw.ZosWorkload())
	deployments[k.Node] = dl
	return deployments, nil
}

func (k *GatewayFQDNDeployer) Deploy(ctx context.Context, sub subi.SubstrateExt) error {
	if err := k.Validate(ctx, sub, k.Node); err != nil {
		return err
	}
	newDeployments, err := k.GenerateVersionlessDeployments(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't generate deployments data")
	}
	k.NodeDeploymentID, err = k.deployer.Deploy(ctx, sub, k.NodeDeploymentID, newDeployments)
	if k.ID == "" && k.NodeDeploymentID[k.Node] != 0 {
		k.ID = strconv.FormatUint(k.NodeDeploymentID[k.Node], 10)
	}
	return err
}

func (k *GatewayFQDNDeployer) syncContracts(ctx context.Context, sub subi.SubstrateExt) (err error) {
	if err := sub.DeleteInvalidContracts(k.NodeDeploymentID); err != nil {
		return err
	}
	if len(k.NodeDeploymentID) == 0 {
		// delete resource in case nothing is active (reflects only on read)
		k.ID = ""
	}
	return nil
}

// Sync syncs the deployments
func (k *GatewayFQDNDeployer) Sync(ctx context.Context, sub subi.SubstrateExt, cl *threefoldPluginClient) error {
	if err := k.syncContracts(ctx, sub); err != nil {
		return errors.Wrap(err, "couldn't sync contracts")
	}

	dls, err := k.deployer.GetDeployments(ctx, sub, k.NodeDeploymentID)
	if err != nil {
		return errors.Wrap(err, "couldn't get deployment objects")
	}
	dl := dls[k.Node]
	wl, _ := dl.Get(gridtypes.Name(k.Gw.Name))
	k.Gw = workloads.GatewayFQDNProxy{}
	if wl != nil && wl.Result.State.IsOkay() {
		k.Gw, err = workloads.GatewayFQDNProxyFromZosWorkload(*wl.Workload)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *GatewayFQDNDeployer) Cancel(ctx context.Context, sub subi.SubstrateExt) (err error) {
	newDeployments := make(map[uint32]gridtypes.Deployment)

	k.NodeDeploymentID, err = k.deployer.Deploy(ctx, sub, k.NodeDeploymentID, newDeployments)

	return err
}
