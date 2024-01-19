package controller

import (
	"context"
	stackv1alpha1 "github.com/zncdata-labs/argo-workflow-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ArgoWorkFlowReconciler) extractResources(instance *stackv1alpha1.ArgoWorkFlow, ctx context.Context,
	roleGroupExtractor func(*stackv1alpha1.ArgoWorkFlow, context.Context, string, *stackv1alpha1.RoleConfigSpec,
		*runtime.Scheme) (client.Object, error)) ([]client.Object, error) {
	var resources []client.Object
	if instance.Spec.RoleGroups != nil {
		for roleGroupName, roleGroup := range instance.Spec.RoleGroups {
			rsc, err := roleGroupExtractor(instance, ctx, roleGroupName, roleGroup, r.Scheme)
			if err != nil {
				return nil, err
			}
			resources = append(resources, rsc)
		}
	}
	return resources, nil
}
func (r *ArgoWorkFlowReconciler) createOrUpdateResource(ctx context.Context, instance *stackv1alpha1.ArgoWorkFlow,
	roleGroupExtractor func(*stackv1alpha1.ArgoWorkFlow, context.Context, string, *stackv1alpha1.RoleConfigSpec,
		*runtime.Scheme) (client.Object, error)) error {
	resources, err := r.extractResources(instance, ctx, roleGroupExtractor)
	if err != nil {
		return err
	}

	for _, rsc := range resources {
		if rsc == nil {
			continue
		}

		if err := CreateOrUpdate(ctx, r.Client, rsc); err != nil {
			r.Log.Error(err, "Failed to create or update Resource", "resource", rsc)
			return err
		}
	}
	return nil
}

func (r *ArgoWorkFlowReconciler) fetchResource(ctx context.Context, obj client.Object,
	instance *stackv1alpha1.ArgoWorkFlow) error {
	name := obj.GetName()
	kind := obj.GetObjectKind()
	if err := r.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: name}, obj); err != nil {
		opt := []any{"ns", instance.Namespace, "name", name, "kind", kind}
		if apierrors.IsNotFound(err) {
			r.Log.Error(err, "Fetch resource NotFound", opt...)
		} else {
			r.Log.Error(err, "Fetch resource occur some unknown err", opt...)
		}
		return err
	}
	return nil
}
