package apiservice

import (
	"context"
	"reflect"

	apiservicev1alpha1 "github.com/ligangty/api-service/pkg/apis/apiservice/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_apiservice")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ApiService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileApiService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("apiservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ApiService
	err = c.Watch(&source.Kind{Type: &apiservicev1alpha1.ApiService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ApiService
	// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &apiservicev1alpha1.ApiService{},
	// })
	// if err != nil {
	// 	return err
	// }

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ApiService
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &apiservicev1alpha1.ApiService{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &apiservicev1alpha1.ApiService{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileApiService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileApiService{}

// ReconcileApiService reconciles a ApiService object
type ReconcileApiService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ApiService object and makes changes based on the state read
// and what is in the ApiService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileApiService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ApiService")

	// Fetch the ApiService instance
	apiService := &apiservicev1alpha1.ApiService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, apiService)
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

	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: apiService.Name, Namespace: apiService.Namespace}, deployment)
	if err != nil {
		// Define a new Deployment
		dep := r.deploymentForApiService(apiService)
		reqLogger.Info("Creating a new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		// NOTE: that the requeue is made with the purpose to provide the deployment object for the next step to ensure the deployment size is the same as the spec.
		// Also, you could GET the deployment object again instead of requeue if you wish. See more over it here: https://godoc.org/sigs.k8s.io/controller-runtime/pkg/reconcile#Reconciler
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment.")
		return reconcile.Result{}, err
	}

	// Check if the Service already exists, if not create a new one
	// NOTE: The Service is used to expose the Deployment. However, the Service is not required at all for the apiService example to work. The purpose is to add more examples of what you can do in your operator project.
	service := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: apiService.Name, Namespace: apiService.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service object
		ser := r.serviceForApiService(apiService)
		reqLogger.Info("Creating a new Service.", "Service.Namespace", ser.Namespace, "Service.Name", ser.Name)
		err = r.client.Create(context.TODO(), ser)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service.", "Service.Namespace", ser.Namespace, "Service.Name", ser.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service.")
		return reconcile.Result{}, err
	}

	// Update the ApiService status with the pod names
	// List the pods for this apiService's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(apiService.Namespace),
		client.MatchingLabels(labelsForApiService(apiService.Name)),
	}
	err = r.client.List(context.TODO(), podList, listOpts...)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods.", "ApiService.Namespace", apiService.Namespace, "ApiService.Name", apiService.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, apiService.Status.Nodes) {
		apiService.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), apiService)
		if err != nil {
			reqLogger.Error(err, "Failed to update ApiService status.")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileApiService) deploymentForApiService(m *apiservicev1alpha1.ApiService) *appsv1.Deployment {
	ls := labelsForApiService(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "quay.io/ligangty/test-api:latest",
						Name:    "api-service",
						Command: []string{"api-server"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "api-service",
						}},
					}},
				},
			},
		},
	}
	// Set ApiService instance as the owner of the Deployment.
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep

}

// serviceForApiService function takes in a ApiService object and returns a Service for that object.
func (r *ReconcileApiService) serviceForApiService(m *apiservicev1alpha1.ApiService) *corev1.Service {
	ls := labelsForApiService(m.Name)
	ser := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Port: 8080,
					Name: m.Name,
				},
			},
		},
	}
	// Set ApiService instance as the owner of the Service.
	controllerutil.SetControllerReference(m, ser, r.scheme)
	return ser
}

// labelsForApiService returns the labels for selecting the resources
// belonging to the given apiService CR name.
func labelsForApiService(name string) map[string]string {
	return map[string]string{"app": "apiservice", "apiservice_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}