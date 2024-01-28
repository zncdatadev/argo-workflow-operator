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
)

func (r *ArgoWorkFlowReconciler) makeService(instance *stackv1alpha1.ArgoWorkFlow, schema *runtime.Scheme) *corev1.Service {
	labels := instance.GetLabels()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      labels,
			Annotations: instance.Spec.Service.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     instance.Spec.Service.Port,
					Name:     "http",
					Protocol: "TCP",
				},
			},
			Selector: labels,
			Type:     instance.Spec.Service.Type,
		},
	}
	err := ctrl.SetControllerReference(instance, svc, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for service")
		return nil
	}
	return svc
}

func (r *ArgoWorkFlowReconciler) reconcileService(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	obj := r.makeService(instance, r.Scheme)
	if obj == nil {
		return nil
	}

	if err := CreateOrUpdate(ctx, r.Client, obj); err != nil {
		r.Log.Error(err, "Failed to create or update service")
		return err
	}
	return nil
}

func (r *ArgoWorkFlowReconciler) makeDeployment(instance *stackv1alpha1.ArgoWorkFlow, schema *runtime.Scheme) *appsv1.Deployment {
	labels := instance.GetLabels()
	envVars := []corev1.EnvVar{}

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
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.GetNameWithSuffix("-controller"),
					SecurityContext:    instance.Spec.SecurityContext,
					Containers: []corev1.Container{
						{
							Name:            instance.Name,
							Image:           instance.Spec.Image.Repository + ":" + instance.Spec.Image.Tag,
							ImagePullPolicy: instance.Spec.Image.PullPolicy,
							Args: []string{
								"--configmap",
								instance.GetNameWithSuffix("-controller"),
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
							Resources: *instance.Spec.Resources,
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

	CreateScheduler(instance, dep)

	err := ctrl.SetControllerReference(instance, dep, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for deployment")
		return nil
	}
	return dep
}

func (r *ArgoWorkFlowReconciler) reconcileDeployment(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	obj := r.makeDeployment(instance, r.Scheme)
	if obj == nil {
		return nil
	}
	if err := CreateOrUpdate(ctx, r.Client, obj); err != nil {
		logger.Error(err, "Failed to create or update deployment")
		return err
	}

	return nil
}

func (r *ArgoWorkFlowReconciler) makeServiceAccount(instance *stackv1alpha1.ArgoWorkFlow, schema *runtime.Scheme) *corev1.ServiceAccount {
	labels := instance.GetLabels()
	satoken := true
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix("-controller"),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		AutomountServiceAccountToken: &satoken,
	}
	err := ctrl.SetControllerReference(instance, sa, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for ServiceAccount")
		return nil
	}
	return sa
}

func (r *ArgoWorkFlowReconciler) reconcileServiceAccount(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	obj := r.makeServiceAccount(instance, r.Scheme)
	if obj == nil {
		return nil
	}

	if err := CreateOrUpdate(ctx, r.Client, obj); err != nil {
		r.Log.Error(err, "Failed to create or update ServiceAccount")
		return err
	}
	return nil
}

func (r *ArgoWorkFlowReconciler) makeClusterRoleBinding(instance *stackv1alpha1.ArgoWorkFlow, schema *runtime.Scheme) *rbacv1.ClusterRoleBinding {
	labels := instance.GetLabels()
	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      instance.GetNameWithSuffix("-controller"),
		Namespace: instance.Namespace,
	}
	subjects := []rbacv1.Subject{subject}
	crbd := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix("-controller"),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "manager-role",
		},
		Subjects: subjects,
	}
	err := ctrl.SetControllerReference(instance, crbd, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for ClusterRoleBinding")
		return nil
	}
	return crbd
}

func (r *ArgoWorkFlowReconciler) reconcileClusterRoleBinding(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	obj := r.makeClusterRoleBinding(instance, r.Scheme)
	if obj == nil {
		return nil
	}

	if err := CreateOrUpdate(ctx, r.Client, obj); err != nil {
		r.Log.Error(err, "Failed to create or update ServiceAccount")
		return err
	}
	return nil
}

func (r *ArgoWorkFlowReconciler) makeConfigMap(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow, schema *runtime.Scheme) *corev1.ConfigMap {
	labels := instance.GetLabels()

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix("-controller"),
			Namespace: instance.Namespace,
			Labels:    labels,
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

	err := ctrl.SetControllerReference(instance, configMap, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for configmap")
		return nil
	}
	return configMap
}

func (r *ArgoWorkFlowReconciler) reconcileConfigMap(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow) error {
	obj := r.makeConfigMap(ctx, instance, r.Scheme)
	if obj == nil {
		return nil
	}

	if err := CreateOrUpdate(ctx, r.Client, obj); err != nil {
		r.Log.Error(err, "Failed to create or update service")
		return err
	}
	return nil
}
