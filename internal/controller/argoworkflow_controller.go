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

package controller

import (
	"context"

	"github.com/go-logr/logr"
	stackv1alpha1 "github.com/zncdatadev/argo-workflow-operator/api/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/status"
	"github.com/zncdatadev/operator-go/pkg/util"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ArgoWorkFlowReconciler reconciles a ArgoWorkFlow object
type ArgoWorkFlowReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=stack.zncdata.net,resources=argoworkflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stack.zncdata.net,resources=argoworkflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=stack.zncdata.net,resources=argoworkflows/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims;persistentvolumeclaims/finalizers,verbs=get;create;update;delete
// +kubebuilder:rbac:groups="",resources=pods;pods/exec,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=workflows;workflows/finalizers;workflowtasksets;workflowtasksets/finalizers;workflowartifactgctasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=workflowtemplates;workflowtemplates/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=argoproj.io,resources=cronworkflows;cronworkflows/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=argoproj.io,resources=workflowtaskresults,verbs=list;watch;deletecollection
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list
// +kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=create;get;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=create
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,resourceNames=workflow-controller;workflow-controller-lease,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ArgoWorkFlow object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ArgoWorkFlowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.Log.Info("Reconciling ArgoWorkFlow")

	instance := &stackv1alpha1.ArgoWorkFlow{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if client.IgnoreNotFound(err) != nil {
			r.Log.Error(err, "unable to fetch ArgoWorkFlow")
			return ctrl.Result{}, err
		}
		r.Log.Info("ArgoWorkFlow resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	}

	// Get the status condition, if it exists and its generation is not the
	//same as the ArgoWorkFlow's generation, reset the status conditions
	readCondition := apimeta.FindStatusCondition(instance.Status.Conditions, status.ConditionTypeProgressing)
	if readCondition == nil || readCondition.ObservedGeneration != instance.GetGeneration() {
		instance.InitStatusConditions()

		if err := util.UpdateStatus(ctx, r.Client, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.Log.Info("ArgoWorkFlow found", "Name", instance.Name)

	tasks := []ReconcileTask[*stackv1alpha1.ArgoWorkFlow]{
		{
			resourceName:  Deployment,
			reconcileFunc: r.reconcileDeployment,
		},
		{
			resourceName:  Service,
			reconcileFunc: r.reconcileService,
		},
		{
			resourceName:  ServiceAccount,
			reconcileFunc: r.reconcileServiceAccount,
		},
		{
			resourceName:  RoleBinding,
			reconcileFunc: r.reconcileClusterRoleBinding,
		},
		{
			resourceName:  ConfigMap,
			reconcileFunc: r.reconcileConfigMap,
		},
	}

	if err := ReconcileTasks(&tasks, ctx, instance, r, instance.Status, "hiveMetaStore"); err != nil {
		return ctrl.Result{}, err
	}
	r.Log.Info("Successfully reconciled ArgoWorkFlow")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArgoWorkFlowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stackv1alpha1.ArgoWorkFlow{}).
		Complete(r)
}
