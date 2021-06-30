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
		// Create view of VirtualMachine objects
		m := view.NewManager(c)

		vvm, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
		if err != nil {
			return err
		}

		defer vvm.Destroy(ctx)

		// Retrieve summary property for all machines
		// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
		var vms []mo.VirtualMachine

		err = vvm.Retrieve(ctx, []string{"VirtualMachine"}, nil, &vms)
		if err != nil {
			return err
		}

		// Print summary per vm (see also: govc/vm/info.go)

		for _, vm := range vms {
			if vm.Summary.Config.Name == "ucc-demo" {
				//if vm.Summary.Config.Name == "k8s-worker-02" {
				fmt.Printf("%T, %v\n", vm.Network, vm.Network)

				for _, v := range vm.Network {
					fmt.Printf("%T, %v\n", v, v)
					fmt.Println(v.Type)

					var Switchtype string
					var Overallstatus string
					var Netname string
					//var Vlanid int64

					if v.Type == "Network" {
						var n mo.Network
						Switchtype = "Standard"

						pc := property.DefaultCollector(c)
						err = pc.Retrieve(ctx, vm.Network, nil, &n)
						if err != nil {
							return err
						}
						Netname = n.Name
						Overallstatus = string(n.OverallStatus)

					} else if v.Type == "DistributedVirtualPortgroup" {
						fmt.Println("get here")
						var n mo.DistributedVirtualPortgroup
						Switchtype = "Distributed"

						pc := property.DefaultCollector(c)
						err = pc.Retrieve(ctx, vm.Network, nil, &n)
						if err != nil {
							return err
						}
						Netname = n.Name
						Overallstatus = string(n.OverallStatus)

						// get vlanid
						var dvs mo.DistributedVirtualSwitch
						err = pc.RetrieveOne(ctx, *n.Config.DistributedVirtualSwitch, nil, &dvs)
						if err != nil {
							return err
						}
						fmt.Printf("%s", dvs.Uuid)
						//test := dvs.Config.GetDVSConfigInfo()
						//fmt.Printf("%s, %T\n", test, test)
						//test = test.DefaultPortConfig.GetDVPortSetting()
						//fmt.Printf("%s, %T\n", test, test)

					}

					fmt.Printf("%s %s %s\n", Switchtype, Overallstatus, Netname)
				}

			}
		}

		return nil
	})
}
