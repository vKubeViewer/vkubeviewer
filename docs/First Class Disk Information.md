
# First Class Disk Information

This document will hold Information on the [First Class Disk Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/fcdinfo_controller.go) controller.

## Retrievable Infomation 

- Size of the FCD in MB
- File path to the FCD
- The Provisioning Type of the FCD



## API to Retrieve Infomation

These fields are populated Via API calls to the vShere server via [First Class Disk Information API](https://github.com/vKubeViewer/vkubeviewer/blob/main/api/v1/fcdinfo_types.go) as follows

### Set Persistant Volume Name

Target persistant volume name is set in the API call via the following:

```
type FCDInfoSpec struct {
	PVId string `json:"pvId"`
}

```

### Populate Infomation of Target Host

Infomation from the targeted host is populated from the API as follows

```
type FCDInfoStatus struct {
	SizeMB           int64  `json:"sizeMB"`
	FilePath         string `json:"filePath"`
	ProvisioningType string `json:"provisioningType"`
}


```


## Controller Populates relevent YAML file

Once the API has retrived the requested infomation, this is then passed to the [controller](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/fcdinfo_controller.go) where it is sent to the relevent YAML file for subsiquent output:

```
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
	var k8spv = ListK8sPV()

	for _, pv := range k8spv {
		spec := vslmtypes.VslmVsoVStorageObjectQuerySpec{
			QueryField:    "name",
			QueryOperator: "contains",
			QueryValue:    []string{pv},
		}
		query = append(query, spec)
	}
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

	return ctrl.Result{RequeueAfter: time.Duration(1) * time.Minute}, nil
}


```