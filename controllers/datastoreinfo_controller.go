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
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	topologyv1 "vkubeviewer/api/v1"
)

// DatastoreInfoReconciler reconciles a DatastoreInfo object
type DatastoreInfoReconciler struct {
	client.Client
	VC     *vim25.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=datastoreinfoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=datastoreinfoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=topology.vkubeviewer.com,resources=datastoreinfoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatastoreInfo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *DatastoreInfoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = context.Background()
	// ------------
	// Log Session
	// ------------
	log := r.Log.WithValues("DatastoreInfo", req.NamespacedName)
	dsinfo := &topologyv1.DatastoreInfo{}

	// Log Output for failure
	if err := r.Client.Get(ctx, req.NamespacedName, dsinfo); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Error(err, "unable to fecth DatastoreInfo")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log Output for sucess
	msg := fmt.Sprintf("received reconcile request for %q (namespace : %q)", dsinfo.GetName(), dsinfo.GetNamespace())
	log.Info(msg)

	// ------------
	// Retrieve Session
	// ------------

	// Create a view manager

	m := view.NewManager(r.VC)

	// Create a container view of Datastore objects
	// vds - viewer of datastore
	vds, err := m.CreateContainerView(ctx, r.VC.ServiceContent.RootFolder, []string{"Datastore"}, true)

	if err != nil {
		msg := fmt.Sprintf("unable to create container view for Datastore: error %s", err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	defer vds.Destroy(ctx)

	// Retrieve DS information for all DSs
	// dss - datastores
	var dss []mo.Datastore

	err = vds.Retrieve(ctx, []string{"Datastore"}, nil, &dss)
	if err != nil {
		msg := fmt.Sprintf("unable to retrieve Datastore info: error %s", err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	// traverse all the DSs
	for _, ds := range dss {
		// if it's the DS we're looking for
		if ds.Summary.Name == dsinfo.Spec.Datastore {

			// Store info into the status
			dsinfo.Status.Type = ds.Summary.Type
			dsinfo.Status.Status = string(ds.OverallStatus)
			dsinfo.Status.Capacity = ds.Summary.Capacity
			dsinfo.Status.FreeSpace = ds.Summary.FreeSpace
			dsinfo.Status.Accessible = ds.Summary.Accessible

			// get the Hosts attached to this datastore, type []types.DatastoreHostMount
			HostMounts := ds.Host

			if len(dsinfo.Status.HostsMounted) != len(HostMounts) {
				dsinfo.Status.HostsMounted = nil
				// traverse all the HostMount
				for _, HostMount := range HostMounts {

					// get the Host info
					var h mo.HostSystem
					pc := property.DefaultCollector(r.VC)
					err = pc.RetrieveOne(ctx, HostMount.Key, nil, &h)
					if err != nil {
						msg := fmt.Sprintf("unable to retrieve HostSystem info: error %s", err)
						log.Info(msg)
						return ctrl.Result{}, err
					}
					// append the Host's Name into Hosts List
					dsinfo.Status.HostsMounted = append(dsinfo.Status.HostsMounted, h.Summary.Config.Name)
				}
			}

		}
	}

	// ------------
	// Update Session
	// ------------

	// update the status
	if err := r.Status().Update(ctx, dsinfo); err != nil {
		log.Error(err, "unable to update VMInfo status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatastoreInfoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&topologyv1.DatastoreInfo{}).
		Complete(r)
}
