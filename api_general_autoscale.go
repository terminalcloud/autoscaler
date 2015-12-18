package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

type AutoScaler interface {
	Run()
}

type autoscaler struct {
	api                   AdminAPI
	pendingCreateToken    []string
	pendingDelToken       []string
	iterationsSinceCreate int
	fatalErrorCount       int
	filter                string
}

func NewGeneralAutoScaler(aa AdminAPI, filter string) AutoScaler {
	pdt := []string{}
	pct := []string{}
	if filter == "" {
		panic("INVALID FILTER TYPE")
	}
	as := &autoscaler{aa, pdt, pct, 0, 0, filter}
	return as
}

func (as *autoscaler) Run() {

	//Update Progress of things
	as.pendingCreateToken = TrackProgress(as.pendingCreateToken, as.api)
	as.pendingDelToken = TrackProgress(as.pendingDelToken, as.api)

	out, pendC, pendD, pendR, err := as.api.GetHNInfo(as.filter)
	if err == nil {
		//Scale Cluster
		if as.filter == "default" {
			as.autoScale(out, pendC, pendD)
		} else if as.filter == "gpu" {
			as.autoScaleGPU(out, pendC, pendD, pendR)
		} else {
			log.Error("Invalid Filter, no policy found for this filter")
		}
	} else {
		log.Error(err)
	}

}

func (as *autoscaler) autoScale(out []*NodeInfo, pendC int, pendD int) {

	var globtotal int64
	var globused int64
	numActive := 0

	for _, e := range out {
		if !e.Status.Disabled && !e.Status.Terminate {
			globtotal = globtotal + e.RamTotal
			globused = globused + e.RamUsed
			numActive++
		}
	}

	log.Info("Pending Create : ", pendC, "&  Del tokens : ", pendD)
	log.Info("Number of Active Nodes : ", numActive)

	//Addd pending so we dont over provision
	globtotal = globtotal + (int64(pendC) * *nodeTypeRamTotal)

	//Handle the case of 0 resources
	if globtotal == 0 {
		log.Info("Global 0 so creating 1")
		if request_id, err := as.api.CreateHN(); err == nil {
			as.pendingCreateToken = append(as.pendingCreateToken, request_id)
			as.iterationsSinceCreate = 0
		}
	} else {

		should_grow := false
		should_shrink := false

		ratio := float64(globused) / float64(globtotal)

		log.Info("Pressure ratio: ", ratio)

		if ratio > 0.80 {
			log.Info("Growing cluster!")
			should_grow = true
		}

		if globtotal > *nodeTypeRamTotal {
			// imagine the ratio after shrinking
			post_shrink_ratio := float64(globused) / float64(globtotal-*nodeTypeRamTotal)

			log.Info("Post-shrink ram pressure ratio: ", post_shrink_ratio)
			if post_shrink_ratio < 0.70 {
				if as.iterationsSinceCreate >= min_iters_before_shrink {
					fmt.Println("Shrinking cluster!")
					should_shrink = true
				} else {
					log.Info("Would shrink but need to wait iters:", min_iters_before_shrink)
				}
			}
		}

		as.performScaling(should_shrink, should_grow, out)
	}
}

func (as *autoscaler) autoScaleGPU(out []*NodeInfo, pendC int, pendD int, pendR int) {

	var globfree int
	for _, e := range out {
		if !e.Status.Disabled && !e.Status.Terminate && e.RamUsed < 1000 {
			globfree = globfree + 1
		}
	}

	needed := pendR - (globfree + pendC)
	log.Info("New Nodes Needed :", needed, "Pending Node Reqs:", pendC)

	if needed > 0 {
		for i := 1; i <= needed; i++ {
			as.performScaling(false, true, out)
		}
	} else {
		should_grow := false
		should_shrink := false

		//Shrink anything that is unused
		if globfree >= 1 {
			if as.iterationsSinceCreate >= min_iters_before_shrink {
				fmt.Println("Shrinking cluster!")
				should_shrink = true
			} else {
				log.Info("Would shrink but need to wait iters:", min_iters_before_shrink)
			}
		}
		as.performScaling(should_shrink, should_grow, out)
	}
}

func (as *autoscaler) performScaling(should_shrink bool, should_grow bool, nodes []*NodeInfo) {
	//Custom AutoScale Policy Logic Will come Here
	if should_grow && should_shrink {
		log.Error("Wtf dont know what to do.  growing and shrinking at the same time?")
	} else if should_grow {
		//Call to Create
		if request_id, err := as.api.CreateHN(); err == nil && request_id != "" {
			as.pendingCreateToken = append(as.pendingCreateToken, request_id)
			as.iterationsSinceCreate = 0
		} else {
			log.Error("Error: In Creating Node")
		}
	} else if should_shrink {
		//Call To Call to delete
		if len(nodes) > 1 {
			var target string
			if as.filter == "gpu" {
				target = as.pickUnusedNotProtected(nodes)
			} else {
				target = as.pickLeastLoadedNotProtected(nodes)
			}
			if target != "" {
				if request_id, err := as.api.DeleteHN(target); err == nil && request_id != "" {
					as.pendingDelToken = append(as.pendingDelToken, request_id)
				} else {
					log.Error("Error: In Deleting Node")
				}
			}
		}
	} else {
		as.iterationsSinceCreate = as.iterationsSinceCreate + 1
		log.Info("Iterations since last create: ", as.iterationsSinceCreate)
	}
}

func (as *autoscaler) pickLeastLoadedNotProtected(nodes []*NodeInfo) string {
	minip := nodes[0].NodeIp
	minram := nodes[0].RamUsed
	for _, e := range nodes {
		if e.RamUsed < minram && e.Status.Disabled == false && e.Status.Protected == false {
			minram = e.RamUsed
			minip = e.NodeIp
		}
	}
	return minip
}

func (as *autoscaler) pickUnusedNotProtected(nodes []*NodeInfo) string {
	for _, e := range nodes {
		if e.RamUsed < 1000 && e.Status.Disabled == false && e.Status.Protected == false {
			return e.NodeIp
		}
	}
	log.Info("No Available Nodes To Shrink")
	return ""
}
