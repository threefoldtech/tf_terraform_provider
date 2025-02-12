// Package provider is the terraform provider
package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

func resourceKubernetes() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Resource to deploy a kubernetes cluster. A cluster should consist of one master node, and a number (could be zero) of worker nodes.",

		CreateContext: resourceK8sCreate,
		ReadContext:   resourceK8sRead,
		UpdateContext: resourceK8sUpdate,
		DeleteContext: resourceK8sDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Solution name for the created contracts to be consistent across threefold tooling. Must contain only alphanumeric and underscore characters.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringMatch(regexp.MustCompile(nameValidationRegex), nameValidationErrorMessage)),
			},
			"solution_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Solution type for the created contracts to be consistent across threefold tooling.",
			},
			"node_deployment_id": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Mapping from each node to its deployment id (contract id).",
			},
			"network_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The network name to deploy the cluster on.",
			},
			"flist": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Flist used on all nodes, e.g. https://hub.grid.tf/tf-official-apps/threefolddev-k3s-v1.31.0.flist. All flists could be found in `https://hub.grid.tf/`",
			},
			"entrypoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Command to execute as the kubernetes node init.",
			},
			"flist_checksum": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "If present, the flist is rejected if it has a different hash.",
			},
			"ssh_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "SSH key to access the cluster nodes.",
			},
			"token": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The cluster secret token. Each node has to have this token to be part of the cluster. This token should be an alphanumeric non-empty string.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			},
			"nodes_ip_range": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Reserved network IP ranges for nodes in the cluster (this is assigned from grid_network.<network-resource-name>.nodes_ip_range).",
			},
			"master": {
				MaxItems:    1,
				Type:        schema.TypeList,
				Required:    true,
				Description: "Master holds the configuration of master node in the kubernetes cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Master node ZMachine workload name.  This has to be unique within the node. Must contain only alphanumeric and underscore characters.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringMatch(regexp.MustCompile(nameValidationRegex), nameValidationErrorMessage)),
						},
						"node": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Node ID to deploy master node on.",
						},
						"disk_size": {
							Type:             schema.TypeInt,
							Required:         true,
							Description:      "Disk size for master node in GBs. Must be between 1GB and 10240GBs (10TBs).",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10*1024)),
						},
						"publicip": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Flag to enable/disable public ipv4 reservation.",
						},
						"publicip6": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Flag to enable/disable public ipv6 reservation.",
						},
						"flist": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "https://hub.grid.tf/tf-official-apps/threefolddev-k3s-v1.31.0.flist",
							Description: "Flist used on master node, e.g. https://hub.grid.tf/tf-official-apps/threefolddev-k3s-v1.31.0.flist. All flists could be found in `https://hub.grid.tf/`",
						},
						"entrypoint": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "/sbin/zinit init",
							Description: "Command to execute as the kubernetes node init.",
						},
						"flist_checksum": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "if present, the flist is rejected if it has a different hash.",
						},
						"computedip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The reserved public IPv4.",
						},
						"computedip6": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The reserved public IPv6.",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private wireguard IP of master node.",
						},
						"cpu": {
							Type:             schema.TypeInt,
							Required:         true,
							Description:      "Number of virtual CPUs. Must be between 1 and 32.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 32)),
						},
						"memory": {
							Type:             schema.TypeInt,
							Required:         true,
							Description:      "Memory size in MB. Must be between 256MBs and 262144MBs (256GBs).",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(256, 256*1024)),
						},
						"planetary": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Flag to enable Yggdrasil IP allocation.",
						},
						"planetary_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The allocated Yggdrasil IP.",
						},
						"mycelium_ip_seed": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Mycelium seed used to get the same mycelium ip for k8s master. Hex encoded 6 bytes (e.g. b60f2b7ec39c).",
						},
						"mycelium_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The allocated mycelium IP.",
						},
						"console_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The url to access the vm via cloud console on private interface using wireguard.",
						},
						"network_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network name.",
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ssh key.",
						},
						"token": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "cluster token.",
						},
					},
				},
			},
			"workers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Workers is a list holding the workers configuration for the kubernetes cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Worker node ZMachine workload name. This has to be unique within the node. Must contain only alphanumeric and underscore characters.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringMatch(regexp.MustCompile(nameValidationRegex), nameValidationErrorMessage)),
						},
						"flist": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Flist used on worker node, e.g. https://hub.grid.tf/tf-official-apps/threefolddev-k3s-v1.31.0.flist. All flists could be found in `https://hub.grid.tf/`.",
						},
						"entrypoint": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Command to execute as the kubernetes node init.",
						},
						"flist_checksum": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "if present, the flist is rejected if it has a different hash.",
						},
						"disk_size": {
							Type:             schema.TypeInt,
							Required:         true,
							Description:      "Data disk size in GBs. Must be between 1GB and 10240GBs (10TBs).",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10*1024)),
						},
						"node": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Node ID to deploy worker node on.",
						},
						"publicip": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Flag to enable/disable public ipv4 reservation.",
						},
						"computedip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The reserved public ipv4.",
						},
						"publicip6": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Flag to enable/disable public ipv6 reservation.",
						},
						"computedip6": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The reserved public ipv6.",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private IP (computed from nodes_ip_range).",
						},
						"cpu": {
							Type:             schema.TypeInt,
							Required:         true,
							Description:      "Number of virtual CPUs. Must be between 1 and 32.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 32)),
						},
						"memory": {
							Type:             schema.TypeInt,
							Required:         true,
							Description:      "Memory size in MB. Must be between 256MBs and 262144MBs (256GBs).",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(256, 256*1024)),
						},
						"planetary": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Flag to enable Yggdrasil IP allocation.",
						},
						"planetary_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The allocated Yggdrasil IP.",
						},
						"mycelium_ip_seed": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Mycelium seed used to get the same mycelium ip for k8s worker. Hex encoded 6 bytes (e.g. b60f2b7ec39c).",
						},
						"mycelium_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The allocated mycelium IP.",
						},
						"console_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The url to access the vm via cloud console on private interface using wireguard.",
						},
						"network_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network name.",
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ssh key.",
						},
						"token": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "cluster token.",
						},
					},
				},
			},
		},
	}
}

