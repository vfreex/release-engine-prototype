package main

import (
	"fmt"
	"github.com/kolo/xmlrpc"
)

func main() {
	rpc, _ := xmlrpc.NewClient("https://brewhub.engineering.redhat.com/brewhub", nil)
	result := 0
	rpc.Call("getAPIVersion", nil, &result)
	fmt.Printf("Version: %v\n", result)

	nvr := "openshift-4.2.10-201911290432.git.0.888f9c6.el7"

	var buildinfo struct {
		BuildId     int    `xmlrpc:"build_id"`
		OwnerName   string `xmlrpc:"owner_name"`
		PackageName string `xmlrpc:"package_name"`
		State       int    `xmlrpc:"state"`
		Nvr         string `xmlrpc:"nvr"`
		Version     string `xmlrpc:"version"`
		Release     string `xmlrpc:"release"`
		Epoch       string `xmlrpc:"epoch"`
		Extra       struct {
			Source struct {
				OriginalUrl string `xmlrpc:"original_url"`
			} `xmlrpc:"source"`
		} `xmlrpc:"extra"`
	}

	rpc.Call("getBuild", nvr, &buildinfo)
	fmt.Printf("buildinfo: %v\n", buildinfo)

	args := []interface{}{
		"rhaos-4.2-rhel-7",
		struct {
			Starstar bool   `xmlrpc:"__starstar"`
			Package  string `xmlrpc:"package"`
		}{true, "openshift"},
	}
	var result2 [] struct {
		BuildId     int    `xmlrpc:"build_id"`
		OwnerName   string `xmlrpc:"owner_name"`
		PackageName string `xmlrpc:"package_name"`
		State       int    `xmlrpc:"state"`
		Nvr         string `xmlrpc:"nvr"`
		Version     string `xmlrpc:"version"`
		Release     string `xmlrpc:"release"`
	}
	rpc.Call("getLatestBuilds", args, &result2)
	fmt.Printf("result2: %v\n", result2)
}
