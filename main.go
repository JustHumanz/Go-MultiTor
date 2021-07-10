package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/elazarl/goproxy"
	"github.com/gorilla/mux"
	muxlogrus "github.com/pytimer/mux-logrus"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

type TorStruct struct {
	TorList *tor.Tor
	Port    string
	IPAddr  string
	Country string
	City    string
}

func (p *TorStruct) AddCountry(new string) *TorStruct {
	p.Country = new
	return p
}

func (p *TorStruct) AddIP(new string) *TorStruct {
	p.IPAddr = new
	return p
}

var torPath = flag.String("p", "/usr/bin/tor", "path of tor binary file")
var torCircuit = flag.Int("c", 10, "total of torCircuit")
var renewIP = flag.Int("i", 10, "duration of tor ip address")
var exitNode = flag.String("e", "", "specific country torCircuit")
var hostNode = flag.String("host", "localhost", "hostname or ip address")
var ProxyPort = flag.String("proxy", "8080", "http proxy port")
var RestAPIPort = flag.String("api", "2525", "rest api prot")

var ifconfig = "https://ipinfo.io"
var PortUsage = 9090
var ipInfoOri IpinfoIo

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

			t, err := tor.Start(nil, &tor.StartConf{
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

			// Create client
			myClient := &http.Client{
				Transport: tr,
				Timeout:   1 * time.Minute,
			}
			res, err := myClient.Get(ifconfig)
			if err != nil {
				log.Error("ignoring node ", err)
				return
			}

			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Error("ignoring node ", err)
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
			"County":     v.Country,
			"City":       v.City,
		}
	}
	return A
}

func (p TorStruct) DeleteCircuit() {
	p.TorList.Close()
}

func init() {
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})

	myNormalClient := &http.Client{}
	res, err := myNormalClient.Get(ifconfig)
	if err != nil {
		log.Error(err)
	}
	defer res.Body.Close()
	bodyNormal, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
	}
	json.Unmarshal(bodyNormal, &ipInfoOri)

}

func main() {
	torList, err := initTor(*torCircuit)
	if err != nil {
		log.Error(err)
	}
	log.WithFields(log.Fields{
		"Tor circuit": len(torList),
	}).Info("Done testing tor circuit")
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			res, err := CurlTor(r.URL.String(), torList[rand.Intn(len(torList))])
			if err != nil {
				log.Error(err)
			}
			return r, res
		})

	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(TortoMap(torList))
	})

	router.HandleFunc("/add/{new}", func(rw http.ResponseWriter, r *http.Request) {
		newreq := mux.Vars(r)["new"]

		reqint, err := strconv.Atoi(newreq)
		if err != nil {
			log.Error(err)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
			return

		}
		log.Info("Request new tor circuit")
		newTor, err := initTor(reqint)
		if err != nil {
			log.Error(err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}

		torList = append(torList, newTor...)

		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(TortoMap(newTor))
	}).Methods(http.MethodPost)

	router.Use(muxlogrus.NewLogger().Middleware)
	go http.ListenAndServe(":"+*RestAPIPort, router)
	go http.ListenAndServe(":"+*ProxyPort, proxy)

	shutdown := make(chan int)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		for _, v := range torList {
			v.DeleteCircuit()
		}
		log.Warn("Shutting down...")
		shutdown <- 1
	}()

	<-shutdown
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
