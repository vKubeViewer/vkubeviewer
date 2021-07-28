# Host Infomation 

This document contains information on the [Host Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/hostinfo_controller.go) controller.

## Retrievable Infomation 

- Total CPU capacity of the host
- Remaining available CPU capacity of the host
- Total Memory capacity of the host 
- Remaining available Memory capacity of the host
- Total storage capacity of the host
- Remaining available storage capacity of the host
- If the host is or is not in maintanace mode.



## API to Retrieve Infomation

These fields are populated Via API calls to the vShere server via [Host Information API](https://github.com/vKubeViewer/vkubeviewer/blob/main/api/v1/hostinfo_types.go) as follows

### Set Target Host Name

Target host name is set in the API call via the following:

```
type HostInfoSpec struct {
	Hostname string `json:"hostname"`
}

```

### Populate Infomation of Target Host

Infomation from the targeted host is populated from the API as follows

```
type HostInfoStatus struct {
	TotalCPU          int64  `json:"total_cpu,omitempty"`
	FreeCPU           int64  `json:"free_cpu,omitempty"`
	TotalMemory       string `json:"total_memory,omitempty"`
	FreeMemory        string `json:"free_memory,omitempty"`
	TotalStorage      string `json:"total_storage,omitempty"`
	FreeStorage       string `json:"free_storage,omitempty"`
	InMaintenanceMode bool   `json:"in_maintenance_mode"`
}

```


## Controller Populates relevent YAML file


Once the API has retrived the requested infomation, this is then passed to the [controller](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/hostinfo_controller.go) where it is sent to the relevent YAML file for subsiquent output:


```
v, err := m.CreateContainerView(ctx, r.VC.ServiceContent.RootFolder, []string{"HostSystem"}, true)

	if err != nil {
		msg := fmt.Sprintf("unable to create container view for HostSystem: error %s", err)
		log.Info(msg)
		return ctrl.Result{}, err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all hosts
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)

	if err != nil {
		msg := fmt.Sprintf("unable to retrieve HostSystem summary: error %s", err)
		log.Info(msg)
		return ctrl.Result{
			RequeueAfter: time.Duration(1) * time.Minute}, err
	}

	// iterate hostsystem and store the host information to HostInfo's status
	for _, hs := range hss {
		if hs.Summary.Config.Name == hi.Spec.Hostname {
			hi.Status.TotalCPU = int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
			hi.Status.FreeCPU = (int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)) - int64(hs.Summary.QuickStats.OverallCpuUsage)
			hi.Status.TotalMemory = ByteCountIEC(hs.Summary.Hardware.MemorySize)
			hi.Status.FreeMemory = ByteCountIEC(int64(hs.Summary.Hardware.MemorySize) - (int64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024))
		}
	}


```

