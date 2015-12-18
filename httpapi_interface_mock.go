package main

import (
	"github.com/satori/go.uuid"
	"math/rand"
	"strconv"
	"fmt"
)

type AdminAPIMock struct {
	internalInfo map[string]NodeInfo
	internalReq  map[string]Fakeplog
	Health       []*NodeInfo
}

type Fakeplog struct {
	status  string
	purpose string
	ip      string
}

func NewAdminAPIHttpMock() AdminAPI {
	ii := make(map[string]NodeInfo)
	ir := make(map[string]Fakeplog)
	h := []*NodeInfo{}

	aam := AdminAPIMock{ii, ir, h}

	return &aam
}

func (aa *AdminAPIMock) CreateHN() (string, error) {
	u1 := uuid.NewV4().String()
	aa.internalReq[u1] = Fakeplog{"pending", "create", ""}
	return u1, nil
}

func (aa *AdminAPIMock) DeleteHN(ip string) (string, error) {
	u1 := uuid.NewV4().String()
	aa.internalReq[u1] = Fakeplog{"pending", "delete", ip}
	return u1, nil
}

func (aa *AdminAPIMock) GetHNInfo(filter string) ([]*NodeInfo,int,int,int, error) {
	return aa.Health, 0,0,0,nil
}

func (aa *AdminAPIMock) ReqProgress(token string) (*Progress, error) {
	fl := aa.internalReq[token]
	p := Progress{fl.status,""}
	return &p, nil
}

func (aa *AdminAPIMock) completeCreates() {
	for k, v := range aa.internalReq {

		//Pending -> Created
		if v.status == "pending" && v.purpose == "create" {
			v.status = "success"
			iplast := rand.Intn(200)
			ip := "10.0.0." + strconv.Itoa(iplast)
			newNode := NodeInfo{ip, *nodeTypeRamTotal, 1000, StatusInfo{false,false,false,false,false}}

			//Adding new Node
			aa.Health = append(aa.Health, &newNode)
			aa.internalInfo[ip] = newNode
		}
		aa.internalReq[k] = v
	}
}

func (aa *AdminAPIMock) completeDeletes() {
	for k, v := range aa.internalReq {
		//Pending -> Deleted
		if v.status == "pending" && v.purpose == "delete" {
			v.status = "success"
			delete(aa.internalInfo, v.ip)
			aa.rebuildHealthMinus(v.ip)
			fmt.Println(aa.Health)
		}
		aa.internalReq[k] = v
	}
}

func (aa *AdminAPIMock) rebuildHealthMinus(ip string) {
	var newHealth []*NodeInfo
	for _, e := range aa.Health {
		if e.NodeIp != ip {
			newHealth = append(newHealth, e)
		}
	}
	aa.Health = newHealth
}

func (aa *AdminAPIMock) modifyLoad(ip string, value int64) {
	var newHealth []*NodeInfo
	for _, e := range aa.Health {
		if e.NodeIp == ip {
			e.RamUsed = value
			newHealth = append(newHealth, e)
		}
	}
	aa.Health = newHealth
}

func (aa *AdminAPIMock) populateHealth(load int64, numNodes int) {

	newHealth := []*NodeInfo{}
	newInternalInfo := make(map[string]NodeInfo)

	for i := 0; i < numNodes; i++ {
		iplast := rand.Intn(200)
		ip := "10.0.0." + strconv.Itoa(iplast)
		newNode := NodeInfo{ip, *nodeTypeRamTotal, load, StatusInfo{false,false,false,false,false}}
		newHealth = append(newHealth, &newNode)
		newInternalInfo[ip] = newNode
	}

	aa.Health = newHealth
	aa.internalInfo = newInternalInfo
}
