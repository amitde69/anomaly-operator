/*
Copyright 2023.

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

package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DetectorSpec defines the desired state of Detector
type DetectorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Detector. Edit detector_types.go to remove/update
	PromUrl      string      `json:"prom_url,required"`
	Image      string      `json:"image,required"`
	IntervalMins string      `json:"interval_mins,required"`
	Queries      []QuerySpec `json:"queries,required"`
	PodSpec		v1.PodTemplateSpec	`json:"pod_spec,omitempty"`
}

type QuerySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Detector. Edit detector_types.go to remove/update
	Name        string `json:"name,required"`
	Query       string `json:"query,required"`
	Train_Window string `json:"train_window,required"`
	Detection_Window_Hours int64 `json:"detection_window_hours,omitempty"`
	Flexibility string `json:"flexibility,omitempty"`
	Buffer_Pct   int64  `json:"buffer_pct,omitempty"`
	Resolution  int64  `json:"resolution,omitempty"`
}

// DetectorStatus defines the observed state of Detector
type DetectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	IsCreated bool `json:"iscreated,omitempty"`
	Deployment string `json:"deployment,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Detector is the Schema for the detectors API
type Detector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DetectorSpec   `json:"spec,omitempty"`
	Status DetectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DetectorList contains a list of Detector
type DetectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Detector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Detector{}, &DetectorList{})
}

type Config struct {
    IntervalMins string      `yaml:"interval_mins"`
    PromURL      string   `yaml:"prom_url"`
    Queries      []Query `yaml:"queries"`
}

type Query struct {
    Buffer_Pct   *int     `yaml:"buffer_pct,omitempty"`
    Flexibility float64  `yaml:"flexibility,omitempty"`
    Name        string   `yaml:"name"`
    Query       string   `yaml:"query"`
    Resolution  int      `yaml:"resolution,omitempty"`
	Detection_Window_Hours int64 `yaml:"detection_window_hours,omitempty"`
    Train_Window string   `yaml:"train_window"`
}