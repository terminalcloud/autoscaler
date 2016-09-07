package autoscaler

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type AdminAPI interface {
	CreateHN() (string, error)
	DeleteHN(ip string) (string, error)
	GetHNInfo(filter string) ([]*NodeInfo, int, int, int, error)
	ReqProgress(token string) (*Progress, error)
}

type adminAPI struct {
	clusterApi string
	utoken     string
	atoken     string
}

func NewAdminAPIHttp(base string, ut string, at string) AdminAPI {
	aa := adminAPI{base, ut, at}
	return &aa
}

func (aa *adminAPI) CreateHN() (string, error) {
	path := "create_hn"
	command := "POST"

	var jsonStr = []byte(`{"instance": "` + *nodeType + `", "storage": "` + *nodeStorage + `" }`)

	out, err := aa.callAPI(command, path, jsonStr)
	if err != nil {
		return "", err
	}

	var rt RequestToken
	err = json.Unmarshal([]byte(out), &rt)
	if err != nil {
		return "", err
	}

	return rt.Request_id, nil

}

func (aa *adminAPI) DeleteHN(ip string) (string, error) {
	path := "delete_hn"
	command := "POST"
	var jsonStr = []byte(`{"id": "` + ip + `" }`)

	out, err := aa.callAPI(command, path, jsonStr)
	if err != nil {
		return "", err
	}

	var rt RequestToken
	err = json.Unmarshal([]byte(out), &rt)
	if err != nil {
		return "", err
	}

	return rt.Request_id, nil
}

func (aa *adminAPI) GetHNInfo(filter string) ([]*NodeInfo, int, int, int, error) {
	path := "get_hn_health"
	command := "POST"
	var jsonStr = []byte(`{"filter": "` + filter + `" }`)

	//Json Return
	var retJson map[string]interface{}
	var nodelist []*NodeInfo
	var pendingC = 0
	var pendingD = 0
	var pendingR = 0

	out, err := aa.callAPI(command, path, jsonStr)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	//Parse node struct
	err = json.Unmarshal([]byte(out), &retJson)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	for k, v := range retJson {

		if k == "pendingCreates" {
			pendingItems, isCorrectType := v.([]interface{})
			if isCorrectType {
				pendingC = len(pendingItems)
			} else {
				log.Info("No Pending Create Tokens")
			}
		}

		if k == "pendingDeletes" {
			pendingItems, isCorrectType := v.([]interface{})
			if isCorrectType {
				pendingD = len(pendingItems)
			} else {
				log.Info("No Pending Create Tokens")
			}
		}

		if k == "requestQueue" {
			requestQueue, isCorrectType := v.([]interface{})
			if isCorrectType {
				pendingR = len(requestQueue)
			} else {
				log.Info("No Pending Create Tokens")
			}
		}

		// Parse out all the node info
		if k == "nodes" {

			nodeInfo, isCorrectType := v.(map[string]interface{})
			if !isCorrectType {
				fmt.Println("error parsing level 1")
				return nil, 0, 0, 0, errors.New("bad json")
			}

			for ip, y := range nodeInfo {

				nodeInfoItems, isCorrectType := y.(map[string]interface{})
				if !isCorrectType {
					fmt.Println("errorl parsing level 2")
					return nil, 0, 0, 0, errors.New("bad json")
				}

				//Create a new Node and populate Info
				newNode := &NodeInfo{}
				newNode.NodeIp = ip

				for info, value := range nodeInfoItems {
					if info == "ram" {
						ramInfo, isCorrectType := value.(map[string]interface{})
						if !isCorrectType {
							log.Error("Invalid Ram struct node")
							return nil, 0, 0, 0, errors.New("bad json")
						}

						used, isOku := ramInfo["used"].(string)
						total, isOkh := ramInfo["total"].(string)

						if !isOku || !isOkh {
							log.Error("Invalid Ram numbers for node")
							return nil, 0, 0, 0, errors.New("bad json")
						}

						u, errInt := strconv.ParseInt(used, 10, 64)
						if errInt != nil {
							u = 0
						}
						t, errInt := strconv.ParseInt(total, 10, 64)
						if errInt != nil {
							t = 0
						}
						newNode.RamTotal = t
						newNode.RamUsed = u
					}
					if info == "containers" {
						ctInfo, isCorrectType := value.(map[string]interface{})
						if !isCorrectType {
							log.Error("Invalid containers struct node")
							return nil, 0, 0, 0, errors.New("bad json")
						}

						total, isOkh := ctInfo["total"].(string)

						if !isOkh {
							log.Error("Invalid containers numbers for node")
							return nil, 0, 0, 0, errors.New("bad json")
						}

						t, errInt := strconv.ParseInt(total, 10, 64)
						if errInt != nil {
							t = 0
						}
						newNode.ContainersTotal = t
					}
					if info == "status" {

						status, isCorrectType := value.(map[string]interface{})
						if !isCorrectType {
							log.Error("Bad status JSON")
							return nil, 0, 0, 0, errors.New("bad json")
						}

						disabled, isOkd := status["disabled"].(bool)
						terminating, isOkt := status["terminate"].(bool)
						notreporting, isOkr := status["notReporting"].(bool)
						protected, isOkp := status["protected"].(bool)

						//Because the Cluster API is broken this hack will default it to false
						//This is because no value is sent if false
						statusObj := StatusInfo{false, false, false, false, false}

						if isOkd {
							statusObj.Disabled = disabled
						}
						if isOkt {
							statusObj.Terminate = terminating
						}
						if isOkr {
							statusObj.NotReporting = notreporting
						}
						if isOkp {
							statusObj.Protected = protected
						}

						newNode.Status = statusObj
					}
				}
				nodelist = append(nodelist, newNode)
			}
		}
	}

	return nodelist, pendingC, pendingD, pendingR, nil
}

func (aa *adminAPI) ReqProgress(token string) (*Progress, error) {
	path := "request_progress"
	command := "POST"

	var jsonStr = []byte(`{"request_id": "` + token + `" }`)
	out, err := aa.callAPI(command, path, jsonStr)

	var rp Progress
	err = json.Unmarshal([]byte(out), &rp)
	if err != nil {
		return nil, err
	}

	return &rp, nil
}

func (aa *adminAPI) callAPI(command string, url string, jsoncontent []byte) (string, error) {

	fmt.Println(aa.clusterApi + url)

	req, err := http.NewRequest(command, aa.clusterApi+url, bytes.NewBuffer(jsoncontent))
	req.Header.Set("user-token", aa.utoken)
	req.Header.Set("access-token", aa.atoken)
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Request Error")
		fmt.Println(err)
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("Got error : ", resp.Status)
		return "", errors.New(string(resp.Status))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body), err
}
