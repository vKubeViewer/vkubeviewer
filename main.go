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

package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	topologyv1 "vkubeviewer/api/v1"
	"vkubeviewer/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(topologyv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// createDatastoreInfo creates DatastoreInfo CR in current k8s cluster
func createDatastoreInfo(ctx context.Context, mgr manager.Manager, vim25client *vim25.Client) error {
	m := view.NewManager(vim25client)
	vds, err := m.CreateContainerView(ctx, vim25client.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		msg := fmt.Sprintf("unable to create container view for Datastore: error %s", err)
		setupLog.Info(msg)
	} else {
		msg := fmt.Sprintf("succeed to create container view for Datastore")
		setupLog.Info(msg)
	}
	defer vds.Destroy(ctx)

	var dss []mo.Datastore

	err = vds.Retrieve(ctx, []string{"Datastore"}, []string{"name"}, &dss)
	if err != nil {
		msg := fmt.Sprintf("unable to retrieve Datastore info: error %s", err)
		setupLog.Info(msg)
	} else {
		msg := fmt.Sprintf("succeed to retrieve Datastore info")
		setupLog.Info(msg)
	}

	// ------------
	// Create DatastoreInfo with K8s CRD
	// ------------

	c := mgr.GetClient()

	for _, ds := range dss {
		datastore := &topologyv1.DatastoreInfo{
			TypeMeta:   metav1.TypeMeta{Kind: "DatastoreInfo", APIVersion: "topology.vkubeviewer.com/v1"},
			ObjectMeta: metav1.ObjectMeta{Name: strings.ToLower(ds.Name), Namespace: "default"},
			Spec:       topologyv1.DatastoreInfoSpec{Datastore: ds.Name},
			Status:     topologyv1.DatastoreInfoStatus{},
		}

		if err := c.Create(ctx, datastore); err != nil {

			setupLog.Error(err, "unable to create Datastore")
		} else {
			msg := fmt.Sprintf("Create DatastoreInfo object %s", ds.Name)
			setupLog.Info(msg)
		}

	}
	return err
}

