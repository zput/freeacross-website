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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebsiteDeploymentSpec defines the desired state of WebsiteDeployment
type WebsiteDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DeployTemplateConfig  `json:"deploy_template_config"`
	ServiceTemplateConfig `json:"service_template_config"`
}

type DeployTemplateConfig struct {
	Containers []Container `json:"containers"`

	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []corev1.Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
}

type ServiceTemplateConfig struct {
	Type string `json:"type"` // NodePort
	// +optional
	Ports []Port `json:"ports,omitempty"`
}

type Container struct {
	// Docker image name.
	// +kubebuilder:validation:Required
	Image string `json:"image,omitempty"`
	// 容器名称
	Name string `json:"name"`

	// +optional
	Ports []Port `json:"ports,omitempty"`

	// Pod volumes to mount into the container's filesystem.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,9,rep,name=volumeMounts"`

	// List of environment variables to set in the Sandbox.
	// Cannot be updated.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// NetworkProtocol define network protocols supported of the Sandbox.
type NetworkProtocol string

const (
	// ProtocolHTTP is the HTTP protocol.
	ProtocolHTTP NetworkProtocol = "HTTP"
	// ProtocolTCP is the TCP protocol.
	ProtocolTCP NetworkProtocol = "TCP"

	ProtocolHELIOS NetworkProtocol = "HELIOS"
)

// SandboxNetWork contain information on sandbox network.
type Port struct {
	// +optional
	Port int32 `json:"port,omitempty"`
	// +optional
	Protocol NetworkProtocol `json:"protocol,omitempty"`
}

// DeploymentTemplate Deployment 模板
//type DeploymentTemplate struct {
//	metav1.ObjectMeta `json:"metadata,omitempty"`
//
//	// Deployment spec template
//	// +kubebuilder:validation:Required
//	Spec appsv1.DeploymentSpec `json:"spec,omitempty"`
//}
//
//// ServiceTemplate Service 模板
//type ServiceTemplate struct {
//	metav1.ObjectMeta `json:"metadata,omitempty"`
//
//	// +kubebuilder:validation:Required
//	Spec corev1.ServiceSpec `json:"spec,omitempty"`
//}

// WebsiteDeploymentStatus defines the observed state of WebsiteDeployment
type WebsiteDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WebsiteDeployment is the Schema for the websitedeployments API
type WebsiteDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebsiteDeploymentSpec   `json:"spec,omitempty"`
	Status WebsiteDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WebsiteDeploymentList contains a list of WebsiteDeployment
type WebsiteDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebsiteDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebsiteDeployment{}, &WebsiteDeploymentList{})
}
