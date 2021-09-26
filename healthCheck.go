package main

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

//HealthCheck tor circuit checker
func HealthCheck(t []TorStruct) {
	log.WithFields(log.Fields{
		"Len": len(t),
	}).Info("Start check tor circuit")

	for i, v := range t {
		dialSocksProxy, err := proxy.SOCKS5("tcp", *hostNode+":"+v.Port, nil, proxy.Direct)
		if err != nil {
			log.Error(err)
		}
		tr := &http.Transport{
			Dial: dialSocksProxy.Dial,
		}
		_, res, err := Curl(&http.Client{
			Transport: tr,
			Timeout:   10 * time.Second,
		})
		if err != nil || res.StatusCode != http.StatusOK {
			log.WithFields(log.Fields{
				"Error":   err,
				"IP Addr": v.IPAddr,
				"Country": v.Country,
			}).Info("Delete dirty circuit")
			v.DeleteCircuit()
			t = RemoveTorList(t, i)

			log.Info("Request new tor circuit")
			newTor, err := initTor(1)
			if err != nil {
				log.Error(err)
			}
			t = append(t, newTor...)

		} else {
			log.WithFields(log.Fields{
				"IP Addr": v.IPAddr,
				"Country": v.Country,
			}).Info("Circuit OK")
		}
	}
}
