KoT
---

Steps
----

## 0. Getting ready

The step by step directions are also available at <shortlink>

Please do not work ahead! Much of the value in this Tutorial will the explanations we
provide as we go.

Following along as we go trough this tutorial is encouraged, but not
required.

Warning: While we hope everyone is able to follow along. We've got xyz people in a large
room and the wifi might not cooperate. We're doing a couple things to try to make this work:

- We will try to minimize the bandwidth required
- We've provided two ways to catch up if you fall behind:
  - Each step of the process we show here has a corresponding git commit, the command to catch up will
    be displayed on the bottom on each slide
  - All docker images we create during the session are available in docker.io. If you're unable to get
    one built and published, just use the one we provide.
  - If you are unable get the UI working, or just don't want to enable it, we have a command line
    operation that provides similar output

## 1. Create a cluster

### Bring your own cluster:

Feel free to use your own kubernetes cluster.

Any 1.16+ cluster should be fine.
Local clusters via minikube or `kubernetes/local-up-cluster.sh` should be fine.

(We have a 1.15 example, but it does not highlight the functionality nearly as well)

### Cloud:

Create a gke cluster. Must be kubernetes 1.16! Need to click on "rapid" to find the 1.16 option.

Enable gcloud build: https://pantheon.corp.google.com/apis/library/cloudbuild.googleapis.com (we'll be using free tier)

TODO: add detailed cluster creation steps

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
$ kubectl apply -f simulator/manifests.yaml
```

It provides a web UI. It can be accessed either via kube-proxy
or by adding an ingress.

### Accessing Simulator UI - kube-proxy

```
export POD=$(kubectl -n deepsea get pods -o name -l app=deepsea-simulator)
kubectl -n deepsea port-forward "${POD}" 8080:8085
```

Navigate to http://localhost:8080 in a browser.

### Accessing Simulator UI - ingress

```
$ kubectl get ing 
NAME                HOSTS   ADDRESS          PORTS   AGE
simulator-ingress   *       35.244.159.176   80      156m
```

When the ingress is ready, we can navigate to the simulator-ingress IP in a browser.
For now, just continue on getting things set up.

## 4. Deploy the tutorial controllers into the cluster

TODO: Include how types are generated before this?

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

(at this step, the simulator should be varying the pressure, but nothing else should be happening).

If the UI is inaccessible, or you just would rather use the terminal, instead, do:

```
watch -n 0.1 kubectl get -n examples devices
```

This provides outputs for each field using the additionalPrintColumns in the CRD.

## 6. Add reconciliation to the controller

Edit controllers/devicereconciler.go

Find the `ReconcilePressure` function. Find the TODO to add calculate how to change pressure. Implement it.

## 7. Build and publish a docker image of the controller

### Bring your own cluster

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

Edit `controllers/manifests.yaml`, and set the image to the just published to the repo.

```
kubectl apply -f controllers/manifests.yaml
```

## 9. Introduce v1 of our CRDs

In `manifests/kubernetes-1.16-crds/devices-crd.yaml`
note that a `v1` version is defined but disabled.
We cannot enable it until we have a way to converting
between v1alpha1 and v1.

## 10. Implement a conversion webhook

TODO: What code edits should we add here?

### Install the conversion webhook service

#### Bring your own cluster:

```
$ make build-conversion
$ make push-conversion
```

#### Cloud

```
$ gcloud builds submit --config cloudbuild/conversion.yaml .
````

### Register the conversion webhook

Next, register the webhook. Update `manifests/kubernetes-1.15-crds/devices-crd.yaml` and add the following:

