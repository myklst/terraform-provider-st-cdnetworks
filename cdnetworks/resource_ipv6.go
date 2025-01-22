package cdnetworks

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type ipv6ResourceModel struct {
	DomainId   types.String `tfsdk:"domain_id"`
	EnableIpv6 types.Bool   `tfsdk:"enable_ipv6"`
}

type ipv6Resource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource              = &ipv6Resource{}
	_ resource.ResourceWithConfigure = &ipv6Resource{}
)

func NewIpv6Resource() resource.Resource {
	return &ipv6Resource{}
}

func (r *ipv6Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv6_config"
}

func (r *ipv6Resource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Update DNS region IP version, available value: 'V6'",
		Attributes: map[string]schema.Attribute{
			"domain_id": &schema.StringAttribute{
				Description: "Domain id",
				Required:    true,
			},
			"enable_ipv6": &schema.BoolAttribute{
				Description: "Ipv6",
				Required:    true,
			},
		},
	}
}

func (r *ipv6Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *ipv6Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model ipv6ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// By Default, IPv4 is enabled
	ipVersions := []string{"V4"}
	if model.EnableIpv6.ValueBool() {
		ipVersions = append(ipVersions, "V6")
	}

	addIPv6ConfigResponse, err := r.client.UpdateIPv6Config(model.DomainId.ValueString(), cdnetworksapi.UpdateIPv6ConfigRequest{
		IpVersion: ipVersions,
	})
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Add IPv6", err.Error())
		return
	}

	if *addIPv6ConfigResponse.Code != "0" {
		resp.Diagnostics.AddError("[API ERROR] Error response non 0 code, failed to Add IPv6", *addIPv6ConfigResponse.Message)
		return
	}

	if r.waitForIPv6Config(model) {
		resp.State.Set(ctx, &model)
	} else {
		resp.Diagnostics.AddError("[API ERROR] Failed to Add IPv6", "Timeout")
	}
}

func (r *ipv6Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *ipv6ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryIPv6Response, err := r.client.QueryIPv6Config(state.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query IPv6", err.Error())
		return
	}

	state.DomainId = types.StringPointerValue(queryIPv6Response.DomainId)
	state.EnableIpv6 = types.BoolPointerValue(queryIPv6Response.UseIpv6)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ipv6Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan ipv6ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipVersions := []string{"V4"}
	if plan.EnableIpv6.ValueBool() {
		ipVersions = append(ipVersions, "V6")
	}

	updateIpv6Response, err := r.client.UpdateIPv6Config(state.DomainId.ValueString(), cdnetworksapi.UpdateIPv6ConfigRequest{
		IpVersion: ipVersions,
	})
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Update IPv6", err.Error())
		return
	}

	if *updateIpv6Response.Code != "0" {
		resp.Diagnostics.AddError("[API ERROR] Failed to Update IPv6", *updateIpv6Response.Message)
		return
	}

	if r.waitForIPv6Config(plan) {
		state.DomainId = plan.DomainId
		state.EnableIpv6 = plan.EnableIpv6
		resp.State.Set(ctx, state)
	} else {
		resp.Diagnostics.AddError("[API ERROR] Failed to Add IPv6", "Timeout")
	}
}

func (r *ipv6Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ipv6ResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Due to only have Update() func, force it revert to ipv4 only.
	deleteIPv6Response, err := r.client.UpdateIPv6Config(model.DomainId.ValueString(), cdnetworksapi.UpdateIPv6ConfigRequest{
		IpVersion: []string{"V4"},
	})
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Del IPv6", err.Error())
		return
	}

	if *deleteIPv6Response.Code != "0" {
		resp.Diagnostics.AddError("[API ERROR] Failed to Del IPv6", *deleteIPv6Response.Message)
		return
	}
}

func (r *ipv6Resource) waitForIPv6Config(model ipv6ResourceModel) bool {
	checkStatus := func() error {
		queryIPv6Response, err := r.client.QueryIPv6Config(model.DomainId.ValueString())
		if err != nil {
			return err
		}

		if *queryIPv6Response.UseIpv6 == model.EnableIpv6.ValueBool() {
			return nil
		}

		return errors.New("deployment is in progress")
	}

	s := backoff.NewExponentialBackOff()
	s.InitialInterval = 10 * time.Second
	s.MaxElapsedTime = 0 // set as infinite retries.

	err := backoff.Retry(checkStatus, s)
	return err == nil
}
