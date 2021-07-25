# Tag Information

This document will hold information on the [Tag Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/taginfo_controller.go) controller

## Retrievable Infomation 

- The catagory of the Tag
- List of data centers with this tag
- list of clusters with this tag
- list of hosts with this tag
- list of virtual machines with this tag



## API to Retrieve Infomation

These fields are populated Via API calls to the vShere server via [Tag Information API](https://github.com/vKubeViewer/vkubeviewer/blob/main/api/v1/taginfo_types.go) as follows

### Set Target Tag Name

Target tag name is set in the API call via the following:

```
type TagInfoSpec struct {
	Tagname string `json:"tagname,omitempty"`
}


```

### Populate Infomation of Target Host

Infomation from the targeted host is populated from the API as follows

```
type TagInfoStatus struct {
	Category       string   `json:"category,omitempty"`
	DatacenterList []string `json:"datacenter_list,omitempty"`
	ClusterList    []string `json:"cluster_list,omitempty"`
	HostList       []string `json:"host_list,omitempty"`
	VMList         []string `json:"vm_list,omitempty"`
}

```


## Controller Populates relevent YAML file

Once the API has retrived the requested infomation, this is then passed to the [controller](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/taginfo_controller.go) where it is sent to the relevent YAML file for subsiquent output:

```
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

	// tags.NewManager creates a new Manager instance with the rest.Client to retrieve tags information
	tm := tags.NewManager(r.VC_rest)

	// get type tags.Tag with Tagname
	tag, err := tm.GetTag(ctx, taginfo.Spec.Tagname)
	if err != nil {
		msg := fmt.Sprintf("unable to get tags.Tag based on the %s : error %s", taginfo.Spec.Tagname, err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	category, err := tm.GetCategory(ctx, tag.CategoryID)
	if err != nil {
		msg := fmt.Sprintf("unable to get tags.Category based on the %s : error %s", tag.CategoryID, err)
		log.Info(msg)
		return ctrl.Result{}, err
	}
	taginfo.Status.Category = category.Name

	// list ListAttachedObjects with tag.ID
	objs, err := tm.ListAttachedObjects(ctx, tag.ID)
	if err != nil {
		msg := fmt.Sprintf("unable to list attachedobjects on the tag %s : error %s", taginfo.Spec.Tagname, err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	// retrieve the managedobjects with managedobjectreference by property's Retrieve
	pc := property.DefaultCollector(r.VC_vim25)

	// refmap stores the managedobjectreference based on the managedobject type
	refmap := make(map[string][]types.ManagedObjectReference)
	for _, obj := range objs {
		refmap[obj.Reference().Type] = append(refmap[obj.Reference().Type], obj.Reference())
	}

	// define current ManagedObjects list
	var curDatacenterList []string
	var curClusterList []string
	var curHostList []string
	var curVMList []string
	// store the node list via k8s api
	var k8snode = ListK8sNodes()

	// traverse refmap, according its type, retrieve the managedobject and append the name to the ManagedObjects list
	for key, element := range refmap {
		switch key {
		case "Datacenter":
			var dcs []mo.Datacenter
			err = pc.Retrieve(ctx, element, []string{"name"}, &dcs)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve information of %s : error %s", key, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}
			// store name into list
			for _, dc := range dcs {
				curDatacenterList = append(curDatacenterList, dc.Name)
			}
		case "ClusterComputeResource":
			var ccs []mo.ClusterComputeResource
			err = pc.Retrieve(ctx, element, []string{"name"}, &ccs)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve information of %s : error %s", key, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}
			for _, cc := range ccs {
				curClusterList = append(curClusterList, cc.Name)
			}
		case "HostSystem":
			var hss []mo.HostSystem
			err = pc.Retrieve(ctx, element, []string{"name"}, &hss)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve information of %s : error %s", key, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}
			for _, hs := range hss {
				curHostList = append(curHostList, hs.Name)
			}

		case "VirtualMachine":
			var vms []mo.VirtualMachine
			err = pc.Retrieve(ctx, element, nil, &vms)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve information of %s : error %s", key, err)
				log.Info(msg)
				return ctrl.Result{}, err
			}

			// traverse virtual machines
			for _, vm := range vms {
				// RPref - resourcePool Reference
				RPref := vm.ResourcePool
				var resourcepool mo.ResourcePool
				err = pc.RetrieveOne(ctx, *RPref, nil, &resourcepool)
				if err != nil {
					msg := fmt.Sprintf("unable to retrieve ResourcePool MO of %s : error %s", RPref.Value, err)
					log.Info(msg)
					return ctrl.Result{}, err
				}

				// RPref - ClusterComputeResource Reference
				CCRref := resourcepool.Parent
				var clustercomputeresource mo.ClusterComputeResource
				err = pc.RetrieveOne(ctx, *CCRref, nil, &clustercomputeresource)
				if err != nil {
					msg := fmt.Sprintf("unable to retrieve ClusterComputeResource MO of %s : error %s", CCRref.Value, err)
					log.Info(msg)
					return ctrl.Result{}, err
				}

				// check whether virtual machine is a k8s node or not
				if !stringInSlice(vm.Name, k8snode) {
					str := []string{vm.Name, "[", clustercomputeresource.Name, "]"}
					curVMList = append(curVMList, strings.Join(str, " "))
				} else {
					// if the vm is a k8s node, add marker "k8s"
					str := []string{vm.Name, "[", clustercomputeresource.Name, "]", "[ CURRENT ]"}
					curVMList = append(curVMList, strings.Join(str, " "))
				}
			}

		}
	}
	// if current Lists are different from the one stored in status, replace them
	taginfo.Status.DatacenterList = UpdateStatusList(taginfo.Status.DatacenterList, curDatacenterList)
	taginfo.Status.ClusterList = UpdateStatusList(taginfo.Status.ClusterList, curClusterList)
	taginfo.Status.HostList = UpdateStatusList(taginfo.Status.HostList, curHostList)
	taginfo.Status.VMList = UpdateStatusList(taginfo.Status.VMList, curVMList)
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


```

