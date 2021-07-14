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
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	topologyv1 "vkubeviewer/api/v1"
)

// TagInfoReconciler reconciles a TagInfo object
type TagInfoReconciler struct {
	client.Client
	VC_vim25 *vim25.Client
	VC_rest  *rest.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=taginfoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=taginfoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=taginfoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TagInfo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *TagInfoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = context.Background()

	// ------------
	// Log Session
	// ------------
	log := r.Log.WithValues("TagInfo", req.NamespacedName)
	taginfo := &topologyv1.TagInfo{}

	// Log Output for failure
	if err := r.Client.Get(ctx, req.NamespacedName, taginfo); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Error(err, "unable to fecth TagInfo")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log Output for sucess
	msg := fmt.Sprintf("received reconcile request for %q (namespace : %q)", taginfo.GetName(), taginfo.GetNamespace())
	log.Info(msg)

	// ------------
	// Retrieve Session
	// ------------

	// tags.NewManager creates a new Manager instance with the rest.Client to retrieve tags information
	tm := tags.NewManager(r.VC_rest)

	// get type tags.Tag with Tagname
	tag, err := tm.GetTag(ctx, taginfo.Spec.Tagname)
	if err != nil {
		msg := fmt.Sprintf("unable to get tags.Tag based on the %s : error %s", taginfo.Spec.Tagname, err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	// list ListAttachedObjects with tag.ID
	objs, err := tm.ListAttachedObjects(ctx, tag.ID)
	if err != nil {
		msg := fmt.Sprintf("unable to list attachedobjects on the tag %s : error %s", taginfo.Spec.Tagname, err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	// retrieve the managedobjects with managedobjectreference by property's Retrieve
	pc := property.DefaultCollector(r.VC_vim25)

	// refmap stores the managedobjectreference based on the managedobject type
	refmap := make(map[string][]types.ManagedObjectReference)
	for _, obj := range objs {
		refmap[obj.Reference().Type] = append(refmap[obj.Reference().Type], obj.Reference())
	}

	// define current ManagedObjects list
	var curDatacenterList []string
	var curClusterList []string
	var curHostList []string
	var curVMList []string
	// store the node list via k8s api
	var k8snode = ListK8sNodes()

	// traverse refmap, according its type, retrieve the managedobject and append the name to the ManagedObjects list
	for key, element := range refmap {
		switch key {
		case "Datacenter":
			var dcs []mo.Datacenter
			err = pc.Retrieve(ctx, element, []string{"name"}, &dcs)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve information of %s : error %s", key, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}
			// store name into list
			for _, dc := range dcs {
				curDatacenterList = append(curDatacenterList, dc.Name)
			}
		case "ClusterComputeResource":
			var ccs []mo.ClusterComputeResource
			err = pc.Retrieve(ctx, element, []string{"name"}, &ccs)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve information of %s : error %s", key, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}
			for _, cc := range ccs {
				curClusterList = append(curClusterList, cc.Name)
			}
		case "HostSystem":
			var hss []mo.HostSystem
			err = pc.Retrieve(ctx, element, []string{"name"}, &hss)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve information of %s : error %s", key, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}
			for _, hs := range hss {
				curHostList = append(curHostList, hs.Name)
			}

		case "VirtualMachine":
			var vms []mo.VirtualMachine
			err = pc.Retrieve(ctx, element, nil, &vms)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve information of %s : error %s", key, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}

			// traverse virtual machines
			for _, vm := range vms {
				// RPref - resourcePool Reference
				RPref := vm.ResourcePool
				var resourcepool mo.ResourcePool
				err = pc.RetrieveOne(ctx, *RPref, nil, &resourcepool)
				if err != nil {
					msg := fmt.Sprintf("unable to retrieve ResourcePool MO of %s : error %s", RPref.Value, err)
					log.Info(msg)
					return ctrl.Result{}, err
				}

				// RPref - ClusterComputeResource Reference
				CCRref := resourcepool.Parent
				var clustercomputeresource mo.ClusterComputeResource
				err = pc.RetrieveOne(ctx, *CCRref, nil, &clustercomputeresource)
				if err != nil {
					msg := fmt.Sprintf("unable to retrieve ClusterComputeResource MO of %s : error %s", CCRref.Value, err)
					log.Info(msg)
					return ctrl.Result{}, err
				}

				// check whether virtual machine is a k8s node or not
				if !stringInSlice(vm.Name, k8snode) {
					str := []string{vm.Name, "[", clustercomputeresource.Name, "]"}
					curVMList = append(curVMList, strings.Join(str, " "))
				} else {
					// if the vm is a k8s node, add marker "k8s"
					str := []string{"k8s", vm.Name, "[", clustercomputeresource.Name, "]"}
					curVMList = append(curVMList, strings.Join(str, " "))
				}
			}

		}
	}
	// if current Lists are different from the one stored in status, replace them
	taginfo.Status.DatacenterList = UpdateStatus(curDatacenterList, taginfo.Status.DatacenterList)
	taginfo.Status.ClusterList = UpdateStatus(curClusterList, taginfo.Status.ClusterList)
	taginfo.Status.HostList = UpdateStatus(curHostList, taginfo.Status.HostList)
	taginfo.Status.VMList = UpdateStatus(curVMList, taginfo.Status.VMList)
	// ------------
	// Update Session
	// ------------

	// update the status
	if err := r.Status().Update(ctx, taginfo); err != nil {
		log.Error(err, "unable to update TagInfo status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: time.Duration(1) * time.Minute,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TagInfoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&topologyv1.TagInfo{}).
		Complete(r)
}
