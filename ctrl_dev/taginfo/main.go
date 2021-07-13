package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vapi/rest"
	_ "github.com/vmware/govmomi/vapi/simulator"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

var (
	//tagger         *faastagger.Client
	//err            error
	ctx                = context.Background()
	vCenterURL         string
	vcUser             string
	vcPass             string
	tagname            string
	insecure           bool
	DatacenterList     []string
	ComputeClusterList []string
	HostList           []string
	VMList             []string
	// DatastoreList      []string
	// NetworkList        []string

)

// func GetStrList(ctx context.Context, element []types.ManagedObjectReference, collect *property.Collector, pro string, dst mo.ManagedEntity) []string {
// 	_ = collect.Retrieve(ctx, element, []string{pro}, &dst)
// 	var lis []string
// 	for _, ds := range dst {
// 		lis = append(lis, dst.Name)
// 	}
// 	return lis
// }

func main() {
	// open reusable connection to vCenter
	// not checking env variables here as faastagger.New would throw error when connecting to VC
	vCenterURL = os.Getenv("GOVMOMI_URL")
	vcUser = os.Getenv("GOVMOMI_USERNAME")
	vcPass = os.Getenv("GOVMOMI_PASSWORD")
	tagname = "zone-a"

	if os.Getenv("GOVMOMI_INSECURE") == "true" {
		insecure = true
	}

	u, err := soap.ParseURL(vCenterURL)
	if err != nil {
		log.Printf("could not parse vCenter client URL: %v", err)

	}

	u.User = url.UserPassword(vcUser, vcPass)
	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		log.Printf("could not get vCenter client: %v", err)

	}

	s := &cache.Session{
		URL:      u,
		Insecure: true,
	}

	vim25client := new(vim25.Client)
	err = s.Login(ctx, vim25client, nil)
	if err != nil {
		log.Printf("could not get vim24 client: %v", err)
	}

	r := rest.NewClient(c.Client)
	err = r.Login(ctx, u.User)

	if err != nil {
		log.Printf("could not get VAPI REST client: %v", err)
	}

	tm := tags.NewManager(r)

	tag, _ := tm.GetTag(ctx, tagname)

	// taglist, _ := tm.GetTagsForCategory(ctx, "k8s-zone")

	refmap := make(map[string][]types.ManagedObjectReference)

	objs, _ := tm.ListAttachedObjects(ctx, tag.ID)
	pc := property.DefaultCollector(vim25client)
	for _, obj := range objs {
		fmt.Println(obj.Reference().Type, obj.Reference().Value)

		refmap[obj.Reference().Type] = append(refmap[obj.Reference().Type], obj.Reference())
	}
	for key, element := range refmap {
		switch key {

		case "Datacenter":
			var dcs []mo.Datacenter
			_ = pc.Retrieve(ctx, element, []string{"name"}, &dcs)
			for _, dc := range dcs {
				DatacenterList = append(DatacenterList, dc.Name)
			}
		case "ClusterComputeResource":
			var ccs []mo.ClusterComputeResource
			_ = pc.Retrieve(ctx, element, []string{"name"}, &ccs)
			for _, cc := range ccs {
				ComputeClusterList = append(ComputeClusterList, cc.Name)
			}
		case "HostSystem":
			var hss []mo.HostSystem
			_ = pc.Retrieve(ctx, element, []string{"name"}, &hss)
			// fmt.Println(hss)
			for _, hs := range hss {
				fmt.Println(hs.Name)
				HostList = append(HostList, hs.Name)
			}
		case "VirtualMachine":
			var vms []mo.VirtualMachine
			_ = pc.Retrieve(ctx, element, []string{"name"}, &vms)
			fmt.Println(vms)
			for _, vm := range vms {
				VMList = append(VMList, vm.Name)
			}
			// case "Network":
			// 	var nets []mo.Network
			// 	_ = pc.Retrieve(ctx, element, []string{"name"}, &nets)
			// 	for _, net := range nets {
			// 		NetworkList = append(NetworkList, net.Name)
			// 	}
			// case "VmwareDistributedVirtualSwitch":
			// 	var dvss []mo.VmwareDistributedVirtualSwitch
			// 	_ = pc.Retrieve(ctx, element, []string{"name"}, &dvss)
			// 	for _, dvs := range dvss {
			// 		NetworkList = append(NetworkList, dvs.Name)
			// 	}

			// case "Datastore":
			// 	var dss []mo.Datastore
			// 	_ = pc.Retrieve(ctx, element, []string{"name"}, &dss)
			// 	for _, ds := range dss {
			// 		DatastoreList = append(DatastoreList, ds.Name)
			// 	}
		}
	}

	fmt.Println("refmap", refmap)

	fmt.Println("Datastore", DatacenterList)
	fmt.Println("ComputeCluster", ComputeClusterList)
	fmt.Println("VM", VMList)

	fmt.Println("Host", HostList)
	// fmt.Println("Datastore", DatastoreList)

	// fmt.Println("Network", NetworkList)
}
