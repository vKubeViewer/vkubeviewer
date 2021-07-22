# vKubeViewer

This document contains a Kubernetes Operator that uses VMware's **Govmomi** to return some simple ESXi host information through the status fields of a **Custom Resource (CR)**, which are called HostInfo, NodeInfo, FCDInfo etc... This will require us to extend Kubernetes with a new **Custom Resource Definition (CRD)**. The code shown is one way in which a Kubernetes controller/operator can access the underlying vSphere infrastructure for the purposes of querying resources.

You can think of a CRD as representing the desired state of a Kubernetes object or Custom Resource, and the function of the operator is to run the logic or code to make that desired state happen - in other words the operator has the logic to do whatever is necessary to achieve the object's desired state.

## **What are we building here?**

We will create a few CRDs called HostInfo, NodeInfo, FCDInfo etc... For instance, HostInfo will contain the name of an ESXi host in its specification. When a Custom Resource (CR) is created and subsequently queried, we will call an operator (logic in a controller) whereby the Total CPU and Free CPU from the ESXi host will be returned via the status fields of the object through Govmomi API calls.

In this document, we will talk about the initial stages of the operator which runs with 3 CRDs namely  HostInfo, NodeInfo and FCDInfo. However, we intend to populate it with more in the future.

# Operator Development Steps

## **Step 1 - Software Requirements**

You will need the following components pre-installed on your desktop or workstation before we can build the CRD and operator.

