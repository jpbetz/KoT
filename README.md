KoT
---

Steps
----

1. Create a cluster

TODO

2. Development environment

DYI:

<run a local kubernetes cluster, or use an existing dev cluster you have available>
<Must be kubernetes 1.16!>

$ git clone https://github.com/jpbetz/KoT.git
$ cd KoT

<Open the project in your editor of choice>

Cloud:

<log into a gce account>
<create a gke cluster for kubernetes 1.16>
<open cloud shell from the GKE cluster view so that the terminal is configured to connect to the cluster>

$ git clone https://github.com/jpbetz/KoT.git
$ cd KoT

<Click on editor button on top right of cloud shell screen to get a basic editor>

2. Run the simulator

$ kubectl apply -f manifests/simulator
$ kubectl get ing 
NAME                HOSTS   ADDRESS          PORTS   AGE
simulator-ingress   *       35.244.159.176   80      156m

<navigate to the simulator-ingress IP in a browser, it can take awhile for the ingress to get ready...>

3. Run the controller

$ kubectl apply -f manifests/controllers

4. Create our main CRDs and resources

$ kubectl apply -f manifests/v1beta1-crds
$ kubectl apply -f examples/command-module
$ kubectl apply -f examples/crew-module
$ kubectl apply -f examples/research-module

<check the simulator UI, it should be more interesting now>

5. Add reconciliation to the controller

<edit controllers/devicereconciler.go>
<go to the ReconcilePressure function>
<find the TODO to add calculate how to change pressure>
<explain how the pump rules work>

6. Build and publish a docker image of the controller

DYI:

$ make build-controllers
$ make push-controllers


Cloud:

<enable gcr?>
<enable cloud build>

$gcloud builds submit --config cloudbuild/controllers.yaml .

...
DONE
------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
ID                                    CREATE_TIME                DURATION  SOURCE                                                                                    IMAGES                                                     STATUS
3589a20b-2bfd-4139-8c09-081339b88677  2019-11-06T01:02:42+00:00  4M55S     gs://jpbetz-gke-dev_cloudbuild/source/1573002074.94-2e8786cc06794d6ab74a553cba67b298.tgz  gcr.io/jpbetz-gke-dev/things-conversion-webhook (+1 more)  SUCCESS


7. Run the updated controller

<edit manifests/controller/controllers-replicaset.yaml, update the container image>

<update the running pods. TODO: use deployment instead of replica set?>

8. Introduce v1 of our CRDs

TODO

9. Add a conversion webhook

TODO

10. Demonstrate how we can access our reousrces via the v1 API

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
