// Package provider is the terraform provider
package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"github.com/threefoldtech/terraform-provider-grid/internal/state"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/subi"
)

const errTerraformOutSync = "Error reading data from remote, terraform state might be out of sync with the remote state"
const nameValidationRegex = "^[a-zA-Z0-9_]+$"
const nameValidationErrorMessage = "must only include alphanumeric and underscore characters"
const gpuValidationRegex = "^[A-Za-z0-9:.]+/[A-Za-z0-9]+/[A-Za-z0-9]+$"
const gpuValidationErrMsg = "not a valid gpu id"

// New returns a new schema.Provider instance, and an open substrate connection
func New(version string, st state.Getter) (func() *schema.Provider, subi.SubstrateExt) {
	var substrateConnection subi.SubstrateExt
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"mnemonic": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("MNEMONIC", nil),
				},
				"key_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "key type registered on substrate (ed25519 or sr25519)",
					DefaultFunc: schema.EnvDefaultFunc("KEY_TYPE", "sr25519"),
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
						[]string{"ed25519", "sr25519"},
						false,
					)),
				},
				"network": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "grid network, one of: dev test qa main",
					DefaultFunc: schema.EnvDefaultFunc("NETWORK", "dev"),
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
						[]string{"dev", "qa", "test", "main"},
						false,
					)),
				},
				"substrate_url": {
					Type:             schema.TypeString,
					Optional:         true,
					Description:      "substrate url, example: wss://tfchain.dev.grid.tf/ws",
					DefaultFunc:      schema.EnvDefaultFunc("SUBSTRATE_URL", nil),
					ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithScheme([]string{"wss"})),
				},
				"relay_url": {
					Type:             schema.TypeString,
					Optional:         true,
					Description:      "relay url, example: wss://relay.dev.grid.tf",
					DefaultFunc:      schema.EnvDefaultFunc("RELAY_URL", nil),
					ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithScheme([]string{"wss"})),
				},
				"proxy_url": {
					Type:             schema.TypeString,
					Optional:         true,
					Description:      "proxy url, example: https://gridproxy.dev.grid.tf",
					DefaultFunc:      schema.EnvDefaultFunc("PROXY_URL", nil),
					ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithScheme([]string{"https"})),
				},
				"rmb_timeout": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "timeout duration in seconds for rmb calls",
					DefaultFunc: schema.EnvDefaultFunc("RMB_TIMEOUT", 10),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"grid_gateway_domain": dataSourceGatewayDomain(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"grid_scheduler":  resourceScheduler(),
				"grid_deployment": resourceDeployment(),
				"grid_network":    resourceNetwork(),
				"grid_kubernetes": resourceKubernetes(),
				"grid_name_proxy": resourceGatewayNameProxy(),
				"grid_fqdn_proxy": resourceGatewayFQDNProxy(),
			},
		}
		configFunc, sub := providerConfigure(st)
		substrateConnection = sub
		p.ConfigureContextFunc = configFunc

		return p
	}, substrateConnection
}

func providerConfigure(st state.Getter) (func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics), subi.SubstrateExt) {
	var substrateConn subi.SubstrateExt
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		mnemonic := d.Get("mnemonic").(string)
		keyType := d.Get("key_type").(string)
		network := d.Get("network").(string)
		substrateURL := d.Get("substrate_url").(string)
		relayURL := d.Get("relay_url").(string)
		proxyURL := d.Get("proxy_url").(string)
		timeout := d.Get("rmb_timeout").(int)
		debug := false

		opts := []deployer.PluginOpt{
			deployer.WithNetwork(network),
			deployer.WithTwinCache(),
		}

		if timeout > 0 {
			opts = append(opts, deployer.WithRMBTimeout(timeout))
		}

		if len(strings.TrimSpace(keyType)) != 0 {
			opts = append(opts, deployer.WithKeyType(keyType))
		}

		if len(strings.TrimSpace(substrateURL)) > 0 {
			opts = append(opts, deployer.WithSubstrateURL(substrateURL))
		}

		if len(strings.TrimSpace(proxyURL)) > 0 {
			opts = append(opts, deployer.WithProxyURL(proxyURL))
		}

		if len(strings.TrimSpace(relayURL)) > 0 {
			opts = append(opts, deployer.WithRelayURL(relayURL))
		}

		if debug {
			opts = append(opts, deployer.WithLogs())
		}

		tfPluginClient, err := deployer.NewTFPluginClient(mnemonic, opts...)
		if err != nil {
			return nil, diag.FromErr(errors.Wrap(err, "error creating threefold plugin client"))
		}

		// set state
		tfPluginClient.State.Networks = *st.GetState()

		return &tfPluginClient, nil
	}, substrateConn
}
