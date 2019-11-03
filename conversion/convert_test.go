package conversion

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/diff"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		target  string
		want    string
		wantErr bool
	}{
		{
			"empty v1alpha1",
			`{"apiVersion":"things.kubecon.io/v1alpha1","kind":"Device","spec":{}}`,
			"things.kubecon.io/v1",
			`{"apiVersion":"things.kubecon.io/v1","kind":"Device","spec":{}}`,
			false,
		},
		{
			"v1alpha1 with spec",
			`{"apiVersion":"things.kubecon.io/v1alpha1","kind":"Device","spec":{"inputs":[{"name":"Switch","type":"Float","value":"1"}]}}`,
			"things.kubecon.io/v1",
			`{"apiVersion":"things.kubecon.io/v1","kind":"Device","spec":{"inputs":[{"name":"Switch","float":"1"}]}}`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, _, err := codecs.UniversalDeserializer().Decode([]byte(tt.in), nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			expected, _, err := codecs.UniversalDeserializer().Decode([]byte(tt.want), nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			got, err := convert(obj, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("got unpexpect converted object: %s", diff.ObjectDiff(expected, got)) //diff.StringDiff(marshal(expected), marshal(got)))
			}

			roundtripped, err := convert(got, obj.GetObjectKind().GroupVersionKind().GroupVersion().String())
			if err != nil {
				t.Errorf("unexpected roundtrip error: %v", err)
				return
			}

			if !reflect.DeepEqual(obj, roundtripped) {
				t.Errorf("got unpexpect roundtripped object: %s", diff.StringDiff(marshal(obj), marshal(roundtripped)))
			}
		})
	}
}
