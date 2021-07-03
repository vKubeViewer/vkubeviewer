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

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/property"
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

// NodeNetInfoReconciler reconciles a NodeNetInfo object
type NodeNetInfoReconciler struct {
	client.Client
	VC     *vim25.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=nodenetinfoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=nodenetinfoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=nodenetinfoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeNetInfo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *NodeNetInfoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = context.Background()
	// ------------
	// Log Session
	// ------------
	log := r.Log.WithValues("NodeNetInfo", req.NamespacedName)
	net := &topologyv1.NodeNetInfo{}

	// Log Output for failure
	if err := r.Client.Get(ctx, req.NamespacedName, net); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Error(err, "unable to fecth NodeNetInfo")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log Output for sucess
	msg := fmt.Sprintf("received reconcile request for %q (namespace : %q)", net.GetName(), net.GetNamespace())
	log.Info(msg)

	// ------------
	// Retrieve Session
	// ------------

	// Create a view manager

	m := view.NewManager(r.VC)

	// Create a container view of VirtualMachine objects

	vvm, err := m.CreateContainerView(ctx, r.VC.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)

	if err != nil {
		msg := fmt.Sprintf("unable to create container view for VirtualMachines: error %s", err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	defer vvm.Destroy(ctx)

	// Retrieve VM information for all VMs

	var vms []mo.VirtualMachine

	err = vvm.Retrieve(ctx, []string{"VirtualMachine"}, nil, &vms)
	if err != nil {
		msg := fmt.Sprintf("unable to retrieve VM info: error %s", err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	// traverse all the VM
	for _, vm := range vms {
		// if the VM's name equals to Nodename
		if vm.Summary.Config.Name == net.Spec.Nodename {
			// traverse the network, in our operator, we consider only single network
			for _, ref := range vm.Network {
				if ref.Type == "Network" {
					// if it's a normal Network, define the n as DistributedVirtualPortgroup mo.Network
					var n mo.Network
					net.Status.SwitchType = "Standard"

					// a property collector to retrieve objects by MOR
					pc := property.DefaultCollector(r.VC)
					err = pc.Retrieve(ctx, vm.Network, nil, &n)
					if err != nil {
						msg = fmt.Sprintf("unable to retrieve VM Network: error %s", err)
						log.Info(msg)
						return ctrl.Result{}, err
					}

					// store the info in the status
					net.Status.NetName = string(n.Name)
					net.Status.NetOverallStatus = string(n.OverallStatus)
				} else if ref.Type == "DistributedVirtualPortgroup" {
					// if it's a distributed network, define the n as mo.DistributedVirtualPortgroup
					var pg mo.DistributedVirtualPortgroup
					net.Status.SwitchType = "Distributed"

					// a property collector to retrieve objects by MOR
					pc := property.DefaultCollector(r.VC)
					err = pc.Retrieve(ctx, vm.Network, nil, &pg)
					if err != nil {
						msg = fmt.Sprintf("unable to retrieve VM DVPortGroup: error %s", err)
						log.Info(msg)
						return ctrl.Result{}, err
					}

					// store the info in the status
					net.Status.NetName = string(pg.Name)
					net.Status.NetOverallStatus = string(pg.OverallStatus)

					// get vlanID
					portConfig := pg.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)
					vlan := portConfig.Vlan.(*types.VmwareDistributedVirtualSwitchVlanIdSpec)
					net.Status.VlanId = vlan.VlanId

				}
			}

		}
	}

	// ------------
	// Update Session
	// ------------

	// update the status
	if err := r.Status().Update(ctx, net); err != nil {
		log.Error(err, "unable to update VMInfo status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeNetInfoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&topologyv1.NodeNetInfo{}).
		Complete(r)
}
