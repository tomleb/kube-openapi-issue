package main

import (
	"reflect"
	"testing"

	"k8s.io/kube-openapi/pkg/util/proto"
	oapitesting "k8s.io/kube-openapi/pkg/util/proto/testing"
)

func TestRefs(t *testing.T) {
	documents := []string{
		// The fake values are taken as is from
		// https://github.com/kubernetes/kube-openapi/blob/835d969ad83aca3ef62637b255332df649856429/pkg/util/proto/testdata/openapi_v3_0_0/apps/v1.json
		"apps-v1-fake",

		// The real values are taken from k3s clusters with the following command:
		// kubectl get --raw /openapi/v3/apis/apps/v1 | jq

		// The real clusters don't consider references as *proto.Ref. The difference between
		// the two is the following:

		// Fake:
		//
		//     "ownerReferences": {
		//       "description": "List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller.",
		//       "type": "array",
		//       "items": {
		//         "default": {},
		//         "$ref": "#/components/schemas/io.k8s.apimachinery.pkg.apis.meta.v1.OwnerReference"
		//       },
		//       ...
		//     },
		// Real:
		//     "ownerReferences": {
		//       "description": "List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller.",
		//       "type": "array",
		//       "items": {
		//         "default": {},
		//         "allOf": [
		//           {
		//             "$ref": "#/components/schemas/io.k8s.apimachinery.pkg.apis.meta.v1.OwnerReference"
		//           }
		//         ]
		//       },
		//       ...
		//     },

		// v1.28.6 cluster
		"apps-v1-1.28.6",
		// v1.30.0 cluster
		"apps-v1-1.30.0",
	}
	for _, document := range documents {
		t.Run(document, func(t *testing.T) {
			fake := oapitesting.FakeV3{
				Path: "testdata",
			}
			doc, err := fake.OpenAPIV3Schema(document)
			if err != nil {
				t.Fatal(err)
			}

			models, err := proto.NewOpenAPIV3Data(doc)
			if err != nil {
				t.Fatal(err)
			}

			schema := models.LookupModel("io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta")
			metadata := schema.(*proto.Kind)
			ownerRefsSchema := metadata.Fields["ownerReferences"]
			ownerRefs := ownerRefsSchema.(*proto.Array)
			switch ownerRefs.SubType.(type) {
			case *proto.Ref:
			default:
				t.Fatal("Expected ownerReference to be a proto.Ref, but it was a", reflect.TypeOf(ownerRefs.SubType))
			}
		})
	}
}