func resourceK8sCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tfPluginClient, ok := meta.(*deployer.TFPluginClient)
	if !ok {
		return diag.FromErr(fmt.Errorf("failed to cast meta into threefold plugin client"))
	}

	k8sCluster, err := newK8sFromSchema(d)
	if err != nil {
		return diag.Errorf("couldn't load k8s cluster data with error: %v", err)
	}

	if err := tfPluginClient.K8sDeployer.Deploy(ctx, k8sCluster); err != nil {
		return diag.Errorf("couldn't deploy k8s cluster with error: %v", err)
	}

	err = tfPluginClient.K8sDeployer.UpdateFromRemote(ctx, k8sCluster)
	if err != nil {
		return diag.Errorf("couldn't update k8s cluster from remote with error: %v", err)
	}

	err = storeK8sState(d, k8sCluster)
	if err != nil {
		diags = diag.FromErr(err)
	}

	d.SetId(uuid.New().String())
	return diags
}

func resourceK8sUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tfPluginClient, ok := meta.(*deployer.TFPluginClient)
	if !ok {
		return diag.FromErr(fmt.Errorf("failed to cast meta into threefold plugin client"))
	}

	k8sCluster, err := newK8sFromSchema(d)
	if err != nil {
		return diag.Errorf("couldn't load k8s cluster data with error: %v", err)
	}

	if err := tfPluginClient.K8sDeployer.Deploy(ctx, k8sCluster); err != nil {
		return diag.Errorf("couldn't update k8s cluster with error: %v", err)
	}

	err = tfPluginClient.K8sDeployer.UpdateFromRemote(ctx, k8sCluster)
	if err != nil {
		return diag.Errorf("couldn't update k8s cluster from remote with error: %v", err)
	}

	err = storeK8sState(d, k8sCluster)
	if err != nil {
		diags = diag.FromErr(err)
	}

	return diags
}

func resourceK8sRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tfPluginClient, ok := meta.(*deployer.TFPluginClient)
	if !ok {
		return diag.FromErr(fmt.Errorf("failed to cast meta into threefold plugin client"))
	}

	k8sCluster, err := newK8sFromSchema(d)
	if err != nil {
		return diag.Errorf("couldn't load k8s cluster data with error: %v", err)
	}

	if err := tfPluginClient.K8sDeployer.Validate(ctx, k8sCluster); err != nil {
		return diag.FromErr(err)
	}

	if err := k8sCluster.InvalidateBrokenAttributes(tfPluginClient.SubstrateConn); err != nil {
		return diag.FromErr(errors.Wrap(err, "couldn't invalidate broken attributes"))
	}

	err = tfPluginClient.K8sDeployer.UpdateFromRemote(ctx, k8sCluster)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  errTerraformOutSync,
			Detail:   err.Error(),
		})
		return diags
	}

	err = storeK8sState(d, k8sCluster)
	if err != nil {
		diags = diag.FromErr(err)
	}

	return diags
}

func resourceK8sDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tfPluginClient, ok := meta.(*deployer.TFPluginClient)
	if !ok {
		return diag.FromErr(fmt.Errorf("failed to cast meta into threefold plugin client"))
	}

	k8sCluster, err := newK8sFromSchema(d)
	if err != nil {
		return diag.Errorf("couldn't load k8s cluster data with error: %v", err)
	}

	if err := tfPluginClient.K8sDeployer.Cancel(ctx, k8sCluster); err != nil {
		return diag.Errorf("couldn't cancel k8s cluster with error: %v", err)
	}

	d.SetId("")
	return nil
}
