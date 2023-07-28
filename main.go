package main

import (
	"flag"
	"fmt"
	"github.com/a0s/terraform-provider-kubernetes-awaiter/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/octago/sflags/gen/gflag"
	"os"
)

const VERSION = "0.0.2"

type Config struct {
	Version bool `flag:"version" desc:"show version"`
}

func main() {
	config := &Config{}

	err := gflag.ParseToDef(config)
	if err != nil {
		panic(err.Error())
	}
	flag.Parse()

	if config.Version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return provider.Provider()
		},
	})
}
