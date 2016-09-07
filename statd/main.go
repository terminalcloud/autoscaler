package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/terminalcloud/autoscaler"
)

var aa autoscaler.AdminAPI

func GetRedirectUrl() (string, error) {
	nodes, _, _, _, err := aa.GetHNInfo("default")
	if err != nil {
		return "", err
	}
	url := "http://lablono.com/render?width=1600&height=800&hideLegend=false&from=-1day&target=movingAverage(stats.gauges.10.0.0.201,%221hr%22)"
	for _, v := range nodes {
		url += fmt.Sprintf(`&target=movingAverage(stats.gauges.%s,"1hr")`, v.NodeIp)
	}
	return url, nil
}

func Director(req *http.Request) {
	redurl, err := GetRedirectUrl()
	if err != nil {
		return
	}
	u, err := url.Parse(redurl)
	req.URL = u
}

// keep updated the uptime of servers
// if a server doesn't appear, removed from list
// if it appears, it is there, and recorded with first time seen

var mut = sync.Mutex{}
var uptimes = map[string]time.Time{}

func updateUptimes() (map[string]time.Time, bool) {
	mut.Lock()
	defer mut.Unlock()
	nodes, _, _, _, err := aa.GetHNInfo("default")
	if err != nil {
		log.Println(err)
		return uptimes, false
	}

	now := map[string]bool{}
	for _, v := range nodes {
		now[v.NodeIp] = true
		if _, ok := uptimes[v.NodeIp]; !ok {
			uptimes[v.NodeIp] = time.Now()
		}
	}
	for k, _ := range uptimes {
		if _, ok := now[k]; !ok {
			delete(uptimes, k)
		}
	}
	return uptimes, true
}

func main() {
	autoscaler.Configure()
	aa = autoscaler.AdminClient()
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		url, err := GetRedirectUrl()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	http.Handle("/graph.png", &httputil.ReverseProxy{Director: Director})
	http.HandleFunc("/uptimes", func(w http.ResponseWriter, r *http.Request) {
		ups, _ := updateUptimes()
		now := time.Now()
		for k, v := range ups {
			fmt.Fprintf(w, "%s = %s = %d minutes ago\n", k, v, now.Sub(v)/time.Minute)
		}
	})
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			updateUptimes()
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
