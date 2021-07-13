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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	topologyv1 "vkubeviewer/api/v1"
)

var k8snode []string

func ListK8sNodes() []string {
	var curK8sNode []string
	var kubeconfig *string
	path := homedir.HomeDir() + "/.kube/config"
	kubeconfig = &path
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nodeList := clientSet.CoreV1().Nodes()
	nodes, err := nodeList.List(context.TODO(), v1.ListOptions{})

	if err != nil {
		fmt.Println("Error occurred: ", err)
	}

	for _, item := range nodes.Items {
		curK8sNode = append(curK8sNode, item.ObjectMeta.Name)
	}
	return curK8sNode
}

func stringInSlice(s string, list []string) bool {
	for _, ele := range list {
		if s == ele {
			return true
		}
	}
	return false
}

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

	tm := tags.NewManager(r.VC_rest)
	tag, _ := tm.GetTag(ctx, taginfo.Spec.Tagname)
	refmap := make(map[string][]types.ManagedObjectReference)
	objs, _ := tm.ListAttachedObjects(ctx, tag.ID)
	pc := property.DefaultCollector(r.VC_vim25)
	for _, obj := range objs {
		// fmt.Println(obj.Reference().Type, obj.Reference().Value)
		refmap[obj.Reference().Type] = append(refmap[obj.Reference().Type], obj.Reference())
	}
	var curDatacenterList []string
	var curClusterList []string
	var curHostList []string
	var curVMList []string

	k8snode = ListK8sNodes()

	for key, element := range refmap {
		switch key {

		case "Datacenter":
			var dcs []mo.Datacenter
			_ = pc.Retrieve(ctx, element, []string{"name"}, &dcs)
			for _, dc := range dcs {
				curDatacenterList = append(curDatacenterList, dc.Name)
			}
			if !ArrayEqual(curDatacenterList, taginfo.Status.DatacenterList) {
				taginfo.Status.DatacenterList = curDatacenterList
			}

		case "ClusterComputeResource":
			var ccs []mo.ClusterComputeResource
			_ = pc.Retrieve(ctx, element, []string{"name"}, &ccs)
			for _, cc := range ccs {
				curClusterList = append(curClusterList, cc.Name)
			}
			if !ArrayEqual(curClusterList, taginfo.Status.ClusterList) {
				taginfo.Status.ClusterList = curClusterList
			}
		case "HostSystem":
			var hss []mo.HostSystem
			_ = pc.Retrieve(ctx, element, []string{"name"}, &hss)
			// fmt.Println(hss)
			for _, hs := range hss {
				curHostList = append(curHostList, hs.Name)
			}
			if !ArrayEqual(curHostList, taginfo.Status.HostList) {
				taginfo.Status.HostList = curHostList
			}
		case "VirtualMachine":
			var vms []mo.VirtualMachine
			_ = pc.Retrieve(ctx, element, nil, &vms)

			for _, vm := range vms {
				RPref := vm.ResourcePool
				var resourcepool mo.ResourcePool
				_ = pc.RetrieveOne(ctx, *RPref, nil, &resourcepool)
				CCRref := resourcepool.Parent
				var clustercomputeresource mo.ClusterComputeResource
				_ = pc.RetrieveOne(ctx, *CCRref, nil, &clustercomputeresource)

				if !stringInSlice(vm.Name, k8snode) {
					curVMList = append(curVMList, vm.Name)
				} else {
					str := []string{"k8s", vm.Name, "[", clustercomputeresource.Name, "]"}
					curVMList = append(curVMList, strings.Join(str, " "))
				}
			}
			if !ArrayEqual(curVMList, taginfo.Status.VMList) {
				taginfo.Status.VMList = curVMList
			}
		}
	}
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
