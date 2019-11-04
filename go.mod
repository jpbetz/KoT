module github.com/jpbetz/KoT

go 1.12

require (
	github.com/coreos/bbolt v1.3.1-coreos.6 // indirect
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/validate v0.19.2 // indirect
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jpbetz/KoT/apis v0.0.0-20191103203357-766b1d8a78d6
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/spf13/pflag v1.0.3
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7 // indirect
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0 // indirect
	k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20191029223907-9f431a56fdbc
	k8s.io/component-base v0.0.0-20190918200425-ed2f0867c778 // indirect
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
	sigs.k8s.io/structured-merge-diff v0.0.0-20190817042607-6149e4549fca
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48

replace github.com/jpbetz/KoT/controllers => ./controllers
