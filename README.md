KoT
---

Local
-----

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

Steps
----

1. Create a cluster

2. Development environment

Cloud shell

$ git clone https://github.com/jpbetz/KoT.git
$ cd KoT

<Click on editor button on top right of cloud shell screen to get a basic editor>

$ export CLUSTER=<name of cluster>
$ export PROJECT=<name of project>
$ gcloud container clusters get-credentials "${CLUSTER}" --zone us-central1-a --project "${PROJECT}"

2. Run the simulator

$ kubectl apply -f manifests/simulator
$ kubectl get ing 
NAME                HOSTS   ADDRESS          PORTS   AGE
simulator-ingress   *       35.244.159.176   80      156m

<navigate to the simulator IP in a browser>

3. Run the controller

$ kubectl apply -f manifests/controllers

4. Create our main CRDs and resources

$ kubectl apply -f manifests/v1beta1-crds
$ kubectl apply -f examples/command-module
$ kubectl apply -f examples/crew-module
$ kubectl apply -f examples/research-module

<check the simulator UI, all the modules should appear>

5. Add reconciliation to the controller

<edit controllers/devicereconciler.go>
<go to the ReconcilePressure function>
<find the TODO to add calculate how to change pressure>
<explain how the pump rules work>

$ make build-controllers
$ TODO: make push-controllers, where to?

<edit manifests/controller/controllers-replicaset.yaml, update the container image>

<update the running pods. TODO: use deployment instead of replica set?>

6. Introduce v1 of our CRDs

TODO

7. Add a conversion webhook

TODO

8. Demonstrate how we can access our reousrces via the v1 API