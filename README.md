KoT
---

Steps
----

## 1. Create a cluster

### Bring your own cluster:

Must be kubernetes 1.16!

### Cloud:

Log into a GCE account.

Create a gke cluster. Must be kubernetes 1.16!

## 2. Development environment

### Bring your own cluster:

```
$ git clone https://github.com/jpbetz/KoT.git
$ cd KoT
```

Open the project in your editor of choice.

### Cloud:

Open cloud shell from the GKE cluster view so that the terminal is configured to connect to the cluster. (cloud only)

Checkout the tutorial project:

Click on editor button on top right of cloud shell screen to get a basic editor.

## 3. Deploy the tutorial simulator into the cluster

```
$ kubectl apply -f manifests/simulator
$ kubectl get ing 
NAME                HOSTS   ADDRESS          PORTS   AGE
simulator-ingress   *       35.244.159.176   80      156m
```

Navigate to the simulator-ingress IP in a browser, it can take awhile for the ingress to get ready.

## 4. Deploy the tutorial controller into the cluster

```
$ kubectl apply -f manifests/controllers
```

## 5. Install CRDs and resources

```
$ kubectl apply -f manifests/kubernetes-1.16-crds
$ kubectl apply -f examples/command-module
$ kubectl apply -f examples/crew-module
$ kubectl apply -f examples/research-module
```

Check the simulator UI, it should be more interesting now.

## 6. Add reconciliation to the controller

Edit controllers/devicereconciler.go

Find the `ReconcilePressure` function. Find the TODO to add calculate how to change pressure. Implement it.

## 7. Build and publish a docker image of the controller

### DYI

```
$ make build-controllers
$ make push-controllers
```

### Cloud

TODO: Enable GCR?
TODO: Enable cloud build?

```
$ gcloud builds submit --config cloudbuild/controllers.yaml .


DONE
------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
ID                                    CREATE_TIME                DURATION  SOURCE                                                                                    IMAGES                                                     STATUS
3589a20b-2bfd-4139-8c09-081339b88677  2019-11-06T01:02:42+00:00  4M55S     gs://jpbetz-gke-dev_cloudbuild/source/1573002074.94-2e8786cc06794d6ab74a553cba67b298.tgz  gcr.io/jpbetz-gke-dev/things-conversion-webhook (+1 more)  SUCCESS
```

## 8. Run the updated controller

Edit `manifests/controller/controllers-replicaset.yaml`, replace the container image with the newly published one.

TODO: restart the controller. Use deployment instead of replica set?

## 9. Introduce v1 of our CRDs

## 10. Add a conversion webhook

### DYI

```
$ make build-conversion
$ make push-conversion
```

### Cloud

```
$ gcloud builds submit --config cloudbuild/conversion.yaml .
````

Enable v1 of our CRDs.

Edit `manifests/kubernetes-1.16-crds/devices-crd.yaml`, set served to true for `v1`.

```
$ kubectl apply -f manifests/kubernetes-1.16-crds
```

## 11. How to access resources at different versions

To get a resource at `v1`:

```
$ kubectl get -n examples devices research-pressure -o yaml
```

To get the resource at `v1alpha1`, ask for it explicitly:

```
$ kubectl get -n examples devices.v1alpha1.things.kubecon.io research-pressure -o yaml
```

## TODO

What else?

Local Docker Testing
--------------------

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
