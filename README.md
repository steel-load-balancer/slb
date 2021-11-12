# slb

A load-balancer forged in the fires of [Sheffield](https://en.wikipedia.org/wiki/Steel_City)

![](Mount_Doom.gif)

The **Steel Load Balancer**


   ![](steel.gif)

## How it works!

The **slb** automates a number of seperate tasks to provide the end user a simple to use load-balancing experience. It provides a simple API endpoint and will automatically create a load-balancer endpoint (IP address), which a user can easily add backend servers too.

**Requirements**

The requirements are very lightweight!!

- Linux host (running a 3.x+ kernel)
- Layer 2 network (for ARP)
- Layer 3 network (for BGP)

**Architecture**

The Machine that is running slb will create an API endpoint on port 10001 that is used to create load-balancers, this API endpoint is used to create/update and delete load-balancer instances. 

- In Layer2 `slb` should be started with an IPAM range to create loadbalancers with
- In layer3 `slb` will use the Equinix Metal API to select an IP address.

This new load-balancer is then advertised to the outside world through either network mechanism. We can now use the API to add backend `server:port` to this load-balancer instance at this point we will use IPVS to create a NAT based round-robin load-balancer instance. 

### Using SLB

#### When running on a layer 2 network

The below example will create an `slb` server that listens on port 10001 and will generate load-balancers in the ipamRange specified.
```
./slb server \
      --arp \
      --adapter eth0 \
      --ipamRange 192.168.0.85-192.168.0.90
```

**Example output**

```
INFO[0000] Starting the Steel Load-Balancer
INFO[0000] API Server will be exposed on [0.0.0.0:10001]
INFO[0000] Enabled IPv4 Forwarding in the kernel
INFO[0000] Enabled Connection tracking in the kernel
```

#### When using Equinix Metal (Layer 3)

The Auth token is required in order for API queries:

```
export PACKET_AUTH_TOKEN=""
```

Also required are the project and facility IDs so that Equinix Metal Elastic IPs are created in the correct facility/project.

```
./slb server \
      --adapter lo \
      --bgp \
      --equinixMetal \
      --project xxxx \
      --facility xxxx
```

**Example output**

```
INFO[0000] Starting the Steel Load-Balancer
INFO[0000] API Server will be exposed on [0.0.0.0:10001]
INFO[0000] Enabled IPv4 Forwarding in the kernel
INFO[0000] Enabled IPv4 Virtual Server connection tracking in the kernel
INFO[0002] Querying BGP settings for [am6-c3.small.x86-01]
INFO[0002] Add a peer configuration for:169.254.255.1    Topic=Peer
INFO[0002] Add a peer configuration for:169.254.255.2    Topic=Peer
INFO[0007] Peer Up                                       Key=169.254.255.1 State=BGP_FSM_OPENCONFIRM Topic=Peer
2021/11/12 13:40:02 conf:<local_as:65000 neighbor_address:"169.254.255.1" peer_as:65530 > state:<local_as:65000 neighbor_address:"169.254.255.1" peer_as:65530 session_state:ESTABLISHED router_id:"147.75.86.18" > transport:<local_address:"10.12.39.1" local_port:34481 remote_port:179 >
INFO[0008] Peer Up                                       Key=169.254.255.2 State=BGP_FSM_OPENCONFIRM Topic=Peer
2021/11/12 13:40:03 conf:<local_as:65000 neighbor_address:"169.254.255.2" peer_as:65530 > state:<local_as:65000 neighbor_address:"169.254.255.2" peer_as:65530 session_state:ESTABLISHED router_id:"147.75.86.19" > transport:<local_address:"10.12.39.1" local_port:45091 remote_port:179 >
```

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