```
spec:
  
  conversion:
    strategy: Webhook
    webhook:
      clientConfig:
        caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUV6RENDQXJRQ0NRQ2ZuNjhmeUlCdWpqQU5CZ2txaGtpRzl3MEJBUXNGQURBb01TWXdKQVlEVlFRRERCMWoKYjI1MlpYSnphVzl1TFhkbFltaHZiMnN1ZEdocGJtZHpMbk4yWXpBZUZ3MHhPVEV4TURNeE1qUXpNamRhRncweQpNREV4TURJeE1qUXpNamRhTUNneEpqQWtCZ05WQkFNTUhXTnZiblpsY25OcGIyNHRkMlZpYUc5dmF5NTBhR2x1ClozTXVjM1pqTUlJQ0lqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FnOEFNSUlDQ2dLQ0FnRUFuemJpSTNyMmJ2eDQKMnIzZ3BqS3lnNXVwVnFiSGsrUjVkRDlpbVdUZDNtVkl4bFJHbGhsSG1ZQk5MVWlFR0hvZjhmZnliVGo3enU3KwpMdVZNOExhZktaTmI2Z0Y4ZmIyYjdVRDNTNmtkRHdxL2F5NklLd1I2cVBKcng0MFNWUzlQN3V4Y3Z2SUloNGRKCkVBNmdBTjVHUnBaL08zMkJRMXZvd01wNTdwNG1Fb2JYTWF0MDVHOUlQWE4xQk05dFdURis2bzd0bHYvSEpxVm8KWnBhYmlJWVo3RWFaV24zZWpBTTVVWWJ5RTJ0LzJTSUJweVg2b0pSUmVaMmN6aEhBNkJ2dUN6N3NRZzJpMHE4eQpPcmloL1RTUlN0U05CcWZEN0pqdDFjMkQyRXNvMGFOZHNGK2ZzaGJTTjhCN2hocEszeENaOHpWUXV4MExzUlFECjJuanl2alhqQVdTRmJkZC8xaXk4TG1NeFpvVlNDOExIWW9JUXdqNTUyNERMYURjblVsKzBNUTVISWRCTU42UVoKck5JUHBnTVJVaFAvTS8vYTBEM0VLU24vQk10ZXY0YnRTQmhGUFJzeWdUbWsxc0lrUVB2WHlhK1hLdUVwTnZaQgpjdjg0RVI1WXVmREZVQ2ZDTFg5YTBNQXZjeXpDb0FNa3R0M2svbk1XR3NYZURwN3lVM25DNWxlNkw2ZDBIQTJyCmRUSnRwRWZwZk5BZXlQcmE1dUlNZzRFL0tNWCt1K24yQXBKUTJxTFhDbktkZjUzZ0E4K1NUWFBaK0lEa1ZSMWoKbWd4UVFaamNXd1lwdkxjdEpSWFByNjhWemtyTkhqR0FoU1BQdHFnVGtxVldWYzl6RkpaOFZYM1V1dnlzVjVKeQpGQ283R21Fd1J1djhCejNuendMUU42UlgzZEFZalpjQ0F3RUFBVEFOQmdrcWhraUc5dzBCQVFzRkFBT0NBZ0VBCkNBdWUyUnM3TnpCeksySktoMXA2N3poV2dUTVN5T1JNRTc0YkduaklmbGlGL2lramx6dDFLYlAyU0hGYVkxTTYKZGZabnBTL3ZYbGU5Z0tKeTJTd0NVMVR3blZtd3hpVnd3S1RvUWZONVFGNGxNWU5wWlRDNWFhYk5PbzNEcmdqMQpBRVgxVHVjT3liTEx1Y3lVTllqZ3RjNmJuTytwYXFYemtBeVEyZ0pCMjFsZVQxT3RuemFVQUlidmdLZVdnMmF3CjZHRUNSSkVTT09PQ0w5OVNESm5jVGxteHQ4VytKU2prQ0g1Q2tmNjZ6NHhXN212OEVkQ095MEdmamgvVkhtTkoKbHZESDkwVFBxcHN4QW5xSDhOY2Z1U21UcUViQW9RbW9wZ2hBbGlhVlh5amVZMmNpenNyZ2NsK0ZqRnViTTdJcgpQeXZvVmsycGpuR1M1cit2MERjRzcwaXMwWG5KUnJVWTBlSk1WSFVFZzkrOHNTODE4TXZVS1FPcUNicHIvYmFuCnJMRWltMXVZS1RvN3ZZdTFRT1RzUmxBY0s0Z05pZ3c4dFlpODZLbDZSODVNc0pFbGY1azlxeTZ3djk5aFRmbmoKSGtueE1sclk5WXJQaG9rRnRSUkJJbGZDa2VKYW5EaW5XRXdxWkIxdWQ4K3duUlJ4dWhNQ2lVVlNBUCs2eXdjaApFZVBReklqcElOWDdXaVBWWWg5UXo1SjhXTnNHTC83SDVBY2Q2TDJ1OVRYUXFmSGp0bzRFRXJXV0ZvMURUVFUxCksrSGhqOGNjbCtWUnUvU1NGSFZvMVdzTXBmaVU0THdYYzBianFZWXY1VWdCWE40K2lnVDRWUUtxK0t0NGx6cFIKZVpMaFhEcXdJd2x5VndRbzdoT00xZUs0NE04d3g2MFhKaVY1L0hnN2pTYz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
        service:
          namespace: things
          name: conversion-webhook
          path: /convert/v1/devices
      conversionReviewVersions:
      - v1beta1
```

```
kubectl apply -f manifests/kubernetes-1.15-crds/devices-crd.yaml
```

### Enable v1 of our CRDs

Edit `manifests/kubernetes-1.16-crds/devices-crd.yaml`, set served to true for `v1`.

```
$ kubectl apply -f manifests/kubernetes-1.16-crds
```

## 11. How to access resources at different versions

To get a resource at `v1`:

```
$ kubectl get -n examples devices research-pressure -o yaml
...
status:
  outputs:
  - float: 10253m
    name: pressure
...

```

To get the resource at `v1alpha1`, ask for it explicitly:

```
$ kubectl get -n examples devices.v1alpha1.things.kubecon.io research-pressure -o yaml
...
status:
  outputs:
  - name: pressure
    type: Float
    value: 10229m
```

## 12. Bonus

Install the admission webhook

Local Docker Testing
--------------------

```sh

# start kubernetes locally
sudo PATH=$PATH hack/local-up-cluster.sh

# start the controllers
docker run -it --network=host -e KUBECONFIG=/var/run/kubernetes/admin.kubeconfig \
  --mount type=bind,source=/var/run/kubernetes/admin.kubeconfig,target=/var/run/kubernetes/admin.kubeconfig \
  jpbetz/deepsea-controllers:latest -metrics-addr :8082

# start the simulator
docker run -it -p8085:8085 jpbetz/deepsea-simulator:latest

# open http://localhost:8085 in a browser
```
