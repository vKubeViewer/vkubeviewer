package main

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/examples"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

func main() {
	examples.Run(func(ctx context.Context, c *vim25.Client) error {
		m := view.NewManager(c)
		fmt.Println("get here 1")
		// vds - viewer of datastore
		vds, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
		if err != nil {
			return err
		}
		fmt.Println("get here 1")
		defer vds.Destroy(ctx)

		// dss - datastores
		var dss []mo.Datastore

		err = vds.Retrieve(ctx, []string{"Datastore"}, nil, &dss)
		if err != nil {
			return err
		}

		for _, ds := range dss {
			if ds.Summary.Name == "vsan-OCTO-Cluster-A" {
				fmt.Println("find the datastore")

				var Type string
				var Status string
				var Capacity int64
				var FreeSpace int64
				var Accessible bool
				var Hosts []string

				Type = ds.Summary.Type
				Status = string(ds.OverallStatus)
				Capacity = ds.Summary.Capacity
				FreeSpace = ds.Summary.FreeSpace
				Accessible = ds.Summary.Accessible

				HostMounts := ds.Host

				for _, HostMount := range HostMounts {
					var h mo.HostSystem
					pc := property.DefaultCollector(c)
					err = pc.RetrieveOne(ctx, HostMount.Key, nil, &h)
					if err != nil {
						return err
					}
					Hosts = append(Hosts, h.Summary.Config.Name)
				}

				fmt.Printf("%s, %s, %d, %d, %v\n", Type, Status, Capacity, FreeSpace, Accessible)
				fmt.Println(Hosts)
			}

		}
		return nil
	})
}
