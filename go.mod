module github.com/jpbetz/KoT

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/spf13/pflag v1.0.3
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	k8s.io/api v0.0.0-20191105190043-25240d7d6d90
	k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery v0.0.0-20191105185716-00d39968b57e
	k8s.io/apiserver v0.0.0-20191105191200-ab2ea16b1965
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20191029223907-9f431a56fdbc
	k8s.io/component-base v0.0.0-20191105110211-1d7e08732f45
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.3.1-0.20191011155739-8b53f2bca0e7

)
