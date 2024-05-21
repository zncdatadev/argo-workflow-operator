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
	"github.com/zncdatadev/operator-go/pkg/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ArgoWorkFlow is the Schema for the argoworkflows API
type ArgoWorkFlow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArgoWorkFlowSpec `json:"spec,omitempty"`
	Status status.Status    `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ArgoWorkFlowList contains a list of ArgoWorkFlow
type ArgoWorkFlowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArgoWorkFlow `json:"items"`
}

// ArgoWorkFlowSpec defines the desired state of ArgoWorkFlow
type ArgoWorkFlowSpec struct {

	// +kubebuilder:validation:Optional
	RoleConfig *RoleConfigSpec `json:"roleConfig"`

	// +kubebuilder:validation:Optional
	RoleGroups map[string]*RoleConfigSpec `json:"roleGroups"`

	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels"`
}

type RoleConfigSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas int32 `json:"replicas"`

	// +kubebuilder:validation:Optional
	Image *ImageSpec `json:"image"`

	// +kubebuilder:validation:Required
	Service *ServiceSpec `json:"service"`

	// +kubebuilder:validation:Optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext"`

	// +kubebuilder:validation:Optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector"`

	// +kubebuilder:validation:Optional
	Tolerations *corev1.Toleration `json:"tolerations"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	ServiceAccount *bool `json:"serviceAccount"`

	// +kubebuilder:validation:Optional
	Config *ConfigSpec `json:"config"`
}

type ConfigSpec struct {
	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources"`
}

type ImageSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="bitnami/argo-workflow-controller"
	Repository string `json:"repository,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="3.5.0"
	Tag string `json:"tag,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

type ServiceSpec struct {
	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +kubebuilder:validation:enum=ClusterIP;NodePort;LoadBalancer;ExternalName
	// +kubebuilder:default=ClusterIP
	Type corev1.ServiceType `json:"type,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=18080
	Port int32 `json:"port"`
}

// ArgoWorkFlowStatus defines the observed state of ArgoWorkFlow
type ArgoWorkFlowStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"condition,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ArgoWorkFlow{}, &ArgoWorkFlowList{})
}

func (a *ArgoWorkFlow) GetNameWithSuffix(suffix string) string {
	// return ArgoWorkFlow.GetName() + rand.String(5) + suffix
	return a.GetName() + suffix
}
func (a *ArgoWorkFlow) InitStatusConditions() {
	a.Status.InitStatus(a)
	a.Status.InitStatusConditions()
}
