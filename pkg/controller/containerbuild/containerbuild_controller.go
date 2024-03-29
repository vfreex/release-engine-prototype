package containerbuild

import (
	"context"
	"fmt"
	"github.com/kolo/xmlrpc"
	artv1alpha1 "github.com/vfreex/release-engine-prototype/pkg/apis/art/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
	"time"
)

var log = logf.Log.WithName("controller_containerbuild")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ContainerBuild Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileContainerBuild{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("containerbuild-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ContainerBuild
	err = c.Watch(&source.Kind{Type: &artv1alpha1.ContainerBuild{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ContainerBuild
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &artv1alpha1.ContainerBuild{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileContainerBuild implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileContainerBuild{}

// ReconcileContainerBuild reconciles a ContainerBuild object
type ReconcileContainerBuild struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ContainerBuild object and makes changes based on the state read
// and what is in the ContainerBuild.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileContainerBuild) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ContainerBuild")

	// Fetch the ContainerBuild instance
	instance := &artv1alpha1.ContainerBuild{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
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

	instance.Status = artv1alpha1.ContainerBuildStatus{}
	instance.Status.Conditions = make(map[string]string)
	instance.Status.PullSpecs = []string{}

	if instance.Spec.BuildSystem == "brew" {
		rpc, _ := xmlrpc.NewClient("https://brewhub.engineering.redhat.com/brewhub", nil)
		nvr := instance.Spec.Component + "-" + instance.Spec.Version + "-" + instance.Spec.Release
		var buildinfo struct {
			BuildId      int     `xmlrpc:"build_id"`
			OwnerName    string  `xmlrpc:"owner_name"`
			PackageName  string  `xmlrpc:"package_name"`
			State        int     `xmlrpc:"state"`
			Nvr          string  `xmlrpc:"nvr"`
			Version      string  `xmlrpc:"version"`
			Release      string  `xmlrpc:"release"`
			Epoch        string  `xmlrpc:"epoch"`
			CreationTs   float64 `xmlrpc:"creation_ts"`
			StartTs      float64 `xmlrpc:"start_ts"`
			CompletionTs float64 `xmlrpc:"completion_ts"`
			Source       string  `xmlrpc:"source"`

			Extra struct {
				Source struct {
					OriginalUrl string `xmlrpc:"original_url"`
				} `xmlrpc:"source"`
				Image struct {
					Index struct {
						UniqueTags []string `xmlrpc:"unique_tags"`
						Pull       []string `xmlrpc:"pull"`
						Digests    struct {
							ManifestList string `xmlrpc:"application/vnd.docker.distribution.manifest.list.v2+json"`
						} `xmlrpc:"digests"`
					} `xmlrpc:"index"`
				} `xmlrpc:"image"`
			} `xmlrpc:"extra"`
		}
		if err = rpc.Call("getBuild", nvr, &buildinfo); err != nil {
			return reconcile.Result{}, err
		}
		if buildinfo.State != 1 {
			instance.Status.Phase = "Error"
			return reconcile.Result{}, fmt.Errorf("Brew build state is %v", buildinfo.State)
		}
		instance.Status.Phase = "Prepared"
		instance.Status.Conditions["buildId"] = strconv.Itoa(buildinfo.BuildId)
		instance.Status.Conditions["sourceUrl"] = buildinfo.Source
		instance.Status.Conditions["creationTime"] = time.Unix(int64(buildinfo.CreationTs), 0).String()
		instance.Status.Conditions["startTime"] = time.Unix(int64(buildinfo.StartTs), 0).String()
		instance.Status.Conditions["completionTime"] = time.Unix(int64(buildinfo.CompletionTs), 0).String()
		instance.Status.PullSpecs = buildinfo.Extra.Image.Index.Pull
		if buildinfo.Extra.Image.Index.Digests.ManifestList != "" {
			instance.Status.Conditions["contentType"] = "application/vnd.docker.distribution.manifest.list.v2+json"
			instance.Status.Digest = buildinfo.Extra.Image.Index.Digests.ManifestList
		}
	}

	if err := r.client.Status().Update(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	//pod := newPodForCR(instance)
	//
	//// Set ContainerBuild instance as the owner and controller
	//if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	//	return reconcile.Result{}, err
	//}
	//
	//// Check if this Pod already exists
	//found := &corev1.Pod{}
	//err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	//if err != nil && errors.IsNotFound(err) {
	//	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	//	err = r.client.Create(context.TODO(), pod)
	//	if err != nil {
	//		return reconcile.Result{}, err
	//	}
	//
	//	// Pod created successfully - don't requeue
	//	return reconcile.Result{}, nil
	//} else if err != nil {
	//	return reconcile.Result{}, err
	//}
	//
	//// Pod already exists - don't requeue
	//reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *artv1alpha1.ContainerBuild) *corev1.Pod {
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
