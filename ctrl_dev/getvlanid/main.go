package main

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/examples"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func main() {
	// the code is an example to show how to use type assertion to access extened types
	// that extend the base type in govmomi
	examples.Run(func(ctx context.Context, c *vim25.Client) error {
		m := view.NewManager(c)
		v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"DistributedVirtualSwitch"}, true)
		if err != nil {
			return err
		}
		defer v.Destroy(ctx)
		var dvs []mo.DistributedVirtualSwitch
		err = v.Retrieve(ctx, []string{"DistributedVirtualSwitch"}, nil, &dvs)
		if err != nil {
			return err
		}
		for _, s := range dvs {
			// VMwareDVSConfigInfor is an extended DVSConfigInfo
			config := s.Config.(*types.VMwareDVSConfigInfo)
			// VMwareDVSPortSetting is an extended DVPortSetting
			portConfig := config.DefaultPortConfig.(*types.VMwareDVSPortSetting)
			// VmwareDistributedVirtualSwitchVlanIdSpec is an extended VmwareDistributedVirtualSwitchVlanSpec
			vlan := portConfig.Vlan.(*types.VmwareDistributedVirtualSwitchVlanIdSpec)
			fmt.Printf("vlan id=%d\n", vlan.VlanId)
		}

		return err
	})
}
