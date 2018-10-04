# k8s-udp-multicaster

Kubernetes does not natively implement UDP multicasting using `Services`.

This container image can be used to proxy UDP traffic and multicast it to all of the `Endpoints` of a given `Service`. This is useful if you don't want to be broadcasting the UDP packets to every Pod in your cluster

for example, this pod will listen for UDP packets on the port `9782` and forward them to all the `Endpoints` of `udp-listeners`, using the port named `udp`
```
apiVersion: apps/v1
kind: Pod
metadata: 
  name: udp-multicaster
spec: 
  serviceAccountName: udp-multicaster
  containers: 
  - image: "k8s-udp-multicaster:latest"
      name: udp-multicaster
      ports: 
      - containerPort: 9782
      env: 
      - name: LISTEN_PORT
      value: "9782"
      - name: SERVICE_NAME
      value: "udp-listeners"
      - name: SERVICE_PORT
      value: "udp"
```

see `examples/packet-tool-example.yaml` for a more complete example, including the minimum required `ClusterRole`
