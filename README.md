# slb
A Load-balancer made from steel

![](steel.gif)

If I figure out how to make it work..

## How it works!

The slb LB PoC uses BGP/IPVS and Equinix Metal EIPs to create a load-balancer service.

**Requirements**

A Machine will need to be created inside of Equinix Metal to host, slb (you'll also need the various IDs for the Facility/project and an AUTH token.

**Architecture**

The EM Machine that is running slb will create an API endpoint on port 10001 that is used to create load-balancers. When we create a new loadBalancer with the API, slb will use the EM API to request a public IP address and it will then use BGP to advertise that address to the ToR switches. At this point our loadbalancer address will be public and we should be able to hit it with a `ping`. We can now use the API to add backend `server:port` to this load-balancer instance at this point we will use IPVS to create a NAT based round-robin load-balancer instance. 

### Set the environment

```
export FACILITY_ID=""
export PACKET_AUTH_TOKEN=""
export PROJECT_ID=""
```

### Start slb

```
$ sudo -E ./vippy 
INFO[0000] Starting Equinix Metal - Vippy LoadBalancer  
INFO[0000] Creating Equinix Metal Client                
INFO[0000] Creating BGP                                 
INFO[0002] Querying BGP settings for [am6-c3.small.x86-01] 
INFO[0003] Add a peer configuration for:169.254.255.1    Topic=Peer
INFO[0003] Add a peer configuration for:169.254.255.2    Topic=Peer
INFO[0003] API Server is now listening on :10001        
INFO[0010] Peer Up                                       Key=169.254.255.1 State=BGP_FSM_OPENCONFIRM Topic=Peer
INFO[0010] Peer Up                                       Key=169.254.255.2 State=BGP_FSM_OPENCONFIRM Topic=Peer
INFO[0010] conf:<local_as:65000 neighbor_address:"169.254.255.1" peer_as:65530 > state:<local_as:65000 neighbor_address:"169.254.255.1" peer_as:65530 session_state:ESTABLISHED router_id:"147.75.86.18" > transport:<local_address:"10.12.39.1" local_port:49627 remote_port:179 >  
INFO[0010] conf:<local_as:65000 neighbor_address:"169.254.255.2" peer_as:65530 > state:<local_as:65000 neighbor_address:"169.254.255.2" peer_as:65530 session_state:ESTABLISHED router_id:"147.75.86.19" > transport:<local_address:"10.12.39.1" local_port:48247 remote_port:179 >  
```

We can see above that Vippy has started, and has connected to the ToR switches in the facility ready to advertise our first load-balancer!

### Create a load-balancer

`curl -X POST http://<host>:10001/loadbalancer -H 'Content-Type: application/json' -d '{"name":"test loadbalancer", "port": 6443}'`

We can view this new loadbalancer through the API or on the cli where `vippy` is running!

```
curl -s http://<host>:10001/loadbalancers | jq
[
  {
    "uuid": "087fcb0f-8e07-4123-a32f-e8e83a281023",
    "name": "test loadbalancer",
    "eip": "147.75.84.41",
    "port": 6443,
    "backends": null
  }
]
```

### Add a backend

We can use the API to add a backend service to our load-balancer instance!

`curl -X POST http://<host>:10001/loadbalancer/087fcb0f-8e07-4123-a32f-e8e83a281023/backend -d '{"ip":"10.80.96.25", "port": 6443}'`

```
curl -s http://<node>:10001/loadbalancers | jq
[
  {
    "uuid": "087fcb0f-8e07-4123-a32f-e8e83a281023",
    "name": "test loadbalancer",
    "eip": "145.40.96.223",
    "port": 6443,
    "backends": [
      {
        "uuid": "779344ce-72ca-4aec-a1c0-d9c5c4d72419",
        "ip": "10.80.96.25",
        "port": 6443
      }
    ]
  }
]
```

#### View load-balancer in ipvs

On our vippy node we can inspect in more detail our load-balancer setup

```
ipvsadm -ln
IP Virtual Server version 1.2.1 (size=4096)
Prot LocalAddress:Port Scheduler Flags
  -> RemoteAddress:Port           Forward Weight ActiveConn InActConn
TCP  145.40.96.223:6443 rr
  -> 10.80.96.25:6443             Masq    1      0          0
  ```

### Delete a load-balancer

The following API call will delete the EIP, stop bgp and remove the IPVS service.
```
curl -X DELETE http://<node>:10001/loadbalancer/087fcb0f-8e07-4123-a32f-e8e83a281023
```
