package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/threefoldtech/substrate-client"
)

type Marshalable interface {
	Marshal(d *schema.ResourceData)
	sync(ctx context.Context, sub *substrate.Substrate, cl *ApiClient) (err error)
}

type Action func(context.Context, *substrate.Substrate, *schema.ResourceData, *ApiClient) (Marshalable, error)

func ResourceFunc(a Action) func(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, i interface{}) (diags diag.Diagnostics) {
		return resourceFunc(a, false)(ctx, d, i)
	}
}
func ResourceReadFunc(a Action) func(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, i interface{}) (diags diag.Diagnostics) {
		diags = resourceFunc(a, true)(ctx, d, i)
		if diags.HasError() {
			for idx := range diags {
				diags[idx] = diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Error reading data from remote, terraform state might be out of sync with the remote state",
					Detail:   diags[idx].Summary,
				}
			}
		}
		return diags
	}
}

func resourceFunc(a Action, reportSync bool) func(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, i interface{}) (diags diag.Diagnostics) {
		cl := i.(*ApiClient)
		if err := validateAccountMoneyForExtrinsics(cl.SubstrateConn, cl.Identity); err != nil {
			return diag.FromErr(err)
		}

		obj, err := a(ctx, cl.SubstrateConn, d, cl)
		if err != nil {
			diags = diag.FromErr(err)
		}
		if obj != nil {
			if err := obj.sync(ctx, cl.SubstrateConn, cl); err != nil {
				if reportSync {
					diags = append(diags, diag.FromErr(err)...)
				} else {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Warning,
						Summary:  "failed to read deployment data (terraform refresh might help)",
						Detail:   err.Error(),
					})
				}
			}
			obj.Marshal(d)
		}
		return diags
	}
}
