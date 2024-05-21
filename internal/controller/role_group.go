package controller

import (
	stackv1alpha1 "github.com/zncdatadev/argo-workflow-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func (r *ArgoWorkFlowReconciler) getRoleGroupLabels(config *stackv1alpha1.RoleConfigSpec) map[string]string {
	additionalLabels := make(map[string]string)
	if configLabels := config.MatchLabels; configLabels != nil {
		for k, v := range config.MatchLabels {
			additionalLabels[k] = v
		}
	}
	return additionalLabels
}

func (r *ArgoWorkFlowReconciler) mergeLabels(instanceLabels map[string]string,
	roleGroup *stackv1alpha1.RoleConfigSpec) Map {
	var mergeLabels *Map
	mergeLabels.MapMerge(instanceLabels, true)
	mergeLabels.MapMerge(r.getRoleGroupLabels(roleGroup), true)
	return *mergeLabels
}

func (r *ArgoWorkFlowReconciler) getServiceInfo(instanceSvc *stackv1alpha1.ServiceSpec,
	roleGroup *stackv1alpha1.RoleConfigSpec) (int32, corev1.ServiceType, map[string]string) {
	var targetSvc = instanceSvc
	if roleGroup != nil && roleGroup.Service != nil {
		targetSvc = roleGroup.Service
	}
	return targetSvc.Port, targetSvc.Type, targetSvc.Annotations
}

func (r *ArgoWorkFlowReconciler) getDeploymentInfo(instance *stackv1alpha1.ArgoWorkFlow,
	roleGroup *stackv1alpha1.RoleConfigSpec) (*stackv1alpha1.ImageSpec, *corev1.PodSecurityContext, int32,
	[]corev1.EnvVar, *corev1.ResourceRequirements) {
	var (
		image           = instance.Spec.RoleConfig.Image
		securityContext = instance.Spec.RoleConfig.SecurityContext
		replicas        = instance.Spec.RoleConfig.Replicas
		envVars         []corev1.EnvVar
		resources       = instance.Spec.RoleConfig.Config.Resources
	)
	if roleGroup != nil {
		if rgImage := roleGroup.Image; rgImage != nil {
			image = rgImage
		}
		if rgSecurityContext := roleGroup.SecurityContext; rgSecurityContext != nil {
			securityContext = rgSecurityContext
		}
		if rgReplicas := roleGroup.Replicas; rgReplicas != 0 {
			replicas = rgReplicas
		}
		if rgResources := roleGroup.Config.Resources; rgResources != nil {
			resources = rgResources
		}
	}
	envVars = append(envVars, corev1.EnvVar{
		Name: "ARGO_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	})
	envVars = append(envVars, corev1.EnvVar{
		Name: "LEADER_ELECTION_IDENTITY",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})
	return image, securityContext, replicas, envVars, resources
}
