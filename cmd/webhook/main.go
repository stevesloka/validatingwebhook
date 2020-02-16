package main

import (
	"fmt"
	"log"
	"net/http"

	"io/ioutil"

	"encoding/json"

	ingressroutev1 "github.com/projectcontour/contour/apis/contour/v1beta1"
	projectcontour "github.com/projectcontour/contour/apis/projectcontour/v1"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got request...")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	ar := v1beta1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)

		payload, err := json.Marshal(&v1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		})
		if err != nil {
			fmt.Println(err)
		}
		w.Write(payload)
	}

	admitResponse := &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: true,
		},
	}

	if ar.Request.Kind.Kind == "Pod" {
		pod := v1.Pod{}
		json.Unmarshal(ar.Request.Object.Raw, &pod)
		for _, container := range pod.Spec.Containers {
			if container.Name == "steve" {
				fmt.Println("BLOCK container from running...")
				admitResponse.Response.Allowed = false
				admitResponse.Response.Result = &metav1.Status{
					Message: "Ah ah ahhhh, you can't do this! [STEVE]",
				}
				break
			} else {
				fmt.Println("Container is a-ok!")
			}
		}
	} else if ar.Request.Kind.Kind == "IngressRoute" {
		ir := ingressroutev1.IngressRoute{}
		json.Unmarshal(ar.Request.Object.Raw, &ir)

		if ir.GetNamespace() != "root" && ir.GetNamespace() != "" {
			fmt.Println("BLOCK ingressroute, can't run outside root namespace...")
			admitResponse.Response.Allowed = false
			admitResponse.Response.Result = &metav1.Status{
				Message: "Attempt to deploy to non-root namespace",
			}
		}
	} else if ar.Request.Kind.Kind == "HTTPProxy" {
		ir := projectcontour.HTTPProxy{}
		json.Unmarshal(ar.Request.Object.Raw, &ir)

		if ir.GetNamespace() != "root" && ir.GetNamespace() != "" {
			fmt.Println("BLOCK httpproxy, can't run outside root namespace...")
			admitResponse.Response.Allowed = false
			admitResponse.Response.Result = &metav1.Status{
				Message: "Attempt to deploy to non-root namespace",
			}
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	payload, err := json.Marshal(admitResponse)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(payload)
}

func main() {
	fmt.Println("webhook starting up...")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServeTLS(":9443", "/certs/server.crt", "/certs/server.key", nil))
}
