package autoscaler

import (
	"flag"
	"fmt"
	"time"
)

//Base API Url
var base *string

//Node Type Info
var nodeTypeRamTotal *int64
var nodeType *string
var nodeStorage *string

//Scale Policy
var min_time_before_shrink *int64
var autoscale_freq *int64

// min_time_before_shrink / autoscale_freq
var min_iters_before_shrink int

//Auth
var usertoken *string
var accessstoken *string
var apiurl *string
var policy *string

//Failsafe counter
var globalFail int

type NodeInfo struct {
	NodeIp          string
	RamTotal        int64
	RamUsed         int64
	ContainersTotal int64
	Status          StatusInfo
}

type StatusInfo struct {
	Pending      bool `json:",omitempty"`
	Disabled     bool `json:",omitempty"`
	Protected    bool `json:",omitempty"`
	Terminate    bool `json:",omitempty"`
	NotReporting bool `json:",omitempty"`
}

type RamInfo struct {
	Total int `json:"total"`
	Used  int
	Free  int
	Held  int
}

type RequestToken struct {
	Request_id string `json:"request_id"`
}

type Progress struct {
	Status string `json:"status"`
	State  string `json:"state"`
}

// Helper Functions //

func TrackProgress(tokenArray []string, api AdminAPI) []string {
	var remainingItems []string
	for _, e := range tokenArray {
		rp, err := api.ReqProgress(e)
		if err != nil || rp.Status == "failed" {
			//Remove the item from Array and Slack
			fmt.Println("Node Creation Failed")
			globalFail = globalFail + 1
			if globalFail > 4 {
				panic("To many Failed Starts")
			}
		} else if rp.Status != "success" {
			remainingItems = append(remainingItems, e)
		}
	}
	return remainingItems
}

func (as *autoscaler) HandleFatalAPIFailure() {
	//Increatement  + Check + Fail if needed
}

// Main CLI //
func Configure() {

	//User Auth Infromation
	usertoken = flag.String("usertoken", "", "Terminal Cluster user token")
	accessstoken = flag.String("accesstoken", "", "Terminal Cluster user token")

	apiurl = flag.String("apiurl", "", "Terminal Cluster API Path eg foo.com/api/v0.2/")

	//Default Node type Infromation
	nodeType = flag.String("nodetype", "m4.large", "Name for node type to spin up, e.g. m4.large")
	nodeTypeRamTotal = flag.Int64("nodetyperam", 7834, "Amount of ram available on an empty node, e.g. 7834")
	nodeStorage = flag.String("nodestorage", "ebs", "Storage type for nodes to spin up (ebs or ephemeral)")

	min_time_before_shrink = flag.Int64("tts", 3600, "Minimum time after your cluster grows, in seconds, before shrinking (e.g. if on AWS might as well be 1 hour)")
	autoscale_freq = flag.Int64("frequency", 5, "Amount of time to wait between loop of polling cluster state, in seconds")

	policy = flag.String("policy", "general", "The type of autoscaler you want to run general / gpu / other")

	flag.Parse()

	// NOTE: this must be after flag.Parse()!
	min_iters_before_shrink = int(*min_time_before_shrink / *autoscale_freq)

	if *apiurl == "" {
		panic("Invalid Base API Url")
	}

}

func StartAutoscaler() {
	aa := NewAdminAPIHttp(*apiurl, *usertoken, *accessstoken)

	var auto AutoScaler

	if *policy == "general" {
		auto = NewGeneralAutoScaler(aa, "default")
	} else if *policy == "gpu" {
		auto = NewGeneralAutoScaler(aa, "gpu")
	} else {
		panic("Invalid policy for autoscale")
	}

	for true {
		auto.Run()
		time.Sleep(time.Duration(*autoscale_freq) * time.Duration(time.Second))
	}
}

func AdminClient() AdminAPI {
	return NewAdminAPIHttp(*apiurl, *usertoken, *accessstoken)
}
