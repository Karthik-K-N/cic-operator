# cic-operator



## Development

1. Create cluster using [kind](https://kind.sigs.k8s.io/)
```
$ kind create cluster
```

Note: If kind not installed locally try installing it, One of the way is
```
$ GO111MODULE="on" go get sigs.k8s.io/kind@v0.9.0
```

2. Clone the repository
```
$ git clone https://github.com/Karthik-K-N/cic-operator.git
```

3. Export necessary environmental variables used for authentication

```
export OS_USERNAME=FILLIN
export OS_PASSWORD=FILLIN
export OS_IDENTITY_API_VERSION=3
export OS_AUTH_URL=https://FILLIN:5000/v3/
export OS_CACERT=/Users/karthikkn/Downloads/icic.crt
export OS_REGION_NAME=RegionOne
export OS_PROJECT_DOMAIN_NAME=Default
export OS_PROJECT_NAME=ibm-default
export OS_TENANT_NAME=$OS_PROJECT_NAME
export OS_USER_DOMAIN_NAME=Default
```

4Run the following commands to install crd and start the controller
```
$ make manifests
$ make install
$ make run
```

4. Create a VM resource

```
$ kubectl create -f config/samples/cloud_v1_vm.yaml
```

5. Get the VM resource

```
$ kubectl get vm
 NAME        STATUS
 vm-sample   ACTIVE
```