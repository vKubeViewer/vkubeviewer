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
			if vm.Summary.Config.Name == "k8s-worker-02" {
				fmt.Println(len(vm.Network))

				pc := property.DefaultCollector(c)
				var n mo.Network

				err = pc.Retrieve(ctx, vm.Network, nil, &n)
				if err != nil {
					return err
				}

				fmt.Printf("%T\n", n)
				fmt.Println(n.Name)

			}
		}

		return nil
	})
}
