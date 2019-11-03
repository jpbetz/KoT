module github.com/jpbetz/KoT

go 1.12

require (
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/spf13/pflag v1.0.3
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/component-base v0.0.0-20191016111319-039242c015a9
	k8s.io/klog v1.0.0
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
