
**Table of Contents** 
- [Demo Environment](#demo-environment)
- [Prerequisite](#prerequisite)
  - [setup a `microk8s` cluster](#setup-a-microk8s-cluster)
  - [setup environment](#setup-environment)
  - [setup repository & workflows](#setup-repository--workflows)
- [Example 1: Round Robin Loadbalancing with gRPC's built-in loadbalancing policy](#example-1-round-robin-loadbalancing-with-grpcs-built-in-loadbalancing-policy)

# Area of problem
gRPC is over HTTP2 by default(HTTP3 is coming), as we generally know that HTTP2 uses the concept of stream, which means we have a single connection that handles multiple requests by multiplexing/demultiplexing, this reduces the overhead of establishing a new connection every time we make an HTTP request. Since we have only one connection, in order to handle a bunch of requests, to balance the load we have to look at the L7 but our applications are running on k8s environment, and the default service kind deployment uses L4 to balance the loads

# Solution
figure out a way to deal with L7 load-balancing

# Demo Environment
- VM Instance on GCP
- Machine type `e2-standard-4`
- CPU Architecture `x86/64`
- OS Ubuntu `20.04-focal-v20240125`

# Prerequisite
## setup a `microk8s` cluster
- install `microk8s` on your device
- run `microk8s enable metallb`
- run `microk8s enable ingress`
- run `microk8s kubectl apply -f ingress-lb-service.yml`
- setup private registry
    - `vim /var/snap/microk8s/current/args/containerd-template.toml`
    - add this section at the end of file
    ```
    # private repository
    [plugins."io.containerd.grpc.v1.cri".registry.configs."<REGISTRY_DOMAIN_NAME>".auth]
    username = "_json_key_base64"
    password = "<BASE64_CREDENTAIL_CONTENT>"
    email = "bitkuber@example.com"
    ```
## setup environment
This demo uses GCR as a private registry, so, we have to export some variables for the deployment files
- `export ARTIFACT_REGION=<REGION>`
- `export ARTIFACT_PROJECT_ID=<PROJECT_ID>`

## setup repository & workflows
- Go to Settings -> Environments then added these variables 
    - `ARTIFACT_PROJECT_ID` put your projectID into this fied as a `Secrets`
    - `GCR_CREDENTIAL` put your service account with base64 encoded as a `Secrets`
    - `ARTIFACT_REGION` put your target region as a `Variables`

**_NOTE_**: `GCR_CREDENTIAL` permission to push an artifact to the registry is needed

# Example 1: Retrieve IPs via DNS and use Round Robin Loadbalancing with gRPC's built-in load-balancing policy
- deploy ingress by applying `microk8s kubectl apply -f ./dnsclient/ingress.yml -n <app_namespace>`
- run `envsubst < ./dnsclient/deployment.yml | microk8s kubectl -n <app_namespace> apply -f -`
- run `envsubst < ./server/deployment.yml | microk8s kubectl -n <app_namespace> apply -f -`
- mapping the ingress hosts into the known hosts file at `/etc/hosts` for instance, 
    ```
    127.0.0.1 localhost
    10.123.1.1 go-grpc-dnsclient.demo

    # 10.123.1.1 is my private network interface that maps with my LB service
    ```

# Example 2: Retrieve IPs via DNS and use Round Robin Loadbalancing with statically configured Envoy proxy
- deploy ingress by applying `microk8s kubectl apply -f ./defaultclient/ingress.yml -n <app_namespace>`
- run `envsubst < ./defaultclient/client-with-envoy-deployment.yml | microk8s kubectl -n <app_namespace> apply -f -`
- run `envsubst < ./server/deployment.yml | microk8s kubectl -n <app_namespace> apply -f -`
- mapping the ingress hosts into the known hosts file at `/etc/hosts` for instance, 
    ```
    127.0.0.1 localhost
    10.123.1.1 go-grpc-defaultclient-with-envoy.demo

    # 10.123.1.1 is my private network interface that maps with my LB service
    ```