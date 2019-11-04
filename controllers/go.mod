module github.com/jpbetz/KoT/controllers

go 1.12

replace (
	github.com/jpbetz/KoT => ../
	k8s.io/api => k8s.io/api v0.0.0-20191016110246-af539daaa43a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115701-31ade1b30762
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016110837-54936ba21026
)

require (
	github.com/go-logr/logr v0.1.0
	github.com/jpbetz/KoT v0.0.0-00010101000000-000000000000
	github.com/jpbetz/KoT/apis v0.0.0-20191103203357-766b1d8a78d6
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
)
