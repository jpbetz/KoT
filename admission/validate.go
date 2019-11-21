/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package admission

import (
	"fmt"
	"io/ioutil"
	"net/http"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/endpoints/handlers/negotiation"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"

	"github.com/jpbetz/KoT/apis/deepsea/install"
	deepseev1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	informers "github.com/jpbetz/KoT/generated/informers/externalversions"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func init() {
	utilruntime.Must(admissionv1beta1.AddToScheme(scheme))
	install.Install(scheme)
}

func ModuleValidation(informers informers.SharedInformerFactory) func(http.ResponseWriter, *http.Request) {
	devicesInformer := informers.Things().V1alpha1().Devices().Informer()
	devicesLister := informers.Things().V1alpha1().Devices().Lister()

	return func(w http.ResponseWriter, req *http.Request) {
		if !devicesInformer.HasSynced() {
			responsewriters.InternalError(w, req, fmt.Errorf("informers not ready"))
			return
		}

		// read body
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			responsewriters.InternalError(w, req, fmt.Errorf("failed to read body: %v", err))
			return
		}

		// decode body as admission review
		reviewGVK := admissionv1beta1.SchemeGroupVersion.WithKind("AdmissionReview")
		obj, gvk, err := codecs.UniversalDeserializer().Decode(body, &reviewGVK, &admissionv1beta1.AdmissionReview{})
		if err != nil {
			responsewriters.InternalError(w, req, fmt.Errorf("failed to decode body: %v", err))
			return
		}
		review, ok := obj.(*admissionv1beta1.AdmissionReview)
		if !ok {
			responsewriters.InternalError(w, req, fmt.Errorf("unexpected GroupVersionKind: %s", gvk))
			return
		}
		if review.Request == nil {
			responsewriters.InternalError(w, req, fmt.Errorf("unexpected nil request"))
			return
		}
		review.Response = &admissionv1beta1.AdmissionResponse{
			UID: review.Request.UID,
		}

		// decode object
		if review.Request.Object.Object == nil {
			var err error
			review.Request.Object.Object, _, err = codecs.UniversalDeserializer().Decode(review.Request.Object.Raw, nil, nil)
			if err != nil {
				review.Response.Result = &metav1.Status{
					Message: err.Error(),
					Status:  metav1.StatusFailure,
				}
				responsewriters.WriteObjectNegotiated(codecs, negotiation.DefaultEndpointRestrictions, gvk.GroupVersion(), w, req, http.StatusOK, review)

				return
			}
		}

		switch module := review.Request.Object.Object.(type) {
		case *deepseev1alpha1.Module:
			var errs []error
			_ = devicesLister
			_ = module

			// put your admission logic here

			err = utilerrors.NewAggregate(errs)
			if err != nil {
				review.Response.Result = &metav1.Status{
					Message: err.Error(),
					Status:  metav1.StatusFailure,
				}
			} else {
				review.Response.Allowed = true
			}

		default:
			review.Response.Result = &metav1.Status{
				Message: fmt.Sprintf("unexpected type %T", review.Request.Object.Object),
				Status:  metav1.StatusFailure,
			}
		}

		responsewriters.WriteObjectNegotiated(codecs, negotiation.DefaultEndpointRestrictions, gvk.GroupVersion(), w, req, http.StatusOK, review)
	}
}
