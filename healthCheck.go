package main

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

func HealthCheck(t []TorStruct) {
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
			t = RemoveTorList(t, i)
		} else {
			log.WithFields(log.Fields{
				"IP Addr": v.IPAddr,
				"Country": v.Country,
			}).Info("Circuit OK")
		}
	}
}
