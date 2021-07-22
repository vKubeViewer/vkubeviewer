This document contains information on the [Host Information](https://github.com/vKubeViewer/vkubeviewer/blob/main/controllers/hostinfo_controller.go) controller.


# Host Infomation 

## Retrievable Infomation 

- Total CPU capacity of the host
- Remaining available CPU capacity of the host
- Total Memory capacity of the host 
- Remaining available Memory capacity of the host

These fields are populated Via API calls to the vShere server via [Host Information API](https://github.com/vKubeViewer/vkubeviewer/blob/Richard/api/v1/hostinfo_types.go) as follows

```
type HostInfoStatus struct {
	TotalCPU    int64  `json:"total_cpu,omitempty"`
	FreeCPU     int64  `json:"free_cpu,omitempty"`
	TotalMemory string `json:"total_memory,omitempty"`
	FreeMemory  string `json:"free_memory,omitempty"`
}

```
