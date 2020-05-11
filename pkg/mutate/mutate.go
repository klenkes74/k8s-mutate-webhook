// Package mutate deals with AdmissionReview requests and responses, it takes in the request body and returns a readily converted JSON []byte that can be
// returned from a http Handler w/o needing to further convert or modify it, it also makes testing Mutate() kind of easy w/o need for a fake http server, etc.
package mutate

import (
	"encoding/json"
	"fmt"
	"log"

	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Mutate mutates
func Mutate(body []byte) ([]byte, error) {

	log.Printf("recv: %s\n", string(body))

	// unmarshal request into AdmissionReview struct
	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var err error
	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admReview.Request
	resp := v1beta1.AdmissionResponse{}

	if ar != nil {

		// get the Pod object and unmarshal it into its struct, if we cannot, we might as well stop here
		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}
		// set response options
		resp.Allowed = true
		resp.UID = ar.UID
		pT := v1beta1.PatchTypeJSONPatch
		resp.PatchType = &pT // it's annoying that this needs to be a pointer as you cannot give a pointer to a constant?

		var auditmessage string
		var patch []byte

		if pod.ObjectMeta.GetAnnotations() == nil || len(pod.ObjectMeta.GetAnnotations()) == 0 {
			patch = []byte(`[{"op":"add","path":"/metadata/annotations","value":{"cluster-autoscaler.kubernetes.io/safe-to-evict":"true"}}]`)

			auditmessage = "Created first annotation: cluster-autoscaler.kubernetes.io/save-to-evict"
		} else {
			patch = []byte(`[{"op":"add","path":"/metadata/annotations/cluster-autoscaler.kubernetes.io~1safe-to-evict","value":"true"}]`)

			auditmessage = "Added annotation 'cluster-autoscaler.kubernetes.io/save-to-evict' to the annotations"
		}

		resp.AuditAnnotations = map[string]string{
			"add-eviction-helper": auditmessage,
		}

		resp.Patch = []byte(patch)

		// Success, of course ;)
		resp.Result = &metav1.Status{
			Status: "Success",
		}

		admReview.Response = &resp
		// back into JSON so we can return the finished AdmissionReview w/ Response directly
		// w/o needing to convert things in the http handler
		responseBody, err = json.Marshal(admReview)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("resp: %s\n", string(responseBody))
	return responseBody, nil
}
