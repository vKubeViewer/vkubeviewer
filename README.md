# vKubeViewer

## About the vKubeViewer project

The vKubeViewer project is a collaborative project between VMware and students of University College Cork (Ireland) with the aim of creating a Kubernetes operator that will facilitate the retrival of vSphere resource usage infomation from within Kubernetes.

Upon initial 1.0 release, the project will be open sourced and the Kubernetes community is free to use and update the project with additional features and functionality. 


### Current Version 
Alpha: 1.0

## Problem Statement

The key problem for this project is that Kubernetes has no visibility into the underlying vSphere platform, neither from an infrastructure configuration perspective, nor from a resource usage and availability perspective. Examples of resource information includes CPU / RAM usage, storage usage, and network usage. This lack of visibility results in users having to switch context out of the Kubernetes environment and entering the vSphere environment when wishing to retrieve platform resource information. In many cases, the Kubernetes Platform administrator may not even have access to the vSphere platform, and may need to open a ticket with the vSphere platform team to retrieve this information on their behalf. This represents a large time loss for users and organizations. The time loss and subsequent costs can quickly add up for organizations when there are many virtualized Kubernetes clusters running on a vSphere platform.

## Implementation

The Operator is made up of a set of 5 CRDs, each representing a particular area of vSphere that we are interested in querying. The Operator is written in GO, and relies heavily on VMware's govmomi APIs. Whilst we have included a number of sample outputs below, please be aware that these outputs can be modified to provide pretty much anything that can be retrieved by the govmomi API.

## Usage Example

Below are some sample outputs taken from the vKubeViewer Operator.

The `hostinfo` CRD reports on ESXi hosts / hypervisors.

```
% kubectl get hostinfo
NAME                             HOSTNAME                         TOTALCPU   FREECPU   TOTALMEMORY   FREEMEMORY   TOTALSTORAGE   FREESTORAGE   INMAINTENANCEMODE
esxi-dell-e.rainpole.com         esxi-dell-e.rainpole.com         43980      42715     127.91 GiB    74.53 GiB    54.82 TiB      48.23 TiB     false
esxi-dell-f.rainpole.com         esxi-dell-f.rainpole.com         43980      40352     127.91 GiB    70.76 GiB    54.82 TiB      48.23 TiB     false
esxi-dell-g.rainpole.com         esxi-dell-g.rainpole.com         43980      42310     127.91 GiB    74.13 GiB    54.82 TiB      48.23 TiB     false
esxi-dell-h.rainpole.com         esxi-dell-h.rainpole.com         43980      43899     127.91 GiB    111.06 GiB   51.19 TiB      45.98 TiB     true
esxi-dell-i.rainpole.com         esxi-dell-i.rainpole.com         43980      42765     127.91 GiB    101.70 GiB   51.19 TiB      45.98 TiB     false
esxi-dell-j.rainpole.com         esxi-dell-j.rainpole.com         43980      43291     127.91 GiB    98.51 GiB    52.64 TiB      47.01 TiB     false
esxi-dell-k.rainpole.com         esxi-dell-k.rainpole.com         43980      43450     127.91 GiB    108.41 GiB   52.64 TiB      47.01 TiB     false
esxi-dell-l.rainpole.com         esxi-dell-l.rainpole.com         44000      43828     127.91 GiB    111.46 GiB   52.64 TiB      47.01 TiB     false
```

The `nodeinfo` CRD provides information related to the Kubernetes nodes as virtual machines.

```
% kubectl get nodeinfo
NAME                  VMTOTALCPU   VMTOTALMEM   VMPOWERSTATE   VMIPADDRESS   VMHWVERSION   CLUSTER          HOST                       DATASTORE
k8s-controlplane-01   4            4096         poweredOn      10.27.51.17   vmx-10        OCTO-Cluster-A   esxi-dell-f.rainpole.com   ["vsan-OCTO-Cluster-A"]
k8s-worker-01         4            4096         poweredOn      10.27.51.54   vmx-19        OCTO-Cluster-A   esxi-dell-f.rainpole.com   ["vsan-OCTO-Cluster-A"]
k8s-worker-02         4            4096         poweredOn      10.27.51.25   vmx-18        OCTO-Cluster-B   esxi-dell-k.rainpole.com   ["vsan-OCTO-Cluster-B"]
k8s-worker-03         4            4096         poweredOn      10.27.51.28   vmx-18        OCTO-Cluster-C   esxi-dell-i.rainpole.com   ["vsan-OCTO-Cluster-C"]
k8s-worker-04         4            4096         poweredOn      10.27.51.26   vmx-19        OCTO-Cluster-A   esxi-dell-f.rainpole.com   ["vsan-OCTO-Cluster-A"]
k8s-worker-05         4            4096         poweredOn      10.27.51.32   vmx-10        OCTO-Cluster-B   esxi-dell-j.rainpole.com   ["vsan-OCTO-Cluster-B"]
k8s-worker-06         4            4096         poweredOn      10.27.51.18   vmx-19        OCTO-Cluster-C   esxi-dell-i.rainpole.com   ["vsan-OCTO-Cluster-C"]
```

