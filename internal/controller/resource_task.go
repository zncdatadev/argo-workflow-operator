package controller

import (
	"context"
	"fmt"

	"github.com/zncdatadev/operator-go/pkg/status"
	"github.com/zncdatadev/operator-go/pkg/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceType string

const (
	Deployment     ResourceType = "Deployment"
	Service        ResourceType = "Service"
	Secret         ResourceType = "Secret"
	ConfigMap      ResourceType = "ConfigMap"
	Pvc            ResourceType = "Pvc"
	ServiceAccount ResourceType = "ServiceAccount"
	RoleBinding                 = "RoleBinding"
)

var (
	ResourceMapper = map[ResourceType]string{
		Deployment:     status.ConditionTypeReconcileDeployment,
		Service:        status.ConditionTypeReconcileService,
		Secret:         status.ConditionTypeReconcileSecret,
		ConfigMap:      status.ConditionTypeReconcileConfigMap,
		Pvc:            status.ConditionTypeReconcilePVC,
		ServiceAccount: "ReconcileServiceAccount", //todo: update opgo
		RoleBinding:    "ReconcileRoleBinding",
	}
)

type ReconcileTask[T client.Object] struct {
	resourceName  ResourceType
	reconcileFunc func(ctx context.Context, instance T) error
}

func ReconcileTasks[T client.Object](tasks *[]ReconcileTask[T], ctx context.Context, instance T,
	r *ArgoWorkFlowReconciler, instanceStatus status.Status, serverName string) error {
	for _, task := range *tasks {
		jobFunc := task.reconcileFunc
		if err := jobFunc(ctx, instance); err != nil {
			r.Log.Error(err, fmt.Sprintf("unable to reconcile %s", task.resourceName))
			return err
		}
		if updated := instanceStatus.SetStatusCondition(v1.Condition{
			Type:               ResourceMapper[ResourceType(task.resourceName)],
			Status:             v1.ConditionTrue,
			Reason:             status.ConditionReasonRunning,
			Message:            createSuccessMessage(serverName, task.resourceName),
			ObservedGeneration: instance.GetGeneration(),
		}); updated {
			err := util.UpdateStatus(ctx, r.Client, instance)
			if err != nil {
				r.Log.Error(err, createUpdateErrorMessage(task.resourceName))
				return err
			}
		}
	}
	return nil
}

func createSuccessMessage(serverName string, resourceName ResourceType) string {
	return fmt.Sprintf("%sServer's %s is running", serverName, resourceName)
}

func createUpdateErrorMessage(resourceName ResourceType) string {
	return fmt.Sprintf("unable to update status for %s", resourceName)
}
