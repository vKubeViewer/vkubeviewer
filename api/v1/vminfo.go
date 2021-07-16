
```
// VMInfoStatus defines the observed state of VMInfo
type VMInfoStatus struct {
        GuestId    string `json:"guestId"`
        TotalCPU   int64  `json:"totalCPU"`
        ResvdCPU   int64  `json:"resvdCPU"`
        TotalMem   int64  `json:"totalMem"`
        ResvdMem   int64  `json:"resvdMem"`
        PowerState string `json:"powerState"`
        HwVersion  string `json:"hwVersion"`
        IpAddress  string `json:"ipAddress"`
        PathToVM   string `json:"pathToVM"`
}
```
