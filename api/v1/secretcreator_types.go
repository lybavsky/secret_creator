/*
Copyright 2021.

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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretCreatorSpec defines the desired state of SecretCreator
type SecretCreatorSpec struct {
	//SecretName is name of secret will be created
	SecretName string `json:"secretName"`
	//Secret is a Secret object
	Secret v1.Secret `json:"secret"`

	//Not working: exclude namespaces with creating
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`
}

// SecretCreatorStatus defines the observed state of SecretCreator
type SecretCreatorStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SecretCreator is the Schema for the secretcreators API
type SecretCreator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretCreatorSpec   `json:"spec,omitempty"`
	Status SecretCreatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretCreatorList contains a list of SecretCreator
type SecretCreatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretCreator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretCreator{}, &SecretCreatorList{})
}
