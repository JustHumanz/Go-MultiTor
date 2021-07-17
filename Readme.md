## Multi Tor Go
Multitor with golang


### Purpose
this tools was inspired by [multitor](https://github.com/trimstray/multitor),but multitor was not very scalable with container and have issuse with high memory usage 


### Workflow

```

                                                           ┌───────────────────────┐
                                                           │                       │
                                                           │      Mux (Rest api)   │
                                    ┌──────────────────────┤                       ├───────────────────────────┐
                                    │                      │       Controller      │                           │
                                    │                      │                       │                           │
                                    │                      └────────────┬──────────┘                           │
                                    │                                   │                                      │
                                    │                                   │                                      │
                                    │                                   │                                      │
                                    │                                   │                                      │
                                    │                                   │                                      │
                                    │                                   │                                 xxxxx│xxxxxxxxxx
                                    │                                   │                                 x ┌──┴─────┐   x
                                    │                                   │                                 x │        │   x
                                    │                                   │                                 x │ Tor    │   x
                             ┌──────┴──────┐                            │                ┌────────────────x─┤        │   x
                             │             │                            │                │                x └────────┘   x
                             │  Privoxy    │                            │                │                x  Tor cluster x
                             │             │            ┌───────────────┴────────────────┤                x ┌────────┐   x
                             │             │            │                                │                x │        │   x
Internet────────────────────►│  HTTP proxy ├────────────┤        Socks5 tcp balancer     ├────────────────x─┤ Tor    │   x
                             │             │            │                                │                x │        │   x
                             │             │            └────────────────────────────────┤                x └────────┘   x
                             │             │                                             │                x              x
                             │             │                                             │                x ┌────────┐   x
                             │             │                                             │                x │        │   x
                             └─────────────┘                                             └────────────────x─┤ Tor    │   x
                                                                                                          x │        │   x
                                                                                                          x └────────┘   x
                                                                                                          xxxxxxxxxxxxxxxx
                                                                                                           

```
Create a bunch of tor circuit > load balancing theme with round robin > serve as http proxy with privoxy  
then i create controller,role of controller is create/delete tor circuit, maybe in next update i will health check for every single tor circuit

### Requirement
- tor
- go1.16.5 linux/amd64
- privoxy

### Features
- [x] Scalable tor circuit
- [x] Rest API 
- [x] Http proxy
- [x] Scoks5 load balancer
- [x] Rest API use authentication
- [ ] Multiple tor circuit nodes
- [ ] Tor health check