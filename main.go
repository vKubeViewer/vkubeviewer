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
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"

	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	topologyv1 "vkubeviewer/api/v1"
	"vkubeviewer/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = topologyv1.AddToScheme(scheme)
	utilruntime.Must(topologyv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

//
// - vSphere session login function
//

func vlogin(ctx context.Context, vc, user, pwd string) (*vim25.Client, *govmomi.Client, error) {

	//
	// This section allows for insecure govmomi logins
	//

	var insecure bool
	flag.BoolVar(&insecure, "insecure", true, "ignore any vCenter TLS cert validation error")

	//
	// Create a vSphere/vCenter client
	//
	// The govmomi client requires a URL object, u.
	// You cannot use a string representation of the vCenter URL.
	// soap.ParseURL provides the correct object format.
	//

	u, err := soap.ParseURL(vc)

	if u == nil {
		setupLog.Error(err, "Unable to parse URL. Are required environment variables set?", "controller", "VMInfo")
		os.Exit(1)
	}

	if err != nil {
		setupLog.Error(err, "URL parsing not successful", "controller", "VMInfo")
		os.Exit(1)
	}

	u.User = url.UserPassword(user, pwd)

	//
	// Session cache example taken from https://github.com/vmware/govmomi/blob/master/examples/examples.go
	//
	// Share govc's session cache
	//
	s := &cache.Session{
		URL:      u,
		Insecure: true,
	}

	//
	// Create new client
	//
	c1 := new(vim25.Client)

	//
	// Login using client c and cache s
	//
	err = s.Login(ctx, c1, nil)

	if err != nil {
		setupLog.Error(err, "FCDInfo: vim25 login not successful", "controller", "VMInfo")
		os.Exit(1)
	}

	c2, err := govmomi.NewClient(ctx, u, insecure)

	if err != nil {
		setupLog.Error(err, "FCDInfo: gomvomi login not successful", "controller", "VMInfo")
		os.Exit(1)
	}

	return c1, c2, nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	//
	// Retrieve vCenter URL, username and password from environment variables
	// These are provided via the manager manifest when controller is deployed
	//

	vc := os.Getenv("GOVMOMI_URL")
	user := os.Getenv("GOVMOMI_USERNAME")
	pwd := os.Getenv("GOVMOMI_PASSWORD")

	//
	// Create context, and get vSphere session information
	//

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c1, c2, err := vlogin(ctx, vc, user, pwd)
	if err != nil {
		setupLog.Error(err, "unable to get login session to vSphere")
		os.Exit(1)
	}

	//----

	finder := find.NewFinder(c2.Client, true)

	//
	// -- find and set the default datacenter
	//

	dc, err := finder.DefaultDatacenter(ctx)

	if err != nil {
		setupLog.Error(err, "FCDInfo: Could not get default datacenter")
	} else {
		finder.SetDatacenter(dc)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "982940d6.vkubeviewer.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.FCDInfoReconciler{
		Client: mgr.GetClient(),
		VC1:    c1,
		VC2:    c2,
		Finder: finder,
		Log:    ctrl.Log.WithName("controllers").WithName("FCDInfo"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FCDInfo")
		os.Exit(1)
	}

	if err = (&controllers.VMInfoReconciler{
		Client: mgr.GetClient(),
		VC:     c1,
		Log:    ctrl.Log.WithName("controllers").WithName("VMInfo"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VMInfo")
		os.Exit(1)
	}
	if err = (&controllers.HostInfoReconciler{
		Client: mgr.GetClient(),
		VC:     c1,
		Log:    ctrl.Log.WithName("controllers").WithName("HostInfo"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HostInfo")
		os.Exit(1)
	}
	if err = (&controllers.NodeNetInfoReconciler{
		Client: mgr.GetClient(),
		VC:     c1,
		Log:    ctrl.Log.WithName("controllers").WithName("NodeNetInfo"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeNetInfo")
		os.Exit(1)
	}
	if err = (&controllers.DatastoreInfoReconciler{
		Client: mgr.GetClient(),
		VC:     c1,
		Log:    ctrl.Log.WithName("controllers").WithName("DatastoreInfo"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatastoreInfo")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
