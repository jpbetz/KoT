KoT
---

```sh

# start kubernetes locally
sudo PATH=$PATH hack/local-up-cluster.sh

# start the controllers
docker run -it --network=host -e KUBECONFIG=/var/run/kubernetes/admin.kubeconfig \
  --mount type=bind,source=/var/run/kubernetes/admin.kubeconfig,target=/var/run/kubernetes/admin.kubeconfig \
  jpbetz/deepsea-controllers:latest -metrics-addr :8082 -simulator-addr http://localhost:8085

# start the simulator
docker run -it -p8085:8085 jpbetz/deepsea-simulator:latest

# open http://localhost:8085 in a browser
```