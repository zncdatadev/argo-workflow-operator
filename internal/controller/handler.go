package controller

import (
	"context"
	stackv1alpha1 "github.com/zncdata-labs/argo-workflow-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ArgoWorkFlowReconciler) reconcileService(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	err := r.createOrUpdateResource(ctx, instance, r.extractServiceForRoleGroup)
	if err != nil {
		return err
	}
	return nil
}

func (r *ArgoWorkFlowReconciler) extractServiceForRoleGroup(instance *stackv1alpha1.ArgoWorkFlow, _ context.Context,
	groupName string, roleGroup *stackv1alpha1.RoleConfigSpec, scheme *runtime.Scheme) (client.Object, error) {
	mergeLabels := r.mergeLabels(instance.GetLabels(), roleGroup)
	port, svcType, annotations := r.getServiceInfo(instance.Spec.RoleConfig.Service, roleGroup)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.GetNameWithSuffix(groupName),
			Namespace:   instance.Namespace,
			Labels:      mergeLabels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     port,
					Name:     "http",
					Protocol: "TCP",
				},
			},
			Selector: mergeLabels,
			Type:     svcType,
		},
	}
	err := ctrl.SetControllerReference(instance, svc, scheme)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for service")
		return nil, err
	}
	return svc, nil
}

// reconcile deployment
func (r *ArgoWorkFlowReconciler) reconcileDeployment(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	err := r.createOrUpdateResource(ctx, instance, r.extractDeploymentForRoleGroup)
	if err != nil {
		return err
	}
	return nil
}

// extract deployment for role group
func (r *ArgoWorkFlowReconciler) extractDeploymentForRoleGroup(instance *stackv1alpha1.ArgoWorkFlow, _ context.Context,
	groupName string, roleGroup *stackv1alpha1.RoleConfigSpec, scheme *runtime.Scheme) (client.Object, error) {
	mergeLabels := r.mergeLabels(instance.GetLabels(), roleGroup)
	image, securityContext, replicas, envVars, resources := r.getDeploymentInfo(instance, roleGroup)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix(groupName),
			Namespace: instance.Namespace,
			Labels:    mergeLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: mergeLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: mergeLabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.GetNameWithSuffix(groupName),
					SecurityContext:    securityContext,
					Containers: []corev1.Container{
						{
							Name:            instance.GetNameWithSuffix(groupName),
							Image:           image.Repository + ":" + image.Tag,
							ImagePullPolicy: image.PullPolicy,
							Args: []string{
								"--configmap",
								instance.GetNameWithSuffix(groupName),
								"--executor-image",
								"docker.io/bitnami/argo-workflow-exec:3.5.0-debian-11-r0",
								"--executor-image-pull-policy",
								"IfNotPresent",
								"--loglevel",
								"info",
								"--gloglevel",
								"0",
								"--workflow-workers",
								"32",
							},
							Env:       envVars,
							Resources: *resources,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 18080,
									Name:          "http",
									Protocol:      "TCP",
								},
							},
						},
					},
				},
			},
		},
	}
	CreateScheduler(instance, dep, roleGroup)

	err := ctrl.SetControllerReference(instance, dep, scheme)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for instance")
		return nil, err
	}
	return dep, nil
}

// reconcile service account
func (r *ArgoWorkFlowReconciler) reconcileServiceAccount(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	err := r.createOrUpdateResource(ctx, instance, r.extractServiceAccountForRoleGroup)
	if err != nil {
		return err
	}
	return nil
}

// extract service account for role group
func (r *ArgoWorkFlowReconciler) extractServiceAccountForRoleGroup(instance *stackv1alpha1.ArgoWorkFlow, _ context.Context,
	groupName string, roleGroup *stackv1alpha1.RoleConfigSpec, scheme *runtime.Scheme) (client.Object, error) {
	mergeLabels := r.mergeLabels(instance.GetLabels(), roleGroup)
	satoken := true
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix(groupName),
			Namespace: instance.Namespace,
			Labels:    mergeLabels,
		},
		AutomountServiceAccountToken: &satoken,
	}
	err := ctrl.SetControllerReference(instance, sa, scheme)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for ServiceAccount")
		return nil, err
	}
	return sa, err
}

// reconcile cluster role binding
func (r *ArgoWorkFlowReconciler) reconcileClusterRoleBinding(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	err := r.createOrUpdateResource(ctx, instance, r.extractClusterRoleBindingForRoleGroup)
	if err != nil {
		return err
	}
	return nil
}

// extract cluster role binding for role group
func (r *ArgoWorkFlowReconciler) extractClusterRoleBindingForRoleGroup(instance *stackv1alpha1.ArgoWorkFlow, _ context.Context,
	groupName string, roleGroup *stackv1alpha1.RoleConfigSpec, scheme *runtime.Scheme) (client.Object, error) {
	mergeLabels := r.mergeLabels(instance.GetLabels(), roleGroup)
	crbd := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix(groupName),
			Namespace: instance.Namespace,
			Labels:    mergeLabels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "manager-role",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      instance.GetNameWithSuffix(groupName),
				Namespace: instance.Namespace,
			},
		},
	}
	err := ctrl.SetControllerReference(instance, crbd, scheme)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for ClusterRoleBinding")
		return nil, err
	}
	return crbd, nil
}

// reconcile config map
func (r *ArgoWorkFlowReconciler) reconcileConfigMap(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	err := r.createOrUpdateResource(ctx, instance, r.extractConfigMapForRoleGroup)
	if err != nil {
		return err
	}
	return nil
}

// extract config map for role group
func (r *ArgoWorkFlowReconciler) extractConfigMapForRoleGroup(instance *stackv1alpha1.ArgoWorkFlow, _ context.Context,
	groupName string, roleGroup *stackv1alpha1.RoleConfigSpec, scheme *runtime.Scheme) (client.Object, error) {
	mergeLabels := r.mergeLabels(instance.GetLabels(), roleGroup)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix(groupName),
			Namespace: instance.Namespace,
			Labels:    mergeLabels,
		},
		Data: map[string]string{
			"config": `
				parallelism:
				namespaceParallelism:
				executor:
				  resources:
					limits: {}
					requests: {}
			`,
		},
	}
	err := ctrl.SetControllerReference(instance, configMap, scheme)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for configmap")
		return nil, err
	}
	return configMap, nil
}
