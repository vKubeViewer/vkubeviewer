# vKubeViewer

## About the vKubeViewer project

The vKubeViewer project is a collaborative project from students of University College Cork with the aim to create a Kubernetes operator that will facilitate the retrival of vSphere resource usage infomation from within Kubernetes. 

Upon initial 1.0 release, the project will be open sourced and the Kubernetes community is free to use and update the project with additional features and functionality. 


### Current Version 
Alpha: 1.0

## Problem Statement

The key problem for this project is that Kubernetes has no visibility on the underlying server resource usage information from vSphere. Examples of resource information includes RAM usage, storage usage, and network data usage. This lack of visibility results in users leaving the Kubernetes environment and entering the vSphere environment when attempting to locate this resource usage information. This presents a large time loss for users and organizations. As Kubernetes is designed to manage large amounts of containers, if a user needs to log into vSphere each time they require access to this server recourse usage information, the time loss and subsequent costs can quickly add up for organizations. These costs are most felt by organizations with many Kubernetes users.  

## Usage Example
```
$ kubectl get hostinfo

NAME                             HOSTNAME                         TOTALCPU   FREECPU   TOTALMEMORY   FREEMEMORY   INMAINTENANCEMODE
esxi-dell-e.rainpole.com         esxi-dell-e.rainpole.com         43980      40144     127.91 GiB    59.25 GiB    false
esxi-dell-f.rainpole.com         esxi-dell-f.rainpole.com         43980      42142     127.91 GiB    82.69 GiB    false
esxi-dell-g.rainpole.com         esxi-dell-g.rainpole.com         43980      42419     127.91 GiB    75.74 GiB    false
esxi-dell-h.rainpole.com         esxi-dell-h.rainpole.com         43980      43693     127.91 GiB    110.11 GiB   false
```


```
$ kubectl get nodeinfo 

NAME                  NODENAME              VMTOTALCPU   VMTOTALMEM   VMPOWERSTATE   VMIPADDRESS   VMHWVERSION   CLUSTER          HOST
k8s-controlplane-01   k8s-controlplane-01   4            4096         poweredOn      10.27.51.17   vmx-10        OCTO-Cluster-A   esxi-dell-f.rainpole.com
k8s-worker-01         k8s-worker-01         4            4096         poweredOn      10.27.51.54   vmx-10        OCTO-Cluster-A   esxi-dell-f.rainpole.com
k8s-worker-02         k8s-worker-02         4            4096         poweredOn      10.27.51.25   vmx-10        OCTO-Cluster-B   esxi-dell-k.rainpole.com
k8s-worker-03         k8s-worker-03         4            4096         poweredOn      10.27.51.28   vmx-18        OCTO-Cluster-C   esxi-dell-i.rainpole.com
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
