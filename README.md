# rpcscout

`rpcscout` is a program to test gRPC and REST connections across nodes and
service meshes in a Kubernetes environment.  It is both the client and the
server to itself and support N-to-N connections. For example it can support:

## Deployments

### Simple client server

```
[rpcscout] --[grpc/REST]--> [rpcscout]
```

Execute the following to deploy this setup:

```
kubectl create -f https://raw.githubusercontent.com/lpabon/rpcscout/main/deploy/simple.yml
```

### Example of a more complicated setup

```
[rpcscout x10] ===[grpc/REST]==> [rpcscout x3] ===[grpc/REST]==> [rpcscout x1]
Emulate clients                   Emulate API servers             Emulate a DB
```

Execute the following to deploy this setup:

```
kubectl create -f https://raw.githubusercontent.com/lpabon/rpcscout/main/deploy/app-demo.yml
```


