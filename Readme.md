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

### Features
- [x] Scalable tor circuit
- [x] Rest API 
- [x] Http proxy
- [x] Scoks5 load balancer
- [ ] Rest API use authentication
- [ ] Multiple tor circuit nodes