package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cretz/bine/tor"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

func initTor(n int) ([]TorStruct, error) {
	var TorList []TorStruct
	var wg sync.WaitGroup
	for i := PortUsage; i < PortUsage+n; i++ {
		wg.Add(1)

		go func(wg *sync.WaitGroup, j int) {
			defer wg.Done()
			Port := strconv.Itoa(j)
			Time := *renewIP * 60
			Args := []string{"SOCKSPort", Port, "MaxCircuitDirtiness", strconv.Itoa(Time)}
			if *exitNode != "" {
				Args = append(Args, "StrictNodes", "1", "ExitNodes", "{"+*exitNode+"}")
			}

			t, err := tor.Start(context.Background(), &tor.StartConf{
				ExePath:           *torPath,
				ExtraArgs:         Args,
				EnableNetwork:     true,
				RetainTempDataDir: false,
			})
			if err != nil {
				log.Error(err)
				return
			}
			log.WithFields(log.Fields{
				"Args": Args,
			}).Info("Successfully create new tor circuit")

			dialSocksProxy, err := proxy.SOCKS5("tcp", *hostNode+":"+Port, nil, proxy.Direct)
			if err != nil {
				log.Error(err)
				return
			}
			tr := &http.Transport{Dial: dialSocksProxy.Dial}
			body, _, err := Curl(&http.Client{
				Transport: tr,
				Timeout:   1 * time.Minute,
			})
			if err != nil {
				log.Error(err)
				return
			}
			var ipInfo IpinfoIo

			json.Unmarshal(body, &ipInfo)

			log.WithFields(log.Fields{
				"RealIP":  ipInfoOri.IP,
				"TorIP":   ipInfo.IP,
				"Country": ipInfo.Country,
			}).Info("Testing tor circuit")

			TorList = append(TorList, TorStruct{
				TorList: t,
				Port:    Port,
				Country: ipInfo.Country,
				IPAddr:  ipInfo.IP,
				City:    ipInfo.City,
			})
		}(&wg, i)

	}
	wg.Wait()
	PortUsage = PortUsage + n
	return TorList, nil
}

func CurlTor(s string, p TorStruct) (*http.Response, error) {
	dialSocksProxy, err := proxy.SOCKS5("tcp", *hostNode+":"+p.Port, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{Dial: dialSocksProxy.Dial}

	// Create client
	myClient := &http.Client{
		Transport: tr,
		Timeout:   2 * time.Minute,
	}
	res, err := myClient.Get(s)
	if err != nil {
		return nil, err
	}
	return res, nil

}

func TortoMap(p []TorStruct) map[int]interface{} {
	A := make(map[int]interface{})
	for i, v := range p {
		A[i] = map[string]interface{}{
			"IP Address": v.IPAddr,
			"Socks5":     *hostNode + ":" + v.Port,
			"Country":    v.Country,
			"City":       v.City,
		}
	}
	return A
}

func (p TorStruct) DeleteCircuit() {
	p.TorList.Close()
}

func RemoveTorList(s []TorStruct, index int) []TorStruct {
	return append(s[:index], s[index+1:]...)
}

var Counter = 0

//GetTorLB Get one tor slice
func GetTorLB(i []TorStruct) *TorStruct {
	Circuit := i[Counter]
	Counter++

	if Counter >= (len(i)) {
		Counter = 0
	}
	return &Circuit
}

type IpinfoIo struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
}

func Curl(c *http.Client) ([]byte, *http.Response, error) {
	res, err := c.Get(ifconfig)
	if err != nil {
		return nil, nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	return body, res, nil
}
