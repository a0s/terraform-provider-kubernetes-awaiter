package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap:   map[string]*schema.Resource{"kubernetes_resource_awaiter": kubernetesResourceAwaiter()},
		DataSourcesMap: map[string]*schema.Resource{},
		Schema:         map[string]*schema.Schema{},
	}
}
