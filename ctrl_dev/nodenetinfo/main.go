package main

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/examples"
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
				fmt.Printf("%s\n", vm.Summary.Config.Name)
				fmt.Printf("%T\n", vm.Network)
				fmt.Println(len(vm.Network))
				for n := range vm.Network {
					fmt.Printf("%T\n", n)
					fmt.Println(n)
				}
			}
		}

		return nil
	})
}
