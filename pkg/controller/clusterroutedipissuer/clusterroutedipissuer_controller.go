package clusterroutedipissuer

import (
	"context"

	routedipoperatorv1alpha1 "github.com/jcastillo2nd/routed-ip-operator/pkg/apis/routedipoperator/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("routedipoperator").WithName("controller_clusterroutedipissuer")

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
	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForOwner{
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
	rlog := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rlog.Info("Reconciling ClusterRoutedIPIssuer")

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
		rlog.Error(err, "Unable to Get request object.")
		return reconcile.Result{}, err
	}

	// Define a new Job from updateClusterIPSpec
	updateIPJob := newUpdateClusterIPJob(instance)
	// Set ClusterRoutedIPIssuer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, updateIPJob, r.scheme); err != nil {
		rlog.Error(err, "Unable to set job controller reference.", "job", updateIPJob)
		return reconcile.Result{}, err
	}
	rlog.Info("Created IP update job.", "job", updateIPJob)

	result := reconcile.Result{}

	// Check if this Job already exists
	result, err = retryJobUntilSuccess(r, updateIPJob)

	if err != nil {
		return result, err
	}
	rlog.Info("IP update job succeeded.", "job", updateIPJob)

	// The UpdateIPJob has succeeded, now we need to update the firewall
	// Define a new Job from updateFirewallSpec
	updateFirewallJob := newUpdateFirewallJob(instance)
	// Set ClusterRoutedIPIssuer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, updateFirewallJob, r.scheme); err != nil {
		rlog.Error(err, "Unable to set job controller reference.", "job", updateFirewallJob)
		return reconcile.Result{}, err
	}
	rlog.Info("Created Firewall update job.", "job", updateFirewallJob)

	result, err = retryJobUntilSuccess(r, updateIPJob)

	if err != nil {
		return reconcile.Result{}, nil
	}
	rlog.Info("Firewall update job succeeded.", "job", updateFirewallJob)

	// Both updateIPJob and updateFirewallJob have succeeded
	err = r.client.Delete(context.Background(), updateFirewallJob)

	if err != nil {
		rlog.Error(err, "Unable to clean up firewall update job.", "job", updateFirewallJob)
		return reconcile.Result{}, err
	}

	err = r.client.Delete(context.Background(), updateIPJob)

	if err != nil {
		rlog.Error(err, "Unable to clean up IP update job.", "job", updateIPJob)
		return reconcile.Result{}, err
	}
	rlog.Info("Cleaned up jobs.")
	return reconcile.Result{}, nil
}

// newUpdateClusterIPJob returns a Job with the JobSpec designated for the ClusterRoutedIPIssuer
func newUpdateClusterIPJob(cr *routedipoperatorv1alpha1.ClusterRoutedIPIssuer) *batchv1.Job {
	labels := map[string]string{
		"app":    cr.Name,
		"update": "RoutedIP",
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-ip-job",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: cr.Spec.UpdateRoutedIPSpec,
	}
}

// newUpdateFirewallJob returns a Job with the JobSpec designated for the ClusterRoutedIPIssuer
func newUpdateFirewallJob(cr *routedipoperatorv1alpha1.ClusterRoutedIPIssuer) *batchv1.Job {
	labels := map[string]string{
		"app":    cr.Name,
		"update": "firewall",
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-fw-job",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: cr.Spec.UpdateFirewallSpec,
	}
}

// retryJobUntilSuccees continually returns non-nil err until Job completion
func retryJobUntilSuccess(r *ReconcileClusterRoutedIPIssuer, j *batchv1.Job) (reconcile.Result, error) {
	// Check if this Job already exists
	found := &batchv1.Job{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: j.Name, Namespace: j.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Job not found, creating.", "job", j)
		err = r.client.Create(context.TODO(), j)
		if err != nil {
			log.Error(err, "Failed creating.", "job", j)
			return reconcile.Result{}, err
		}

		// Job created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get job.", "job", j)
		return reconcile.Result{}, err
	}

	// Job already exists - Check if we need to retry, or run through FirewallUpdate
	retry := false
	for i := 0; i < len(found.Status.Conditions); i++ {
		if found.Status.Conditions[i].Status == "True" {
			if found.Status.Conditions[i].Type == batchv1.JobFailed {
				log.Info("Job failed, retrying.", "job", found)
				retry = true
				break
			}
			if found.Status.Conditions[i].Type == batchv1.JobComplete {
				if found.Status.Succeeded > 0 {
					log.Info("Job succeeded.", "job", found)
					break
				}
			}
		}
	}

	if retry {
		// We need to Retry, Delete current Job, and requeue
		err = r.client.Delete(context.TODO(), found)
		if err != nil {
			// Return client Delete error
			log.Error(err, "Failed to delete failed job.", "job", found)
			return reconcile.Result{}, err
		}
		err := errors.NewBadRequest("Retrying job.")
		// Return Job Failed, retrying error
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
