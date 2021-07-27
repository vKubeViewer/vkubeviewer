# Node Information 

This document contains information on the [Node Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/nodeinfo_controller.go) controller.

## Retrievable Information 

### Node Infromation 

- Array containing all tags associated with the node
- The VM Guest ID of the node
- Total CPU usage of the node
- Total Reserved CPU Capacity of the node
- Total Memory usage of the node
- Total Reserved Memory Capacity of the node
- The power state of the node
- The Hw version of the node
- The IP address of the node
- The path to the node.
- Related cluster
- Related host

### Node Network Infromation 

- Network Name
- Network Status
- Network Switch Type
- Network Vlan ID




## API to Retrieve Information

The above fields are populated via API calls to the vShere server via [Node Information API](https://github.com/vKubeViewer/vkubeviewer/blob/main/api/v1/nodeinfo_types.go) as follows

### Set Target Node Name

Target node name is set in the API call via the following:

```
type NodeInfoSpec struct {
	Nodename string `json:"nodename"`
}

```

### Populate Information of Target Node

Information from the targeted node is populated from the API as follows

```

type NodeInfoStatus struct {
	// cpu, memory, vmipaddress, powerstate
	ActtachedTag []string `json:"acttached_tag,omitempty"`
	VMGuestId    string   `json:"vm_guest_id,omitempty"`
	VMTotalCPU   int64    `json:"vm_total_cpu,omitempty"`
	VMResvdCPU   int64    `json:"vm_resvd_cpu,omitempty"`
	VMTotalMem   int64    `json:"vm_total_mem,omitempty"`
	VMResvdMem   int64    `json:"vm_resvd_mem,omitempty"`
	VMPowerState string   `json:"vm_power_state,omitempty"`
	VMHwVersion  string   `json:"vm_hw_version,omitempty"`
	VMIpAddress  string   `json:"vm_ip_address,omitempty"`
	PathToVM     string   `json:"path_to_vm,omitempty"`

	NetName          string `json:"net_name,omitempty"`
	NetOverallStatus string `json:"net_overall_status,omitempty"`
	NetSwitchType    string `json:"net_switch_type,omitempty"`
	NetVlanId        int32  `json:"net_vlan_id,omitempty"`
	RelatedCluster   string `json:"related_cluster,omitempty"`
	RelatedHost      string `json:"related_host,omitempty"`
}

```


## Controller Populates relevent YAML file

Once the API has retrived the requested Information, this is then passed to the [controller](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/nodeinfo_controller.go) where it is sent to the relevent YAML file for subsiquent output:
:

```
// Create a container view of VirtualMachine objects
	// vvm - viewer of virtual machine
	vvm, err := m.CreateContainerView(ctx, r.VC_vim25.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)

	if err != nil {
		msg := fmt.Sprintf("unable to create container view for VirtualMachines: error %s", err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	defer vvm.Destroy(ctx)

	// Retrieve all property for all VMs
	// vms - VirtualMachines
	var vms []mo.VirtualMachine

	err = vvm.Retrieve(ctx, []string{"VirtualMachine"}, nil, &vms)

	if err != nil {
		msg := fmt.Sprintf("unable to retrieve VM infomartion: error %s", err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	// tags.NewManager creates a new Manager instance with the rest.Client to retrieve tags information
	tm := tags.NewManager(r.VC_rest)

	// traverse all the VM
	for _, vm := range vms {
		// if the VM's name equals to Nodename
		if vm.Summary.Config.Name == node.Spec.Nodename {
			// get attachedtags on this virtual machine
			tags, err := tm.GetAttachedTags(ctx, vm.Self)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve tags on %s : error %s", vm.Summary.Config.Name, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}

			// store the attachedtags info in status
			var curTags []string
			for _, tag := range tags {
				curTags = append(curTags, tag.Name)
			}

			node.Status.ActtachedTag = UpdateStatusList(node.Status.ActtachedTag, curTags)

			// store VM information in status
			node.Status.VMGuestId = string(vm.Summary.Guest.GuestId)
			node.Status.VMTotalCPU = int64(vm.Summary.Config.NumCpu)
			node.Status.VMResvdCPU = int64(vm.Summary.Config.CpuReservation)
			node.Status.VMTotalMem = int64(vm.Summary.Config.MemorySizeMB)
			node.Status.VMResvdMem = int64(vm.Summary.Config.MemoryReservation)
			node.Status.VMPowerState = string(vm.Summary.Runtime.PowerState)
			node.Status.VMHwVersion = string(vm.Summary.Guest.HwVersion)
			node.Status.VMIpAddress = string(vm.Summary.Guest.IpAddress)
			node.Status.PathToVM = string(vm.Summary.Config.VmPathName)

			// retrieve related host info
			hostref := vm.Runtime.Host
			pc := property.DefaultCollector(r.VC_vim25)
			var host mo.HostSystem
			err = pc.RetrieveOne(ctx, *hostref, []string{"name", "parent"}, &host)
			if err != nil {
				msg = fmt.Sprintf("unable to retrieve RelatedHost: error %s", err)
				log.Info(msg)
				return ctrl.Result{}, err
			}

			node.Status.RelatedHost = host.Name

			// retrieve related cluster info
			clusterref := host.Parent
			pc = property.DefaultCollector(r.VC_vim25)
			var clustercomputeresource mo.ClusterComputeResource
			err = pc.RetrieveOne(ctx, *clusterref, []string{"name"}, &clustercomputeresource)
			if err != nil {
				msg = fmt.Sprintf("unable to retrieve RelatedHost: error %s", err)
				log.Info(msg)
				return ctrl.Result{}, err
			}

			node.Status.RelatedCluster = clustercomputeresource.Name

			// traverse the network, in our operator, we consider only single network
			for _, ref := range vm.Network {
				if ref.Type == "Network" {
					// if it's a normal Network, define the n as DistributedVirtualPortgroup mo.Network
					var n mo.Network
					node.Status.NetSwitchType = "Standard"

					// a property collector to retrieve objects by MOR
					err = pc.Retrieve(ctx, vm.Network, nil, &n)
					if err != nil {
						msg = fmt.Sprintf("unable to retrieve VM Network: error %s", err)
						log.Info(msg)
						return ctrl.Result{}, err
					}

					// store the info in the status
					node.Status.NetName = string(n.Name)
					node.Status.NetOverallStatus = string(n.OverallStatus)
				} else if ref.Type == "DistributedVirtualPortgroup" {

					// if it's a distributed network, define the n as mo.DistributedVirtualPortgroup
					var pg mo.DistributedVirtualPortgroup
					node.Status.NetSwitchType = "Distributed"

					// a property collector to retrieve objects by MOR
					err = pc.Retrieve(ctx, vm.Network, nil, &pg)
					if err != nil {
						msg = fmt.Sprintf("unable to retrieve VM DVPortGroup: error %s", err)
						log.Info(msg)
						return ctrl.Result{}, err
					}

					// store the info in the status
					node.Status.NetName = string(pg.Name)
					node.Status.NetOverallStatus = string(pg.OverallStatus)

					// get vlanID - more examples to use type assertion to access extended types in govmomi
					// - https://github.com/vKubeViewer/vkubeviewer/blob/main/ctrl_dev/getvlanid/main.go
					portConfig := pg.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)
					vlan := portConfig.Vlan.(*types.VmwareDistributedVirtualSwitchVlanIdSpec)
					node.Status.NetVlanId = vlan.VlanId

				}
			}

		}
	}

```