The `datastoreinfo` CRD informs us which hosts in the vSphere environment mount particular datastores, along with type and usage information.
```
% kubectl get datastoreinfo
NAME                  DATASTORE             TYPE   CAPACITY     FREESPACE    HOSTSMOUNTED
isilon-01             isilon-01             NFS    50.46 TiB    45.35 TiB    ["esxi-dell-h.rainpole.com","esxi-dell-i.rainpole.com","esxi-dell-l.rainpole.com","esxi-dell-e.rainpole.com","esxi-dell-f.rainpole.com","esxi-dell-k.rainpole.com","esxi-dell-g.rainpole.com","esxi-dell-j.rainpole.com"]
vsan-octo-cluster-a   vsan-OCTO-Cluster-A   vsan   4.37 TiB     2.88 TiB     ["esxi-dell-g.rainpole.com","esxi-dell-f.rainpole.com","esxi-dell-e.rainpole.com"]
vsan-octo-cluster-b   vsan-OCTO-Cluster-B   vsan   2.18 TiB     1.66 TiB     ["esxi-dell-k.rainpole.com","esxi-dell-l.rainpole.com","esxi-dell-j.rainpole.com"]
vsan-octo-cluster-c   vsan-OCTO-Cluster-C   vsan   745.20 GiB   644.56 GiB   ["esxi-dell-h.rainpole.com","esxi-dell-i.rainpole.com"]
```

More information can be returned via the CR status fields, and all fields shown here can be modified to display bespoke information that is of interest to the Kubernetes platform admin. Essentially, any information needed by the K8s platform admin can be added to the CRD and displayed in the main view, or in the status fields. The status field information can be displayed using the -o yaml option. Below, we can see some additional vSphere networking information and tags are associated with the node.

```
% kubectl get nodeinfo k8s-worker-01 -o yaml
apiVersion: topology.vkubeviewer.com/v1
kind: NodeInfo
metadata:
  creationTimestamp: "2021-07-28T12:53:25Z"
  generation: 1
  managedFields:
  - apiVersion: topology.vkubeviewer.com/v1
    fieldsType: FieldsV1
    fieldsV1:
      f:spec:
        .: {}
        f:nodename: {}
      f:status:
        .: {}
        f:attached_tag: {}
        f:net_name: {}
        f:net_overall_status: {}
        f:net_switch_type: {}
        f:path_to_vm: {}
        f:related_cluster: {}
        f:related_datastore: {}
        f:related_host: {}
        f:vm_guest_id: {}
        f:vm_hw_version: {}
        f:vm_ip_address: {}
        f:vm_power_state: {}
        f:vm_total_cpu: {}
        f:vm_total_mem: {}
    manager: manager
    operation: Update
    time: "2021-07-28T12:53:28Z"
  name: k8s-worker-01
  namespace: default
  resourceVersion: "24790893"
  uid: 131f70e7-d6de-460b-b654-0c13005b1828
spec:
  nodename: k8s-worker-01
status:
  attached_tag:
  - zone-a
  net_name: VM Network
  net_overall_status: green
  net_switch_type: Standard
  path_to_vm: '[vsan-OCTO-Cluster-A] 14fe8760-f3fc-92ac-297b-246e962f4854/K8s-Worker-01.vmx'
  related_cluster: OCTO-Cluster-A
  related_datastore:
  - vsan-OCTO-Cluster-A
  related_host: esxi-dell-f.rainpole.com
  vm_guest_id: ubuntu64Guest
  vm_hw_version: vmx-19
  vm_ip_address: 10.27.51.54
  vm_power_state: poweredOn
  vm_total_cpu: 4
  vm_total_mem: 4096
  ```

Additional CRDs have been created to display vSphere Tags, which are used for multi-AZ topology deployments, as well as First Class Disk (FCD) information. FCDs are the objects in vSphere used to back Kubernetes Persistent Volumes when they have been deployed on vSphere storage.

The `FCDInfo` CRD reports back vSphere storage infromation for a Persistent Volume:

```
% kubectl get fcdinfo
NAME   PVID                                       SIZEMB   FILEPATH                                                                                           PROVISIONINGTYPE
fcd0   pvc-1afe9426-88e3-4628-9325-fba1a5b4adfe   4096     [vsan-OCTO-Cluster-C] 13528960-0340-9454-9109-246e962f4ab4/90ada30651be4b42820ed60141413cf2.vmdk   thin
fcd1   pvc-fd644107-77fa-4f86-bf5f-5289618f9295   4096     [vsan-OCTO-Cluster-C] 13528960-0340-9454-9109-246e962f4ab4/12565e24feab4f46873fd7339210e977.vmdk   thin
```

## Guides

[Quick-Start Guide](https://github.com/vKubeViewer/vkubeviewer/blob/main/docs/QuickStartGuide.md) 

[Developer Guide](https://github.com/vKubeViewer/vkubeviewer/blob/main/docs/vKubeViewer%20Guide.md) 


## Current Feature Set

[Host Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/docs/Host%20Information.md)

[Node Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/docs/Node%20Information.md)

[First Class Disk Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/docs/First%20Class%20Disk%20Information.md)

[Tag Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/docs/Tag%20Information.md)

[Data Store Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/docs/Data%20Store%20Information.md)


## Project Contributors

* Cormac Hogan @cormachogan </br>
* Adeoluwa Aderibigbe @adeoluade </br>
* Epifania Sylivester Mhagama @epifaniamhagama </br>
* Jialu Wang @jarrywangcn </br>
* Richard Harris @RichardPHarris </br>
* Shaunak Verma @Shaunak1414
