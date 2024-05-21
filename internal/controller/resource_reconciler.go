package controller

import (
	"context"

	stackv1alpha1 "github.com/zncdatadev/argo-workflow-operator/api/v1alpha1"
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
