package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/gorilla/mux"
	muxlogrus "github.com/pytimer/mux-logrus"

	log "github.com/sirupsen/logrus"
)

type TorStruct struct {
	TorList *tor.Tor
	Port    string
	IPAddr  string
	Country string
	City    string
	Load    int
}

func (p *TorStruct) AddCountry(new string) *TorStruct {
	p.Country = new
	return p
}

func (p *TorStruct) AddIP(new string) *TorStruct {
	p.IPAddr = new
	return p
}

var torPath = flag.String("tor", "/usr/bin/tor", "path of tor binary file")
var torCircuit = flag.Int("circuit", 10, "total of torCircuit")
var renewIP = flag.Int("lifespan", 10, "duration of tor ip address")
var exitNode = flag.String("exitnode", "", "specific country torCircuit")
var hostNode = flag.String("host", "0.0.0.0", "hostname or ip address")
var RestAPIPort = flag.String("api", "2525", "rest api port")
var socksLBPort = flag.String("lb", "1412", "socks5 load balancing port")
var LBalgo = flag.String("lbalgp", "rr", "choice algorithm for loadbalancing,rr(round robin)&lc(least connetion)")
var ifconfig = "https://ipinfo.io"
var PortUsage = 9090
var ipInfoOri IpinfoIo
var acessKey = flag.String("key", "", "add api key,if key empty key will be created")

var who, _ = base64.StdEncoding.DecodeString(img)

func init() {
	flag.Parse()
	Secret := ""
	if *acessKey == "" {
		const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789=-_"
		Secret = func(n int) string {
			b := make([]byte, n)
			for i := range b {
				b[i] = letterBytes[rand.Intn(len(letterBytes))]
			}
			return string(b)
		}(36)

	} else {
		Secret = *acessKey
	}

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})
	log.Info("Aceess key ", Secret, " add this key into http header as 'access_key'")

	bodyNormal, _, err := Curl(&http.Client{})
	if err != nil {
		log.Fatalln(err)
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

	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(map[string]interface{}{
			"Date": time.Now(),
			"Proxy": map[string]interface{}{
				"Socks5": "socks5://" + *hostNode + ":" + *socksLBPort,
			},
		})
	})

	router.HandleFunc("/info", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(TortoMap(torList))
	})

	router.HandleFunc("/add/{new:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access_key") == "" {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(who)
			return
		}

		newreq := mux.Vars(r)["new"]
		reqint, err := strconv.Atoi(newreq)
		if err != nil {
			log.Error(err)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
			return

		}
		Create := func() []TorStruct {
			log.Info("Request new tor circuit")
			newTor, err := initTor(reqint)
			if err != nil {
				log.Error(err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(err.Error()))
				return nil
			}

			torList = append(torList, newTor...)
			return torList
		}
		if reqint > 10 {
			go Create()
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(map[string]interface{}{
				"Status": http.StatusOK,
			})
			return
		} else {
			newTor := Create()
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(TortoMap(newTor))
		}

	}).Methods(http.MethodPost)

	Delete := router.PathPrefix("/delete").Subrouter()
	Delete.HandleFunc("/ip/{ip}", func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access_key") == "" {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(who)
			return
		}

		newreq := strings.Split(mux.Vars(r)["ip"], ",")
		log.Info("Delete tor circuit by ip")
		for i, v := range torList {
			for _, v2 := range newreq {
				if v.IPAddr == v2 {
					torList = RemoveTorList(torList, i)
				}
			}
		}

		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(TortoMap(torList))
	}).Methods(http.MethodPost)

	Delete.HandleFunc("/ip/country/{country}", func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access_key") == "" {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(who)
			return
		}

		newreq := strings.Split(mux.Vars(r)["country"], ",")
		log.Info("Delete tor circuit by country")
		for i, v := range torList {
			for _, v2 := range newreq {
				if v.Country == v2 {
					torList = RemoveTorList(torList, i)
				}
			}
		}

		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(TortoMap(torList))
	}).Methods(http.MethodPost)

	Delete.HandleFunc("/ip/city/{city}", func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access_key") == "" {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(who)
			return
		}

		newreq := strings.Split(mux.Vars(r)["city"], ",")
		log.Info("Delete tor circuit by city")
		for i, v := range torList {
			for _, v2 := range newreq {
				if v.City == v2 {
					torList = RemoveTorList(torList, i)
				}
			}
		}

		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(TortoMap(torList))
	}).Methods(http.MethodPost)

	router.Use(muxlogrus.NewLogger().Middleware)

	go func() {
		listener, err := net.Listen("tcp", *hostNode+":"+*socksLBPort)
		if err != nil {
			panic(err)
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Error("error accepting connection ", err)
				continue
			}
			go func() {
				TorCir := func() *TorStruct {
					if *LBalgo == "lc" {
						return GetTorLBWeight(torList)
					} else {
						return GetTorLB(torList)
					}
				}()

				addr := "localhost:" + TorCir.Port
				log.WithFields(log.Fields{
					"Tor IP":         TorCir.IPAddr,
					"Source Address": conn.LocalAddr().String(),
					"Socks5 Address": addr,
					"Circuit Load":   TorCir.Load,
				}).Info("Tcp load balancer")

				conn2, err := net.DialTimeout("tcp", addr, 10*time.Second)
				if err != nil {
					log.Error("error dialing remote addr ", err)
					log.Info("Deleting dirty circuit")
					for i, v := range torList {
						if v.IPAddr == TorCir.IPAddr {
							torList = RemoveTorList(torList, i)
						}
					}
					log.Info("Request new circuit")
					newTor, err := initTor(1)
					if err != nil {
						log.Error(err)
						return
					}
					torList = append(torList, newTor...)
					return
				}
				go io.Copy(conn2, conn)
				io.Copy(conn, conn2)
				conn2.Close()
				conn.Close()
			}()
		}

	}()

	go func() {
		for {
			log.Info("Start tor Health check")
			HealthCheck(torList)
			time.Sleep(2 * time.Hour)
		}
	}()

	http.ListenAndServe(":"+*RestAPIPort, router)
}

func (i *TorStruct) TorStructLoad() *TorStruct {
	i.Load++
	return i
}
