/*
Copyright 2022.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebappSpec defines the desired state of Webapp
type WebappSpec struct {
	//+kubebuilder:validation:Required
	AzureTenantId string `json:"azureTenantId"`
	//+kubebuilder:validation:Required
	AzureSpnId string `json:"azureSpnId"`
	//+kubebuilder: validation:Required
	AzureSpnSecret string `json:"azureSpnSecret"`
	//+kubebuilder:validation:Required
	StorageName string `json:"storageName"`
	// +kubebuilder: default:=$web
	ContainerName string `json:"containerName"`
	// +kubebuilder:default:=index.html
	FileNameToCheck string `json:"filenameToCheck"`
	// +kubebuilder: default:=version
	BlobTagKey string `json:"blobTagKey"`
	//+kubebuilder:validation: Required
	VersionToDeploy string `json:"versionToDeploy"`
	//+kubebuilder:validation:Required
	PackageStorageName string `json:"packageStorageName"`
	// +kubebuilder:default:=packages
	PackageContainerName string `json:"packageContainerName"`
}

// WebappStatus defines the observed state of Webapp
type WebappStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status          string `json:"status"`
	DeployedVersion string `json:"deployed-version"`
	Error           string `json:"error"`
	LastUpdate      string `json:"last-update"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="The status of the last sync"
//+kubebuilder:printcolumn:name="Current Deployed Version",type="string",JSONPath=".status.deployed-version",description="The version currently deployed"
//+kubebuilder:printcolumn:name="Desired Version",type="string",JSONPath=".spec.webappversion",description="The desired version"
//+kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.error",description="Potential error during reconciliation"
// Webapp is the Schema for the webapps API
type Webapp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebappSpec   `json:"spec,omitempty"`
	Status WebappStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// WebappList contains a list of Webapp
type WebappList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webapp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Webapp{}, &WebappList{})
}
