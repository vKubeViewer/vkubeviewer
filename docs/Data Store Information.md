# Host Infomation 

This document will hold Information on the [Data Store Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/datastoreinfo_controller.go) controller

## Retrievable Infomation 

- Type of Datastore
- Status of Datastore
- Capacity 
- Free space available
- Accessible (Y/N)
- List of Hosts mounted


## API to Retrieve Infomation

These fields are populated Via API calls to the vShere server via [Host Information API](https://github.com/vKubeViewer/vkubeviewer/blob/main/api/v1/datastoreinfo_types.go) as follows

### Set Target Datastore Name

Target datastore name is set in the API call via the following:

```
type DatastoreInfoSpec struct {
	Datastore string `json:"datastore"`
}

```

### Populate Infomation of Target Host

Infomation from the targeted host is populated from the API as follows

```
type DatastoreInfoStatus struct {
	Type         string   `json:"type,omitempty"`
	Status       string   `json:"status,omitempty"`
	Capacity     string   `json:"capacity,omitempty"`
	FreeSpace    string   `json:"free_space,omitempty"`
	Accessible   bool     `json:"accessible,omitempty"`
	HostsMounted []string `json:"hosts_mounted,omitempty"`
}

```


## Controller Populates relevent YAML file

Once the API has retrived the requested infomation, this is then passed to the [controller](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/datastoreinfo_controller.go) where it is sent to the relevent YAML file for subsiquent output:

```
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
			dsinfo.Status.Capacity = ByteCountIEC(ds.Summary.Capacity)
			dsinfo.Status.FreeSpace = ByteCountIEC(ds.Summary.FreeSpace)
			dsinfo.Status.Accessible = ds.Summary.Accessible

			// get the Hosts attached to this datastore, type []types.DatastoreHostMount
			HostMounts := ds.Host
			var curHostsMounted []string

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
				// append the Host's Name into current Hosts List
				curHostsMounted = append(curHostsMounted, h.Summary.Config.Name)
			}

			// if curHostsMounted is different from the one stored in status, replace it

			dsinfo.Status.HostsMounted = UpdateStatusList(dsinfo.Status.HostsMounted, curHostsMounted)

		}
	}

	// ------------
	// Update Session
	// ------------

	// update the status
	if err := r.Status().Update(ctx, dsinfo); err != nil {
		log.Error(err, "unable to update Datastore status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: time.Duration(1) * time.Minute,
	}, nil
}


```
