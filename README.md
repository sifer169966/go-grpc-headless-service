
**Table of Contents** 
- [Demo Environment](#demo-environment)
- [Prerequisite](#prerequisite)
  - [setup a `microk8s` cluster](#setup-a-microk8s-cluster)
  - [setup environment](#setup-environment)
- [Example 1: Round Robin Loadbalancing with gRPC's built-in loadbalancing policy](#example-1-round-robin-loadbalancing-with-grpcs-built-in-loadbalancing-policy)



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
- run `microk8s apply -f ingress-lb-service.yml`
- setup private registry
    - vim `/var/snap/microk8s/current/args/containerd-template.toml`
    - add this section at the end of file
    ```
    # private repository
    [plugins.cri.registry]
        [plugins.cri.registry.mirrors]
            [plugins.cri.registry.mirrors."<REGISTRY_DOMAIN>"]
                endpoint = ["https://<REGISTRY_DOMAIN>"]
    [plugins.cri.registry.auths]
        [plugins.cri.registry.auths."https://<REGISTRY_DOMAIN>"]
            username = "_json_key_base64"
            password = "<BASE64_CREDENTAIL_VALUE>"
            email = "example@gmail.com"
    ```
## setup environment
This demo is publish the application artifact on GCR, so, we have to export some variables for the deployment files
- `export ARTIFACT_REGION=<REGION>`
- `export ARTIFACT_PROJECT_ID=<PROJECT_ID>`


# Example 1: Round Robin Loadbalancing with gRPC's built-in loadbalancing policy
- deploy ingress by applying `microk8s apply -f ./client/ingress.yml -n <app_namespace>`
- run `envsubst < ./client/deployment.yml | microk8s -n <app_namespace> apply -f -`
- run `envsubst < ./server/deployment.yml | microk8s -n <app_namespace> apply -f -`

