# Quick Start guide

## A guide to quickly use the vkubeviewer operator.

**Step 1 :** Install the necessary software dependencies:

- A **git** client/command line
- [Go (v1.15+)](https://golang.org/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop)
- [Kubebuilder](https://go.kubebuilder.io/quick-start.html)
- [Kustomize](https://kubernetes-sigs.github.io/kustomize/installation/)
- Access to a Container Image Repository (docker.io, quay.io, harbor)
- Make binary  → `sudo apt install make`

**Step: 1.1:** Run the go-pack.sh script to get required go packages

```
./ go-pack.sh
```

**Step 2:** Get the **vKubeViewer** operator to your desktop

```
git clone https://github.com/vKubeViewer/vkubeviewer.git
```

You can check the currently installed CRDs on your K8s cluster by :

```
kubectl get crd
```

**Step3:**  Install the CRDs from this operator.

```
cd vkubeviewer

make install
```

**Step 4:**  Check your newly installed CRDs.

```
kubectl get crd
```

**Step 5**: **Edit** the CR sample yaml in *config/samples* folder, choose the one you want to view. For instance to view VM information, edit the spec field and put your VM name in **nodename**  field in ***config/samples/topology_v1_vminfo.yaml*** as shown below:

```
cd config/samples
cat topology_v1_vminfo.yaml 
```

```
apiVersion: topology.vkubeviewer.com/v1
kind: VMInfo
metadata:
name: k8s-worker-1
spec:
# Add fields here or edit
nodename: k8s-worker-01
```

**Step 6:** **Apply** the above YAML to create your custom resource

```
kubectl apply -f topology_v1_vminfo.yaml
```

**Step 7:** **Query** the CR we just created, check if the **nodename** field is also printed.

```
kubectl get vminfo

NAME           NODENAME
k8s-worker-1   k8s-worker-01
```

**Step 8:**To build the manager code locally, you can run the following make command: 

**Note:** Skip to step 11 if you want to build the manager on a pod using a publicly accessible image.

```
cd ../..
make manager 
```

This should have build the manager binary in bin/manager. Before running the manager in standalone code, we need to set three environmental variables to allow us to connect to the vCenter Server. They are:

```
export GOVMOMI_URL=Your_Vcenter_URL
export GOVMOMI_USERNAME=Your_Username@vsphere.local
export GOVMOMI_PASSWORD=Your_VC_Password
```

**Step 9:** The manager can now be started in standalone mode, run:

```
bin/manager
```

The output should look like:

```
2021-06-30T16:35:05.649+0100	INFO	controller-runtime.metrics	metrics server is starting to listen	{"addr": ":8080"}
2021-06-30T16:35:05.650+0100	INFO	setup	starting manager
2021-06-30T16:35:05.650+0100	INFO	controller-runtime.manager.controller.fcdinfo	Starting EventSource	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "FCDInfo", "source": "kind source: /, Kind="}
2021-06-30T16:35:05.650+0100	INFO	controller-runtime.manager.controller.vminfo	Starting EventSource	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "VMInfo", "source": "kind source: /, Kind="}
2021-06-30T16:35:05.650+0100	INFO	controller-runtime.manager.controller.hostinfo	Starting EventSource	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "HostInfo", "source": "kind source: /, Kind="}
2021-06-30T16:35:05.650+0100	INFO	controller-runtime.manager	starting metrics server	{"path": "/metrics"}
2021-06-30T16:35:05.751+0100	INFO	controller-runtime.manager.controller.hostinfo	Starting Controller	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "HostInfo"}
2021-06-30T16:35:05.751+0100	INFO	controller-runtime.manager.controller.vminfo	Starting Controller	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "VMInfo"}
2021-06-30T16:35:05.751+0100	INFO	controller-runtime.manager.controller.vminfo	Starting workers	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "VMInfo", "worker count": 1}
2021-06-30T16:35:05.751+0100	INFO	controller-runtime.manager.controller.hostinfo	Starting workers	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "HostInfo", "worker count": 1}
2021-06-30T16:35:05.751+0100	INFO	controller-runtime.manager.controller.fcdinfo	Starting Controller	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "FCDInfo"}
2021-06-30T16:35:05.752+0100	INFO	controller-runtime.manager.controller.fcdinfo	Starting workers	{"reconciler group": "[topology.vkubeviewer.com](http://topology.vkubeviewer.com/)", "reconciler kind": "FCDInfo", "worker count": 1}
2021-06-30T16:35:05.752+0100	INFO	controllers.VMInfo	received reconcile request for "k8s-worker-1" (namespace: "default")	{"VMInfo": "default/k8s-worker-1"}
2021-06-30T16:35:05.811+0100	INFO	controllers.VMInfo	received reconcile request for "k8s-worker-1" (namespace: "default")	{"VMInfo": "default/k8s-worker-1"}
```

You can apply more CRDs from the samples folder for other resources. 

**Step 10** : We can run the below command to see the required fields in the status field of the CRD.

```
kubectl get vminfo -o yaml
```

Output:

```
apiVersion: [topology.vkubeviewer.com/v1](http://topology.vkubeviewer.com/v1)
kind: VMInfo
metadata:
annotations:
[kubectl.kubernetes.io/last-applied-configuration:](http://kubectl.kubernetes.io/last-applied-configuration:) |
{"apiVersion":"[topology.vkubeviewer.com/v1","kind":"VMInfo","metadata":{"annotations":{},"name":"k8s-worker-1","namespace":"default"},"spec":{"nodename":"k8s-worker-01](http://topology.vkubeviewer.com/v1%22,%22kind%22:%22VMInfo%22,%22metadata%22:%7B%22annotations%22:%7B%7D,%22name%22:%22k8s-worker-1%22,%22namespace%22:%22default%22%7D,%22spec%22:%7B%22nodename%22:%22k8s-worker-01)"}}
creationTimestamp: "2021-06-30T15:34:22Z"
generation: 1
name: k8s-worker-1
namespace: default
resourceVersion: "16371295"
uid: ac38c909-81a5-4e16-a6b1-d6efa4602136
spec:
nodename: k8s-worker-01
**status:
guestId: ubuntu64Guest
hwVersion: vmx-18
ipAddress: 10.27.51.54
pathToVM: '[vsan-OCTO-Cluster-A] 14fe8760-f3fc-92ac-297b-246e962f4854/K8s-Worker-01.vmx'
powerState: poweredOn
resvdCPU: 0
resvdMem: 0
totalCPU: 4
totalMem: 4096**
kind: List
metadata:
resourceVersion: ""
selfLink: ""
```

## Running the controller-manager on a pod in your K8s cluster

**Step 11:** Login into [Docker.io](http://docker.io) as you will need to get the controller image stored in vkubeviewer repository.

```
docker login —username dockerID —password 'My_password'
```

Set the environment variable IMG to point at the required image.

```
export IMG=docker.io/vkubeviewer/controller-manager:latest
```

**Step 12:** Create the **namespace** and **secret** used by the controller pod.

```
kubectl create ns vkubeviewer-system
```

**Note:** Do not forget to change the credentials wit your GOVMOMI credentials

```
kubectl create secret generic vc-creds-1 \
--from-literal='GOVMOMI_USERNAME= **Username**' \
--from-literal='GOVMOMI_PASSWORD=**Password**' \
--from-literal='GOVMOMI_URL=192.168.0.100' \
-n vkubeviewer-system
[output]secret/vc-creds-1 created
```

**Step13:** Create the deployment with 1 replica set which ensures that the controller pod keeps running. run:

```
make deploy
```

**Step 14:** Check the pod is running fine with both the containers in ready and running state.

```
kubectl get pods -n vkubeviewer-system

NAME                                          READY   STATUS    RESTARTS   AGE
vkubeviewer-controller-manager-566c6fffdb-fxjr2   2/2     Running   0          2m39s
```

**Step 15:** Re-apply the sample YAMLs for the custom resources to be monitored by the above pod. 

```
kustomize build config/samples | kubectl create -f -
```

**Step 16:** Finally, we can run the below command to see the required fields in the status field of the CRDs.

```
kubectl get hostinfo -o yaml
kubectl get vminfo -o yaml
kubectl get fcdinfo -o yaml
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

### Step 17: Clean Up.

Remove the CRDs

```
kustomize build config/samples | kubectl delete -f -
make uninstall
```

Remove the deployment and related namespace and secret

```
make undeploy
```

## Thank you.

### If you face any issue please reach out to us.