// createFCDInfo creates FCDInfo CR in current k8s cluster
func createFCDInfo(ctx context.Context, mgr manager.Manager, vim25client *vim25.Client) error {
	var k8spv = controllers.ListK8sPV()
	var err error
	c := mgr.GetClient()
	for index, fcd := range k8spv {
		fcdindo := &topologyv1.FCDInfo{
			TypeMeta:   metav1.TypeMeta{Kind: "FCDInfo", APIVersion: "topology.vkubeviewer.com/v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "fcd" + strconv.Itoa(index), Namespace: "default"},
			Spec:       topologyv1.FCDInfoSpec{PVId: fcd},
			Status:     topologyv1.FCDInfoStatus{},
		}

		if err := c.Create(ctx, fcdindo); err != nil {

			setupLog.Error(err, "unable to create FCDInfo object")
		} else {
			msg := fmt.Sprintf("Create FCDInfo object %s", fcd)
			setupLog.Info(msg)
		}

	}
	return err
}

// createHostInfo creates HostInfo CR in current k8s cluster
func createHostInfo(ctx context.Context, mgr manager.Manager, vim25client *vim25.Client) error {
	m := view.NewManager(vim25client)
	vhosts, err := m.CreateContainerView(ctx, vim25client.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		msg := fmt.Sprintf("unable to create container view for HostSystem: error %s", err)
		setupLog.Info(msg)
	} else {
		msg := fmt.Sprintf("succeed to create container view for HostSystem")
		setupLog.Info(msg)
	}
	defer vhosts.Destroy(ctx)

	var hosts []mo.HostSystem

	err = vhosts.Retrieve(ctx, []string{"HostSystem"}, []string{"name"}, &hosts)
	if err != nil {
		msg := fmt.Sprintf("unable to retrieve HostSystem info: error %s", err)
		setupLog.Info(msg)
	} else {
		msg := fmt.Sprintf("succeed to retrieve HostSystem info")
		setupLog.Info(msg)
	}

	// ------------
	// Create DatastoreInfo with K8s CRD
	// ------------

	c := mgr.GetClient()

	for _, host := range hosts {
		hostinfo := &topologyv1.HostInfo{
			TypeMeta:   metav1.TypeMeta{Kind: "HostInfo", APIVersion: "topology.vkubeviewer.com/v1"},
			ObjectMeta: metav1.ObjectMeta{Name: strings.ToLower(host.Name), Namespace: "default"},
			Spec:       topologyv1.HostInfoSpec{Hostname: host.Name},
			Status:     topologyv1.HostInfoStatus{},
		}

		if err := c.Create(ctx, hostinfo); err != nil {

			setupLog.Error(err, "unable to create Host")
		} else {
			msg := fmt.Sprintf("Create HostInfo object %s", host.Name)
			setupLog.Info(msg)
		}

	}
	return err
}

// createNodeInfo creates NodeInfo CR in current k8s cluster
func createNodeInfo(ctx context.Context, mgr manager.Manager, vim25client *vim25.Client) error {
	var k8snodes = controllers.ListK8sNodes()

	// ------------
	// Create DatastoreInfo with K8s CRD
	// ------------

	c := mgr.GetClient()
	var err error
	for _, n := range k8snodes {
		node := &topologyv1.NodeInfo{
			TypeMeta:   metav1.TypeMeta{Kind: "NodeInfo", APIVersion: "topology.vkubeviewer.com/v1"},
			ObjectMeta: metav1.ObjectMeta{Name: strings.ToLower(n), Namespace: "default"},
			Spec:       topologyv1.NodeInfoSpec{Nodename: n},
			Status:     topologyv1.NodeInfoStatus{},
		}

		if err := c.Create(ctx, node); err != nil {
			setupLog.Error(err, "unable to create Node")
			return err
		} else {
			msg := fmt.Sprintf("Create NodeInfo object %s", n)
			setupLog.Info(msg)
		}

	}

	return err
}

// createTagInfo creates TagInfo CR in current k8s cluster
func createTagInfo(ctx context.Context, mgr manager.Manager, restclient *rest.Client) error {
	tm := tags.NewManager(restclient)

	// ------------
	// Create DatastoreInfo with K8s CRD
	// ------------
	tags, err := tm.GetTags(ctx)
	if err != nil {
		msg := fmt.Sprintf("unable to retrieve Tag Lists : error %s", err)
		setupLog.Info(msg)
	} else {
		msg := fmt.Sprintf("succeed to retrieve Tag Lists")
		setupLog.Info(msg)
	}

	c := mgr.GetClient()

	for index, tag := range tags {
		tag := &topologyv1.TagInfo{
			TypeMeta:   metav1.TypeMeta{Kind: "TagInfo", APIVersion: "topology.vkubeviewer.com/v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "tag" + strconv.Itoa(index), Namespace: "default"},
			Spec:       topologyv1.TagInfoSpec{Tagname: tag.Name},
			Status:     topologyv1.TagInfoStatus{},
		}

		if err := c.Create(ctx, tag); err != nil {

			setupLog.Error(err, "unable to create TagInfo object")
		} else {
			msg := fmt.Sprintf("Create TagInfo object %s", tag.Name)
			setupLog.Info(msg)
		}

	}
	return err
}

// vlogin logins with vSphere with two Clients we use in the operator
// vim25.Client, the most usable one, connects SOAP service
// rest.Client extends soap.Client to support JSON encoding
func vlogin(ctx context.Context, vc, user, pwd string) (*vim25.Client, *rest.Client, error) {

	// This section allows for insecure govmomi logins
	var insecure bool
	flag.BoolVar(&insecure, "insecure", true, "ignore any vCenter TLS cert validation error")

	// Create a vSphere/vCenter client
	// The govmomi client requires a URL object, u.
	// You cannot use a string representation of the vCenter URL.
	// soap.ParseURL provides the correct object format.

	u, err := soap.ParseURL(vc)

	if u == nil {
		setupLog.Error(err, "Unable to parse URL. Are required environment variables set?", "controller", "NodeInfo")
		os.Exit(1)
	}

	if err != nil {
		setupLog.Error(err, "URL parsing not successful", "controller", "NodeInfo")
		os.Exit(1)
	}

	u.User = url.UserPassword(user, pwd)

	// Session cache example taken from https://github.com/vmware/govmomi/blob/master/examples/examples.go
	// Share govc's session cache
	s := &cache.Session{
		URL:      u,
		Insecure: true,
	}
	// Create new vim25 client
	vim25client := new(vim25.Client)

	// Login using client vim25client and cache s
	err = s.Login(ctx, vim25client, nil)

	if err != nil {
		setupLog.Error(err, "Vkubeviewer: vim25 login not successful", "manager", "Vkubeviewer")
		os.Exit(1)
	}

	// Create new rest client
	restclient := rest.NewClient(vim25client)
	err = restclient.Login(ctx, u.User)

	if err != nil {
		setupLog.Error(err, "Vkubeviewer: rest login not successful", "controller", "Vkubeviewer")
		os.Exit(1)
	}

	return vim25client, restclient, nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "982940d6.vkubeviewer.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Retrieve vCenter URL, username and password from environment variables
	// These are provided via the manager manifest when controller is deployed

	vc := os.Getenv("GOVMOMI_URL")
	user := os.Getenv("GOVMOMI_USERNAME")
	pwd := os.Getenv("GOVMOMI_PASSWORD")

	// Create context, and get vSphere session information
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vim25client, restclient, err := vlogin(ctx, vc, user, pwd)
	if err != nil {
		setupLog.Error(err, "unable to get login session to vSphere")
		os.Exit(1)
	} else {
		setupLog.Info("succeed to get login session to vSphere")

	}

	finder := find.NewFinder(vim25client, true)

	// find and set the default datacenter

	dc, err := finder.DefaultDatacenter(ctx)

	if err != nil {
		setupLog.Error(err, "Manager: Could not get default datacenter")
	} else {
		finder.SetDatacenter(dc)

	}

	err = createDatastoreInfo(ctx, mgr, vim25client)
	if err != nil {
		setupLog.Error(err, "Manager: Could not create DatastorInfo")
	}

	err = createHostInfo(ctx, mgr, vim25client)
	if err != nil {
		setupLog.Error(err, "Manager: Could not create HostInfo")
	}

	err = createNodeInfo(ctx, mgr, vim25client)
	if err != nil {
		setupLog.Error(err, "Manager: Could not create NodeInfo")
	}

	err = createTagInfo(ctx, mgr, restclient)
	if err != nil {
		setupLog.Error(err, "Manager: Could not create TagInfo")
	}

	err = createFCDInfo(ctx, mgr, vim25client)
	if err != nil {
		setupLog.Error(err, "Manager: Could not create FCDInfo")
	}

	//Modified Reconcile call
	if err = (&controllers.FCDInfoReconciler{
		Client: mgr.GetClient(),
		VC:     vim25client,
		Finder: finder,
		Log:    ctrl.Log.WithName("controllers").WithName("FCDInfo"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FCDInfo")
		os.Exit(1)
	}

	if err = (&controllers.NodeInfoReconciler{
		Client:   mgr.GetClient(),
		VC_vim25: vim25client,
		VC_rest:  restclient,
		Log:      ctrl.Log.WithName("controllers").WithName("NodeInfo"),
		Scheme:   mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeInfo")
		os.Exit(1)
	}

	if err = (&controllers.HostInfoReconciler{
		Client: mgr.GetClient(),
		VC:     vim25client,
		Log:    ctrl.Log.WithName("controllers").WithName("HostInfo"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HostInfo")
		os.Exit(1)
	}

	if err = (&controllers.DatastoreInfoReconciler{
		Client: mgr.GetClient(),
		VC:     vim25client,
		Log:    ctrl.Log.WithName("controllers").WithName("DatastoreInfo"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatastoreInfo")
		os.Exit(1)
	}
	if err = (&controllers.TagInfoReconciler{
		Client:   mgr.GetClient(),
		VC_vim25: vim25client,
		VC_rest:  restclient,
		Log:      ctrl.Log.WithName("controllers").WithName("TagInfo"),
		Scheme:   mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TagInfo")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
