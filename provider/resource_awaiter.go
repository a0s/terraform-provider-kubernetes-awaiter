package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/url"
	"time"
)

func kubernetesResourceAwaiter() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cacert": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "CA certificate content",
				ExactlyOneOf: []string{"cacert", "cacert_path"},
				ForceNew:     true,
			},
			"cacert_path": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "path to CA certificate file",
				ExactlyOneOf: []string{"cacert", "cacert_path"},
				ForceNew:     true,
			},
			"token": {
				Type:        schema.TypeString,
				Description: "bearer token for serviceaccount",
				Required:    true,
				ForceNew:    true,
			},
			"timeout": {
				Type:        schema.TypeString,
				Description: "polling timeout, ex: 2m30s",
				Optional:    true,
				Default:     "5m",
			},
			"poll": {
				Type:        schema.TypeString,
				Description: "polling interval, ex: 1s",
				Optional:    true,
				Default:     "1s",
			},
			"uri": {
				Type:        schema.TypeString,
				Description: "full uri to kubernetes api resource",
				Required:    true,
				ForceNew:    true,
			},
		},
		CreateContext: kubernetesResourceAwaiterCreate,
		ReadContext:   kubernetesResourceAwaiterRead,
		UpdateContext: kubernetesResourceAwaiterUpdate,
		DeleteContext: kubernetesResourceAwaiterDelete,
	}
}

func kubernetesResourceAwaiterCreate(c context.Context, r *schema.ResourceData, _ interface{}) diag.Diagnostics {
	timeout, err := time.ParseDuration(r.Get("timeout").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	poll, err := time.ParseDuration(r.Get("poll").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	uri, err := url.Parse(r.Get("uri").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	token := r.Get("token").(string)

	client := resty.New()
	if caCert, exists := r.GetOk("cacert"); exists {
		client.SetRootCertificateFromString(caCert.(string))
	}
	if caCertPath, exists := r.GetOk("cacert_path"); exists {
		client.SetRootCertificate(caCertPath.(string))
	}

	startedAt := time.Now()

	for {
		resp, err := client.R().
			SetContext(c).
			SetHeader("Accept", "application/json").
			SetAuthToken(token).
			Get(uri.String())
		if err != nil {
			return diag.FromErr(err)
		}

		var jsonResponse map[string]interface{}

		switch resp.StatusCode() {
		case 200:
			r.SetId(uri.Path)
			return diag.Diagnostics{}

		case 404:
			//do nothing, waiting

		default:
			if err := json.Unmarshal([]byte(resp.String()), &jsonResponse); err != nil {
				return diag.FromErr(err)
			}
			return diag.Diagnostics{diag.Diagnostic{
				Severity: diag.Error,
				Summary:  jsonResponse["message"].(string),
				Detail:   resp.String(),
			}}
		}

		if time.Now().After(startedAt.Add(timeout)) {
			return diag.Diagnostics{diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Timeout reached",
				Detail:   fmt.Sprintf("Path %s is still unavalable after %s", uri.Path, time.Now().Sub(startedAt).String()),
			}}
		}

		time.Sleep(poll)
	}
}

func kubernetesResourceAwaiterRead(c context.Context, r *schema.ResourceData, _ interface{}) diag.Diagnostics {
	uri, err := url.Parse(r.Get("uri").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	token := r.Get("token").(string)

	client := resty.New()
	if caCert, exists := r.GetOk("cacert"); exists {
		client.SetRootCertificateFromString(caCert.(string))
	}
	if caCertPath, exists := r.GetOk("cacert_path"); exists {
		client.SetRootCertificate(caCertPath.(string))
	}

	resp, err := client.R().
		SetContext(c).
		SetHeader("Accept", "application/json").
		SetAuthToken(token).
		Get(uri.String())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	var jsonResponse map[string]interface{}
	switch resp.StatusCode() {
	case 200:
		r.SetId(uri.Path)
	case 404:
		r.SetId("")
	default:
		if err := json.Unmarshal([]byte(resp.String()), &jsonResponse); err != nil {
			return diag.FromErr(err)
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  jsonResponse["message"].(string),
			Detail:   resp.String(),
		})
	}

	return diags
}

func kubernetesResourceAwaiterUpdate(c context.Context, r *schema.ResourceData, i interface{}) diag.Diagnostics {
	return kubernetesResourceAwaiterRead(c, r, i)
}

func kubernetesResourceAwaiterDelete(_ context.Context, r *schema.ResourceData, _ interface{}) diag.Diagnostics {
	r.SetId("")
	return diag.Diagnostics{}
}
