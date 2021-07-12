package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/embano1/faastagger"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
	_ "github.com/vmware/govmomi/vapi/simulator"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/soap"
)

var (
	tagger     *faastagger.Client
	err        error
	ctx        = context.Background()
	vCenterURL string
	vcUser     string
	vcPass     string
	tagID      string
	insecure   bool
)

func main() {
	// open reusable connection to vCenter
	// not checking env variables here as faastagger.New would throw error when connecting to VC
	vCenterURL = os.Getenv("GOVMOMI_URL")
	vcUser = os.Getenv("GOVMOMI_USERNAME")
	vcPass = os.Getenv("GOVMOMI_PASSWORD")
	tagID = "zone-a"

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

	r := rest.NewClient(c.Client)
	err = r.Login(ctx, u.User)

	if err != nil {
		log.Printf("could not get VAPI REST client: %v", err)
	}

	tm := tags.NewManager(r)

	tag, _ := tm.GetTag(ctx, tagID)

	taglist, _ := tm.GetTagsForCategory(ctx, "k8s-zone")

	fmt.Println(tag.ID, tag.Description, tag.Name, tag.UsedBy)
	for _, t := range taglist {
		fmt.Println(t.Name, t.ID)
		objs, _ := tm.ListAttachedObjects(ctx, tagID)

		for _, obj := range objs {
			fmt.Println(obj.Reference().Type, obj.Reference().Value)
		}
	}
}
