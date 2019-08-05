package routedip

import (
	"context"

	routedipoperatorv1alpha1 "github.com/jcastillo2nd/routed-ip-operator/pkg/apis/routedipoperator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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

var log = logf.Log.WithName("controller_routedip")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new RoutedIP Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRoutedIP{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("routedip-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RoutedIP
	err = c.Watch(&source.Kind{Type: &routedipoperatorv1alpha1.RoutedIP{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner RoutedIP
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &routedipoperatorv1alpha1.RoutedIP{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRoutedIP implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRoutedIP{}

// ReconcileRoutedIP reconciles a RoutedIP object
type ReconcileRoutedIP struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a RoutedIP object and makes changes based on the state read
// and what is in the RoutedIP.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRoutedIP) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling RoutedIP")

	// Fetch the RoutedIP instance
	instance := &routedipoperatorv1alpha1.RoutedIP{}
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

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set RoutedIP instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// isPodSelectorService checks if the Service is mapped to Pods
func isPodSelectorService(svc *v1.Service) bool {
	// Check for type
	if (svc.Spec.Type == corev1.ServiceTypeNodePort) ||
	   (svc.Spec.Type == corev1.ServiceTypeClusterIP) {
		return true
	}
	return false
}

// getServiceNodes returns a list of Node names from Endpoints
func (r *ReconcileRoutedIP) getServiceNodes(svc *v1.Service) []string {
	// Start with getting endpoints matching the same Namespace and Name
	// https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#Reader
	ep = corev1.Endpoints{}
	// TODO: This should be a Get, if Service <-> Endpoints is a 1:1 relationship
	err := r.client.List(context.TODO(), ep, corev1.ListOption{Name: svc.Name})
        if err != nil {
		// Abort, endpoints don't exist
	}
	var nodes []string
        nodeSet := make(map[string]struct{}, len(nodes))
	// Iterate all Subsets in Endpoints
        for _, subset := range ep.Subsets {
		for _, address := range subset.Addresses {
			_, in = nodeSet[address.NodeName]
			if !in {
				// TODO, add nodename to set
				inSet[]
			}
		}
	}
	// Create Node list from ep.[]Subsets.[]Addresses.NodeName
	nodes := make([]string, len(set))
	for k, v := range set {
		nodes = append(nodes, k)
	}
        return nodes	
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *routedipoperatorv1alpha1.RoutedIP) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
