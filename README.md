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

Edit `controllers/manifests.yaml`, and set the image to the just published to the repo.

```
kubectl apply -f controllers/manifests.yaml
```

TODO: restart the controller. Use deployment instead of replica set?

## 9. Introduce v1 of our CRDs

Open 

## 10. Add a conversion webhook


### Intall the conversion webhook service

First, start the conversion webhook:

```
kubectl apply -f conversion/manifests.yaml
```

### Bring your own cluster:

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
  jpbetz/deepsea-controllers:latest -metrics-addr :8082 -simulator-addr http://localhost:8085

# start the simulator
docker run -it -p8085:8085 jpbetz/deepsea-simulator:latest

# open http://localhost:8085 in a browser
```