- A **git** client/command line
- [Go (v1.15+)](https://golang.org/dl/) - earlier versions may work but I used v1.15.
- [Docker Desktop](https://www.docker.com/products/docker-desktop)
- [Kubebuilder](https://go.kubebuilder.io/quick-start.html)
- [Kustomize](https://kubernetes-sigs.github.io/kustomize/installation/)
- Access to a Container Image Repository (docker.io, quay.io, harbor)
- A **make** binary - used by Kubebuilder
- Run the [go-pack.sh](http://go-pack.sh) script to install required go packages.

## **Step 2 - KubeBuilder Scaffolding**

The CRD is built using [kubebuilder](https://go.kubebuilder.io/). KubeBuilder builds a directory structure containing all of the templates (or scaffolding) necessary for the creation of CRDs. Once this scaffolding is in place, this doc will show you how to add your own specification fields and status fields, as well as how to add your own operator logic. In this example, our logic will log in to vSphere, query and return required fields via a Kubernetes CR / object / Kind called HostInfo, NodeInfo, FCDInfo. The values of which will be used to populate status fields in our CRs.

The following steps will create the scaffolding to get started.

```
mkdir vkubeviewer
```

Next, define the Go module name of your CRD. In this case, we have called it vkubeviewer. This creates a go.mod file with the name of the module and the Go version (v1.16 here)

```
cd vkubeviewer/
go mod init vkubeviewer
go: creating new go.mod: module vkubeviewer

ls
go.mod
```

Now we can proceed with building out the rest of the directory structure. The following kubebuilder commands (init and create api) creates all the scaffolding necessary to build our CRD and operator.

```
kubebuilder init —domain vkubeviewer.com
```

Next, we must define a resource. To do that, we again use kubebuilder to create the resource, specifying the API group, its version and supported kind. Our group is called topology, and kind is called HostInfo, VMinfo, FCDInfo and our initial version is v1.

```
kubebuilder create api --group topology --version v1 --kind HostInfo --resource=true --controller=true
kubebuilder create api --group topology --version v1 --kind VMInfo --resource=true --controller=true
kubebuilder create api --group topology --version v1 --kind FCDInfo --resource=true --controller=true
```

## Step 3 - Create the CRD

Customer Resource Definitions [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) are a way to extend Kubernetes through Custom Resources. We are going to extend a Kubernetes cluster with three custom resources called **HostInfo, NodeInfo , FCDInfo** which will retrieve required information from vSphere, with respect to the name specified in a Custom Resource. Thus, we will need to create a field called **hostname, nodename, PVId** in the respective CRD - this defines the specification of the custom resource. We also add status fields, as these will be used to return information about the ESXI host, Virtual machine or the First Class Disk (FCD) backing the PV.

This is done by modifying the **hostinfo_types.go, NodeInfo_types.go, fcdinfo_types.go** files in **api/v1/** folder ****. Here is the edited **NodeInfo_types.go file** as an example.

```
// NodeInfoSpec defines the desired state of NodeInfo
type NodeInfoSpec struct {
	Nodename string `json:"nodename"`
}

// NodeInfoStatus defines the observed state of NodeInfo
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
}

// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName={"nd"}
// +kubebuilder:printcolumn:name="Nodename",type=string,JSONPath=`.spec.nodename`
// +kubebuilder:printcolumn:name="VMTotalCPU",type=string,JSONPath=`.status.vm_total_cpu`
// +kubebuilder:printcolumn:name="VMTotalMem",type=string,JSONPath=`.status.vm_total_mem`
// +kubebuilder:printcolumn:name="VMPowerState",type=string,JSONPath=`.status.vm_power_state`
// +kubebuilder:printcolumn:name="VMIpAddress",type=string,JSONPath=`.status.vm_ip_address`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
```

This file is modified to include a single spec.nodename field and to return fourteen status fields. There are also a number of kubebuilder fields added, which are used to do validation and other kubebuilder related functions. The shortname "ch" will be used later on in our controller logic. This can also be used with kubectl, e.g kubectl get ch rather than kubectl get NodeInfo.

We are now ready to create the CRD. There is one final step, however, and this involves updating the **Makefile** which kubebuilder has created for us. In the default Makefile created by kubebuilder, the following **CRD_OPTIONS** line appears:

```
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
```

This CRD_OPTIONS entry should be changed to the following:

```
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:preserveUnknownFields=false,crdVersions=v1,trivialVersions=true"
```

Now we can build our CRDs with the spec and status fields that we have place in the **api/v1/**types.go** files.

```
make manifests && make generate
```

## Step 4 - Install the CRD

The CRDs is not currently installed in the Kubernetes Cluster.

```
$ kubectl get crd
NAME                                        CREATED AT
[cnsvspherevolumemigrations.cns.vmware.com](http://cnsvspherevolumemigrations.cns.vmware.com/)   2021-04-28T10:37:18Z
```

To install the CRD, run the following make command:

```
make install
```

Now check to see if the CRD is installed running the same command as before.

```
kubectl get crd
NAME                                        CREATED AT
[cnsvspherevolumemigrations.cns.vmware.com](http://cnsvspherevolumemigrations.cns.vmware.com/)   2021-04-28T10:37:18Z
[fcdinfoes.topology.vkubeviewer.com](http://fcdinfoes.topology.vkubeviewer.com/)          2021-07-03T13:42:35Z
[hostinfoes.topology.vkubeviewer.com](http://hostinfoes.topology.vkubeviewer.com/)         2021-07-03T13:42:35Z
[nodeinfoes.topology.vkubeviewer.com](http://vminfoes.topology.vkubeviewer.com/)           2021-07-03T13:42:35Z
```

Our new CRDs are now visible. Another useful way to check if the CRDs have successfully deployed is to use the following command against our API group. Remember back in step 2 we specified the domain as `vkubeviewer[.com](http://corinternal.com/)` and the group as **`topology`**

```
kubectl api-resources --api-group=[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)
NAME         SHORTNAMES   APIVERSION                    NAMESPACED   KIND
fcdinfoes    fcd          [topology.vkubeviewer.com/v1](http://topology.vkubeviewer.com/v1)   true         FCDInfo
hostinfoes   hi           [topology.vkubeviewer.com/v1](http://topology.vkubeviewer.com/v1)   true         HostInfo
nodeinfoes     ch           [topology.vkubeviewer.com/v1](http://topology.vkubeviewer.com/v1)   true         NodeInfo
```

## Step 5 - Test the CRD

At this point, we can do a quick test to see if our CRD is in fact working. To do that, we can create a manifest file with a Custom Resource that uses our CRD, and see if we can instantiate such an object (or custom resource) on our Kubernetes cluster. Fortunately, kubebuilder provides us with a sample manifest that we can use for this. It can be found in **config/samples**.

```
cd config/samples
ls
topology_v1_fcdinfo.yaml  
topology_v1_hostinfo.yaml  
topology_v1_nodenfo.yaml
```

```
$ cat topology_v1_fcdinfo.yaml
apiVersion: topology.corinternal.com/v1
kind: FCDInfo
metadata:
  name: fcdinfo-sample
spec:
  # Add fields here
  foo: bar
```

We need to slightly modify this sample manifest so that the specification field matches what we added to our CRD. Note the spec: above where it states 'Add fields here'. We have removed the foo field and added a **spec.pvId** field, as per the **api/v1/fcdinfo_types.go** modification earlier. Thus, after a simple modification, the CR manifest looks like this, where pvID is the name of the persistent volume that we wish to query.

```

shaunak@shaunak-desktop:~/vkubeviewer/config/samples$ cat topology_v1_fcdinfo.yaml
apiVersion: [topology.vkubeviewer.com/v1](http://topology.vkubeviewer.com/v1)
kind: FCDInfo
metadata:
name: fcdinfo-sample
spec:
# Add fields here
pvId: pvc-b8458bef-178e-40dd-9bc0-2a05f1ddfd65
```

Similarly, edit the other two yaml providing the ESXi hostname and the virtual machine (k8s nodename) that you wish to query 

```
cat topology_v1_hostinfo.yaml
apiVersion: [topology.vkubeviewer.com/v1](http://topology.vkubeviewer.com/v1)
kind: HostInfo
metadata:
name: hostinfo-host-h
spec: 
#Add fields here
hostname: [esxi-dell-h.rainpole.com](http://esxi-dell-h.rainpole.com/)
```

```
cat topology_v1_nodeinfo.yaml
apiVersion: [topology.vkubeviewer.com/v1](http://topology.vkubeviewer.com/v1)
kind: NodeInfo
metadata:
name: k8s-worker-1
spec:
#Add fields here
nodename: k8s-worker-01
```

To see if it works, we need to create this FCDInfo Custom Resource.

```
kubectl apply -f .
[fcdinfo.topology.vkubeviewer.com/fcdinfo-sample](http://fcdinfo.topology.vkubeviewer.com/fcdinfo-sample) created
[hostinfo.topology.vkubeviewer.com/hostinfo-host-h](http://hostinfo.topology.vkubeviewer.com/hostinfo-host-h) created
[nodeinfo.topology.vkubeviewer.com/k8s-worker-1](http://vminfo.topology.vkubeviewer.com/k8s-worker-1) created
```

Check if the CRs are now present:

```
kubectl get nodeinfo
NAME           NODENAME
k8s-worker-1   k8s-worker-01

kubectl get hostinfo
NAME              HOSTNAME
hostinfo-host-h   [esxi-dell-h.rainpole.com](http://esxi-dell-h.rainpole.com/)

kubectl get fcdinfo
NAME             PVID
fcdinfo-sample   pvc-b8458bef-178e-40dd-9bc0-2a05f1ddfd65
```

## Step 6 - Create the Controller/Manager

This appears to be working as expected. However, there are no **Status** fields displayed with our PV information in the **yaml** output above. To see this information, we need to implement our operator/controller logic to do this. The controller implements the desired business logic. In this controller, we first read the vCenter server credentials from a Kubernetes secret (which we will create shortly). We will then open a session to my vCenter server, and for instance, get a list of First Class Disks that it manages. We then look for the FCD that is specified in the spec.pvId field in the CR, and retrieve the information this it's backing FCD. Finally, we will update the appropriate Status field with this information, and we should be able to query it using the **kubectl get fcdinfo -o yaml** command seen previously.

### Step 6.1 - Open a Session to vSphere

**Note:** Let's first look at the login function which resides in **main.go**. This **vlogin** function creates the vSphere session in main.go. One thing to note is that we are enabling insecure logins (true) by default. This is something that you may wish to change in your code. One other item to note is that
We are testing three different client logins here, **govmomi.Client, vim25.Client and rest.client**. The govmomi.Client uses **Finder** for getting vSphere information and treats the vSphere inventory as a virtual filesystem. The vim25.Client uses **ContainerView**, and tends to generate more response data. As mentioned, this is a tutorial, so this operator shows three login types simply for informational purposes. Most likely, you could achieve the same results using a single login client.

```
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
// Session cache example taken from <https://github.com/vmware/govmomi/blob/master/examples/examples.go>
//
// Share govc's session cache
//
s := &cache.Session{
	URL:      u,
	Insecure: true,
}

//
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
```

Next within the main function, there is a call to the vlogin function with the parameters received from the environment variables shown below.

```
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
```

There is also an updated Reconciler call(for all three)with new fields (VC1 & VC2) which have the vSphere session details. 

```
//Modified Reconcile call
//----
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
	VC:     c1,
	Log:    ctrl.Log.WithName("controllers").WithName("HostInfo"),
	Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
	setupLog.Error(err, "unable to create controller", "controller", "HostInfo")
	os.Exit(1)
}
```

This login information can now be used from within the Reconciler controller function, as we will see shortly.

### Step 6.2 - Controller Reconcile Logic

Now we turn our attention to the business logic of the controller. Once the business logic is added to the controller, it will need to be able to run it in a Kubernetes cluster. To achieve this, a container image to run the controller logic must be built. This will be provisioned in the Kubernetes cluster using a Deployment manifest. The deployment contains a single Pod that runs the container (it is called **manager**). The deployment ensures that the controller manager Pod is restarted in the event of a failure.

This is what kubebuilder provides as controller scaffolding - it is found in **controllers/*_controller.go**. We are most interested in the ***Reconciler** function, first, let's look at the modified FCDInfoReconciler structure for instance, which now has 2 new members representing the different clients, VC1 and VC2.

```
/ FCDInfoReconciler reconciles a FCDInfo object
type FCDInfoReconciler struct {
client.Client
VC1    *vim25.Client
VC2    *govmomi.Client
Finder *find.Finder
Log    logr.Logger
Scheme *runtime.Scheme
}
```

Note that we are switching between the different clients to get different information. In some parts, We are using the vim25 client to get information, and in others, we are using the govmomi client to get information. Again, this is just a learning exercise, to show various ways to retrieve vSphere information from an operator. The flow in the FCD controller is that we first get a list of datastores, then we retrieve the FCDs from each of the datastores. If any of the FCDs is a match for the PV we have specified in our manifest, then we populate the status fields for that PV with the requested information. We have added some additional logging messages to this controller logic, and we can check the manager logs to see these messages later on.

```
func (r *FCDInfoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
ctx = context.Background()
log := r.Log.WithValues("FCDInfo", req.NamespacedName)

fcd := &topologyv1.FCDInfo{}
if err := r.Client.Get(ctx, req.NamespacedName, fcd); err != nil {
	if !k8serr.IsNotFound(err) {
		log.Error(err, "unable to fetch FCDInfo")
	}
	return ctrl.Result{}, client.IgnoreNotFound(err)
}

msg := fmt.Sprintf("received reconcile request for %q (namespace: %q)", fcd.GetName(), fcd.GetNamespace())
log.Info(msg)

//
// Find the datastores available on this vSphere Infrastructure
//

dss, err := r.Finder.DatastoreList(ctx, "*")
if err != nil {
	log.Error(err, "FCDInfo: Could not get datastore list")
	return ctrl.Result{}, err
} else {
	msg := fmt.Sprintf("FCDInfo: Number of datastores found - %v", len(dss))
	log.Info(msg)

	pc := property.DefaultCollector(r.VC2.Client)
	//
	// "finder" only lists - to get really detailed info,
	// Convert datastores into list of references
	//
	var refs []types.ManagedObjectReference
	for _, ds := range dss {
		refs = append(refs, ds.Reference())
	}

	//
	// Retrieve name property for all datastore
	//

	var dst []mo.Datastore
	err = pc.Retrieve(ctx, refs, []string{"name"}, &dst)
	if err != nil {
		log.Error(err, "FCDInfo: Could not get datastore info")
		return ctrl.Result{}, err
	}

	m := vslm.NewObjectManager(r.VC1)

	//
	// -- Display the FCDs on each datastore (held in array dst)
	//

	var objids []types.ID
	var idinfo *types.VStorageObject

	for _, newds := range dst {
		objids, err = m.List(ctx, newds)
		//
		// -- With the list of FCD Ids, we can get further information about the FCD retrievec in VStorageObject
		//
		for _, id := range objids {
			idinfo, err = m.Retrieve(ctx, newds, id.Id)
			//
			// -- Note the TKGS Guest Clusters have a different PV ID
			// -- to the one that is created for them in the Supervisor
			// -- This only works for the Supervisor PV ID
			//
			if idinfo.Config.BaseConfigInfo.Name == fcd.Spec.PVId {
				msg := fmt.Sprintf("FCDInfo: %v matches %v", idinfo.Config.BaseConfigInfo.Name, fcd.Spec.PVId)
				log.Info(msg)

				fcd.Status.SizeMB = int64(idinfo.Config.CapacityInMB)

				backing := idinfo.Config.BaseConfigInfo.Backing.(*types.BaseConfigInfoDiskFileBackingInfo)
				fcd.Status.FilePath = string(backing.FilePath)
				fcd.Status.ProvisioningType = string(backing.ProvisioningType)
			}
			// else {
			// 	msg := fmt.Sprintf("FCDInfo: %v does not match %v", idinfo.Config.BaseConfigInfo.Name, fcd.Spec.PVId)
			// 	log.Info(msg)
			// }
		}
	}

	if err := r.Status().Update(ctx, fcd); err != nil {
		log.Error(err, "unable to update FCDInfo status")
		return ctrl.Result{}, err
	}
}

return ctrl.Result{}, nil
}
```

## Step 7 - Build the Controller

At this point, everything is in place to enable us to deploy the controller to the Kubernete cluster. If you remember back to the prerequisites in step 1, we said that you need access to a container image registry, such as **docker.io** or **quay.io**, or VMware's own [Harbor](https://github.com/goharbor/harbor/blob/master/README.md) registry. This is where we need this access to a registry, as we need to push the controller's container image somewhere that can be accessed from your Kubernetes cluster. In this example, We are using docker.io as the 
repository.

The **Dockerfile** with the appropriate directives is already in place to build the container image and include the controller/manager logic. This was once again taken care of by kubebuilder. You 
must ensure that you login to your image repository, i.e. docker login, before proceeding with the **make** commands, e.g.

```
docker login
Authenticating with existing credentials...
WARNING! Your password will be stored unencrypted in /home/shaunak/.docker/config.json.
Configure a credential helper to remove this warning. See
[https://docs.docker.com/engine/reference/commandline/login/#credentials-store](https://docs.docker.com/engine/reference/commandline/login/#credentials-store)
Login Succeeded
```

set an environment variable called IMG to point to your container image repository along with the name and version of the container image, e.g.:

```
export [IMG=docker.io/vkubeviewer/controller-manager:v3](http://img=docker.io/vkubeviewer/controller-manager:v3)

```

Next, to create the container image of the controller/manager, and push it to the image container repository in a single step, run the following make command.

```
make docker-build docker-push
```

The container image of the controller is now built and pushed to the container image registry. But we have not yet deployed it. We have to do one or two further modifications before we take that step.

## Step 8 - Modify the Manager Manifest to Include Environment Variables

Kubebuilder provides a manager manifest scaffold file for deploying the controller. However, since we need to provide vCenter details to our controller, we need to add these to the controller/manager manifest file. This is found in **config/manager/manager.YAML**. This manifest contains the deployment for the controller. In the spec, we need to add an additional **spec.env** section which has the environment variables defined, as well as the name of our **secret** (which we will create shortly). Below is a snippet of that code. 

```
env:
- name: GOVMOMI_USERNAME
valueFrom:
secretKeyRef:
name: vc-creds-1
key: GOVMOMI_USERNAME
- name: GOVMOMI_PASSWORD
valueFrom:
secretKeyRef:
name: vc-creds-1
key: GOVMOMI_PASSWORD
- name: GOVMOMI_URL
valueFrom:
secretKeyRef:
name: vc-creds-1
key: GOVMOMI_URL
volumes:
- name: vc-creds-1
secret:
secretName: vc-creds-1
```

Note that the secret, called vc-creds above, contains the vCenter credentials. This secret needs to be deployed in the same namespace that the controller is going to run in, which is vkubeviewer-system. Thus, the namespace and secret are created using the following commands, with the environment modified to your own vSphere infrastructure:

```
$ kubectl create ns vkubeviewer-system
namespace/fcdinfo-system created
```

```
$ kubectl create secret generic vc-creds \
--from-literal='GOVMOMI_USERNAME=admin@vsphere.local' \
--from-literal='GOVMOMI_PASSWORD=Password' \
--from-literal='GOVMOMI_URL=192.168.0.100' \
-n fcdinfo-system
```

We are now ready to deploy the controller to the Kubernetes cluster.

## Step 9 - Deploy the Controller

To deploy the controller, we run another **make** command. This will take care of all of the RBAC, cluster roles and role bindings necessary to run the controller, as well as pinging up the correct image, etc.

```
make deploy [IMG=docker.io/vkubeviewer/controller-manager:v3](http://img=docker.io/vkubeviewer/controller-manager:v3)

```

## Step 10 - Check Controller Functionality

Now that our controller has been deployed let's see if it is working. There are a few different commands that we can run to verify the operator is working.

### Step 10.1 - Check the Deployment and Replicaset

The deployment should be READY. Remember to specify the namespace correctly when checking it.

```
$ kubectl get rs -n vkubeviewer-system
NAME                                    DESIRED   CURRENT   READY   AGE
fcdinfo-controller-manager-566c6fffdb   1         1         1       102s

$ kubectl get deploy -n vkubeviewer-system
NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
fcdinfo-controller-manager    1/1     1            1           2m8s
```

### Step 10.2 - Check the Pods

The deployment manages a single controller Pod. There should be 2 containers READY in the controller Pod. One is the **controller/manager** and the other is the **kube-rbac-proxy**. The [kube-rbac-proxy](https://github.com/brancz/kube-rbac-proxy/blob/master/README.md) is a small HTTP proxy that can perform RBAC authorization against the Kubernetes API. It restricts requests to authorized Pods only.

```
$ kubectl get pods -n vkubeviewer-system
NAME                                          READY   STATUS    RESTARTS   AGE
vkubeviewer-controller-manager-566c6fffdb-fxjr2   2/2     Running   0          2m39s
```

If you experience issues with one of the pods not coming online, 
use the following command to display the Pod status and examine the 
events.

```
kubectl describe pod vkubeviewer-controller-manager-566c6fffdb-fxjr2 -n vkubeviewer-system
```

### Step 10.3 - Check the controller / manager logs

If we query the **logs** on the manager container, we should be able to observe successful startup messages as well as successful reconcile requests from the FCDInfo CR that we already deployed back in step 5. These reconcile requests should update the **Status** fields with FCD information as per our controller logic. The command to query the manager container logs in the controller Pod is as follows:

```
kubectl logs vkubeviewer-controller-manager-566c6fffdb-fxjr2 -n vkubeviewer-system manager
```

The output should be somewhat similar to this. Note that there is also a successful **Reconcile** operation reported, which is good. We can also see some log messages which were added to the controller logic.

```
$ kubectl logs vkubeviewer-controller-manager-566c6fffdb-fxjr2 -n vkubeviewer-system manager
2021-01-26T09:33:12.450Z        INFO    controller-runtime.metrics      metrics server is starting to listen    {"addr": "127.0.0.1:8080"}
2021-01-26T09:33:12.450Z        INFO    setup   starting manager
I0126 09:33:12.451024       1 leaderelection.go:242] attempting to acquire leader lease  fcdinfo-system/b610b79e.corinternal.com...
2021-01-26T09:33:12.451Z        INFO    controller-runtime.manager      starting metrics server {"path": "/metrics"}
I0126 09:33:29.856533       1 leaderelection.go:252] successfully acquired lease fcdinfo-system/b610b79e.corinternal.com
2021-01-26T09:33:29.857Z        INFO    controller-runtime.controller   Starting EventSource    {"controller": "fcdinfo", "source": "kind source: /, Kind="}
2021-01-26T09:33:29.857Z        DEBUG   controller-runtime.manager.events       Normal  {"object": {"kind":"ConfigMap","namespace":"fcdinfo-system","name":"[b610b79e.corinternal.com](http://b610b79e.corinternal.com/)","uid":"9bf00d05-f28b-40ff-97fe-49b0d1a72070","apiVersion":"v1","resourceVersion":"32794039"}, "reason": "LeaderElection", "message": "fcdinfo-controller-manager-566c6fffdb-fxjr2_05075012-2765-4de5-bc7d-d71aa7be687e became leader"}
2021-01-26T09:33:29.957Z        INFO    controller-runtime.controller   Starting Controller     {"controller": "fcdinfo"}
2021-01-26T09:33:29.957Z        INFO    controller-runtime.controller   Starting workers        {"controller": "fcdinfo", "worker count": 1}
2021-01-26T09:33:29.958Z        INFO    controllers.FCDInfo     received reconcile request for "fcdinfo-sample" (namespace: "default")  {"FCDInfo": "default/fcdinfo-sample"}
2021-01-26T09:33:29.963Z        INFO    controllers.FCDInfo     FCDInfo: Number of datastores found - 3 {"FCDInfo": "default/fcdinfo-sample"}
2021-01-26T09:33:31.530Z        DEBUG   controller-runtime.controller   Successfully Reconciled {"controller": "fcdinfo", "request": "default/fcdinfo-sample"}
```

### Step 10.4 - Check if statistics are returned in the status

Last but not least, let's see if we can see the  information in the **status** fields of the object created earlier.

```
kubectl get hostinfo -o yaml
kubectl get fcdinfo -o yaml
kubectl get vminfo -o yaml
```

Output of fcdinfo :

```
apiVersion: v1
items:
- apiVersion: topology.vkubeviewer.com/v1
  kind: FCDInfo
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"topology.vkubeviewer.com/v1","kind":"FCDInfo","metadata":{"annotations":{},"name":"fcdinfo-sample","namespace":"default"},"spec":{"pvId":"pvc-b8458bef-178e-40dd-9bc0-2a05f1ddfd65"}}
    creationTimestamp: "2021-07-01T16:02:22Z"
    generation: 1
    name: fcdinfo-sample
    namespace: default
    resourceVersion: "16637703"
    uid: 03439716-ccc4-41ae-9559-fd19c96ef362
  spec:
    pvId: pvc-b8458bef-178e-40dd-9bc0-2a05f1ddfd65
  **status:
    filePath: '[vsan-OCTO-Cluster-B] b2d46d60-cd6e-6724-576b-246e962f4ab4/12a23b18541b4a28965cf4af1e963578.vmdk'
    provisioningType: thin
    sizeMB: 5120**
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

### Step 11: Clean Up.

Remove the CRDs

```
make uninstall
```

Remove the deployment and related namespace and secret

```
make undeploy
```