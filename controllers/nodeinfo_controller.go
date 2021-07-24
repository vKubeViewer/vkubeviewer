/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	topologyv1 "vkubeviewer/api/v1"
)

// NodeInfoReconciler reconciles a NodeInfo object
type NodeInfoReconciler struct {
	client.Client
	VC_vim25 *vim25.Client
	VC_rest  *rest.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=nodeinfoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=nodeinfoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=nodeinfoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeInfo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *NodeInfoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	ctx = context.Background()
	// ------------
	// Log Session
	// ------------
	log := r.Log.WithValues("NodeInfo", req.NamespacedName)

	node := &topologyv1.NodeInfo{}

	// Log Output for failure
	if err := r.Client.Get(ctx, req.NamespacedName, node); err != nil {
		// add some debug information if it's not a NotFound error
		if !k8serr.IsNotFound(err) {
			log.Error(err, "unable to fetch NodeInfo")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log Output for sucess
	msg := fmt.Sprintf("received reconcile request for %q (namespace: %q)", node.GetName(), node.GetNamespace())
	log.Info(msg)

	// ------------
	// Retrieve Session
	// ------------

	// Create a view manager
	m := view.NewManager(r.VC_vim25)

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
				msg = fmt.Sprintf("unable to retrieve Related Compute Cluster: error %s", err)
				log.Info(msg)
				return ctrl.Result{}, err
			}

			node.Status.RelatedCluster = clustercomputeresource.Name

			// retrive related datastore info
			datastoreref := vm.Datastore
			pc = property.DefaultCollector(r.VC_vim25)
			var datastores []mo.Datastore
			err = pc.Retrieve(ctx, datastoreref, []string{"name"}, &datastores)
			if err != nil {
				msg = fmt.Sprintf("unable to retrieve Related Datastore: error %s", err)
				log.Info(msg)
				return ctrl.Result{}, err
			}

			var curDatastores []string
			for _, datastore := range datastores {
				curDatastores = append(curDatastores, datastore.Name)
			}
			node.Status.RelatedDatastore = curDatastores

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

	// ------------
	// Update Session
	// ------------

	// update the status
	if err := r.Status().Update(ctx, node); err != nil {
		log.Error(err, "unable to update NodeInfo status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Duration(1) * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeInfoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&topologyv1.NodeInfo{}).
		Complete(r)
}
