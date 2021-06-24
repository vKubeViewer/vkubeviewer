# vKubeViewer

## About the vKubeViewer project

The vKubeViewer project is a collaborative project between VMware and University College Cork with the aim to create a Kubernetes operator that will facilitate the retrival of vSphere resource usage infomation from within Kubernetes. 

Upon initial 1.0 release, the project will be open sourced and the Kubernetes community is free to use and update the project with additional features and functionality. 


### Current Version 
Alpha: 1.0

## Problem Statement

The key problem for this project is that Kubernetes has no visibility on the underlying server resource usage information from vSphere. Examples of resource information includes RAM usage, storage usage, and network data usage. This lack of visibility results in users leaving the Kubernetes environment and entering the vSphere environment when attempting to locate this resource usage information. This presents a large time loss for users and organizations. As Kubernetes is designed to manage large amounts of containers, if a user needs to log into vSphere each time they require access to this server recourse usage information, the time loss and subsequent costs can quickly add up for organizations. These costs are most felt by organizations with many Kubernetes users.  

## Expected Project outcomes

- Developed knowledge in GO and govmomi for calling vSphere API
- Create a set of sample govmomi scripts which can be used to retrieve vSphere information



## Current Feature Set

[Host Information - Types](https://github.com/vKubeViewer/vkubeviewer/blob/Richard/docs/HostInformation-Types.md)

[Virtual Machine Information - Types](https://github.com/vKubeViewer/vkubeviewer/blob/Richard/docs/VertualMachineInformation-Types.md)

[First Class Disk Information - Types](https://github.com/vKubeViewer/vkubeviewer/blob/Richard/docs/FirstClassDiskInformation-Types.md)



## Project Contributors

* Cormac Hogan @cormachogan </br>
* Adeoluwa Aderibigbe @adeoluade </br>
* Epifania Sylivester Mhagama @epifaniamhagama </br>
* Jialu Wang @jarrywangcn </br>
* Richard Harris @RichardPHarris </br>
* Shaunak Verma @Shaunak1414
