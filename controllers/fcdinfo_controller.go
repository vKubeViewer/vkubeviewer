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
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vslm"
	vslmtypes "github.com/vmware/govmomi/vslm/types"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	topologyv1 "vkubeviewer/api/v1"
)

// FCDInfoReconciler reconciles a FCDInfo object
type FCDInfoReconciler struct {
	client.Client
	VC     *vim25.Client
	Finder *find.Finder
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=fcdinfoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=fcdinfoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=fcdinfoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FCDInfo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *FCDInfoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = context.Background()
	// ------------
	// Log Session
	// ------------
	log := r.Log.WithValues("FCDInfo", req.NamespacedName)
	fcd := &topologyv1.FCDInfo{}

	// Log Output for failure
	if err := r.Client.Get(ctx, req.NamespacedName, fcd); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Error(err, "unable to fetch FCDInfo")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log Output for sucess
	msg := fmt.Sprintf("received reconcile request for %q (namespace: %q)", fcd.GetName(), fcd.GetNamespace())
	log.Info(msg)

	// ------------
	// Retrieve Session
	// ------------

	// connect to the vslm client
	vslmClient, _ := vslm.NewClient(ctx, r.VC)

	// retrieve vstorageID
	m := vslm.NewGlobalObjectManager(vslmClient)
	var query []vslmtypes.VslmVsoVStorageObjectQuerySpec

	spec := vslmtypes.VslmVsoVStorageObjectQuerySpec{
		QueryField:    "name",
		QueryOperator: "equals",
		QueryValue:    []string{fcd.Spec.PVId},
	}
	query = append(query, spec)

	result, _ := m.ListObjectsForSpec(ctx, query, 1000)
	vstorageIDs := result.Id

	var vstorageobject *types.VStorageObject

	// retrieve vstorage objects
	for _, vstorageID := range vstorageIDs {
		vstorageobject, _ = m.Retrieve(ctx, vstorageID)

		if vstorageobject.Config.BaseConfigInfo.Name == fcd.Spec.PVId {
			msg := fmt.Sprintf("FCDInfo: %v matches %v", vstorageobject.Config.BaseConfigInfo.Name, fcd.Spec.PVId)
			log.Info(msg)

			// store information into FCDInfo's status
			fcd.Status.SizeMB = int64(vstorageobject.Config.CapacityInMB)
			backing := vstorageobject.Config.BaseConfigInfo.Backing.(*types.BaseConfigInfoDiskFileBackingInfo)
			fcd.Status.FilePath = string(backing.FilePath)
			fcd.Status.ProvisioningType = string(backing.ProvisioningType)
		}
	}

	// ------------
	// Update Session
	// ------------
	if err := r.Status().Update(ctx, fcd); err != nil {
		log.Error(err, "unable to update FCDInfo status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Duration(30) * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FCDInfoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&topologyv1.FCDInfo{}).
		Complete(r)
}
