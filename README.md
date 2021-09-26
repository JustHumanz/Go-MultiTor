## Go Multi Tor 
Multitor with golang


### Purpose
this tools was inspired by [multitor](https://github.com/trimstray/multitor),but multitor was not very scalable with container and have issuse with high memory usage 


### Workflow

```
                            ┌───────────────────────┐
                            │                       │
                            │      Mux (Rest api)   │
            ┌───────────────┤                       ├───────────────────────────┐
            |               │       Controller      │                           │
            |               │                       │                           │
            |               └────────────┬──────────┘                           │
            |                            │                                      │
            |                            │                                      │
            |                            │                                      │
            |                            │                                      │
            |                            │                                      │
            |                            │                                 xxxxx│xxxxxxxxxx
            |                            │                                 x ┌──┴─────┐   x
            |                            │                                 x │        │   x
            |                            │                                 x │ Tor    │   x
            |                            │                ┌────────────────x─┤        │   x
            │                            │                │                x └────────┘   x
            │                            │                │                x  Tor cluster x
            │            ┌───────────────┴────────────────┤                x ┌────────┐   x
            │            |                                │                x │        │   x
 Internet   ├────────────┤  Socks5 tcp load balancer      ├────────────────x─┤ Tor    │   x
                         │                                │                x │        │   x
                         └────────────────────────────────┤                x └────────┘   x
                                                          │                x              x
                                                          │                x ┌────────┐   x
                                                          │                x │        │   x
                                                          └────────────────x─┤ Tor    │   x
                                                                           x │        │   x
                                                                           x └────────┘   x
                                                                           xxxxxxxxxxxxxxxx
                                                                            

```
Create a bunch of tor circuit > load balancing theme with round robin > serve as ~~http~~ socks5 proxy  
then i create controller,the role of controller is create/delete tor circuit

### Requirement
- tor
- go1.16.5 linux/amd64 or latest
- ~~privoxy~~

### Use
```
docker run -it -d -p 1412:1412 -p 2525:2525 --name=go-multitor justhumanz/go-multitor
while true;do curl -x socks5://localhost:1412 https://ifconfig.me;printf "\n";done
```

### Rest API
soon

### Features
- [x] Scalable tor circuit
- [x] Rest API 
- [x] Http proxy
- [x] Scoks5 load balancer
- [x] Rest API use authentication
- [ ] Multiple tor circuit nodes
- [x] Tor health check
