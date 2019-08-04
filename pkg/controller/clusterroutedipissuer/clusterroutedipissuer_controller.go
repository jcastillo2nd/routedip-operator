package clusterroutedipissuer

import (
	"context"
	"errors"

	routedipoperatorv1alpha1 "github.com/jcastillo2nd/routed-ip-operator/pkg/apis/routedipoperator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	batch "k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_clusterroutedipissuer")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ClusterRoutedIPIssuer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileClusterRoutedIPIssuer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterroutedipissuer-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClusterRoutedIPIssuer
	err = c.Watch(&source.Kind{Type: &routedipoperatorv1alpha1.ClusterRoutedIPIssuer{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &routedipoperatorv1alpha1.ClusterRoutedIPIssuer{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Jobs and requeue the owner ClusterRoutedIPIssuer
	err = c.Watch(&source.Kind{Type: &batch.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &routedipoperatorv1alpha1.ClusterRoutedIPIssuer{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileClusterRoutedIPIssuer implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileClusterRoutedIPIssuer{}

// ReconcileClusterRoutedIPIssuer reconciles a ClusterRoutedIPIssuer object
type ReconcileClusterRoutedIPIssuer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ClusterRoutedIPIssuer object and makes changes based on the state read
// and what is in the ClusterRoutedIPIssuer.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterRoutedIPIssuer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ClusterRoutedIPIssuer")

	// Fetch the ClusterRoutedIPIssuer instance
	instance := &routedipoperatorv1alpha1.ClusterRoutedIPIssuer{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Job from updateClusterIPSpec
	updateIPJob := newUpdateClusterIPJob(instance)
	// Set ClusterRoutedIPIssuer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, updateIPJob, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	result := reconcile.Result{}

	// Check if this Job already exists
	result, err = retryJobUntilSuccess(r, updateIPJob)

	if err != nil {
		return result, err
	}

	// The UpdateIPJob has succeeded, now we need to update the firewall
	// Define a new Job from updateFirewallSpec
	updateFirewallJob := newUpdateFirewallJob(instance)
	// Set ClusterRoutedIPIssuer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, updateFirewallJob, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	result, err = retryJobUntilSuccess(r, updateIPJob)

	if err != nil {
		return reconcile.Result{}, nil
	}

	// Both updateIPJob and updateFirewallJob have succeeded
	_ = r.Client.Delete(context.Background(), updateIPJob)
	_ = r.Client.Delete(context.Background(), updateFirewallJob)
	return reconcile.Result{}, nil
}

// newUpdateClusterIPJob returns a Job with the JobSpec designated for the ClusterRoutedIPIssuer
func newUpdateClusterIPJob(cr *routedipoperatorv1alpha1.ClusterRoutedIPIssuer.ClusterRoutedIPIssuer) *batch.Job {
	labels := map[string]string{
		"app":     cr.Name,
		"update": "RoutedIP",
	}
	job := &batch.Job {
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-ip-job",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: cr.UpdateRoutedIPSpec,
	}
	if err := controllerutil.SetControllerReference(instance, job, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
}

// newUpdateFirewallJob returns a Job with the JobSpec designated for the ClusterRoutedIPIssuer
func newUpdateFirewallJob(cr *routedipoperatorv1alpha1.ClusterRoutedIPIssuer.ClusterRoutedIPIssuer) *batch.Job {
	labels := map[string]string{
		"app":     cr.Name,
		"update": "firewall",
	}
	return &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-fw-job",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: cr.UpdateFirewallSpec,
	}
}

// retryJobUntilSuccees continually returns non-nil err until Job completion
func retryJobUntilSuccess(r *ReconcileClusterRoutedIPIssuer, j *batch.Job) reconcile.Result, error {
	// Check if this Job already exists
	found := &batch.Job{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: j.Name, Namespace: j.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Job", "Job.Namespace", j.Namespace, "Job.Name", j.Name)
		err = r.client.Create(context.TODO(), j)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Job created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Job already exists - Check if we need to retry, or run through FirewallUpdate
	retry bool := false
        for i := 0; i < len(found.Status.Conditions); i++{
		if found.Status.Conditions[i].Status == corev1.ConditionStatus.ConditionTrue{
			if found.Status.Conditions[i].Type == batch.JobConditionType.JobFailed{
				retry = true
				break
			}
			if found.Status.Conditions[i].Type == batch.JobConditionType.JobComplete{
				if foundUpdateIPJob.Status.Succeeded > 0{
					break
				}
			}
		}
	}

	if retry {
		// We need to Retry, Delete current Job, and requeue
		reqLogger.Warn("Job failed, retrying.", "Job.Namespace", found.Namespace, "Job.Name", found.Name)
		err = r.client.Delete(context.TODO(), foundUpdateIPJob)
		if err != nil {
			// Return client Delete error
			return reconcile.Result{}, err
		}
		err := errors.New("Job failed, retrying.")
		// Return Job Failed, retrying error
		return reconcile.Result{}, err
        }
}
