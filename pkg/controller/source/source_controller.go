package source

import (
	"context"
	artv1alpha1 "github.com/vfreex/release-engine-prototype/pkg/apis/art/v1alpha1"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("controller_source")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Source Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSource{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("source-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Source
	err = c.Watch(&source.Kind{Type: &artv1alpha1.Source{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Source
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &artv1alpha1.Source{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSource implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSource{}

// ReconcileSource reconciles a Source object
type ReconcileSource struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Source object and makes changes based on the state read
// and what is in the Source.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Source")

	// Fetch the Source instance
	instance := &artv1alpha1.Source{}
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

	if instance.Status.Phase == "Prepared" {
		return reconcile.Result{}, nil
	}

	instance.Status = artv1alpha1.SourceStatus{}
	instance.Status.Conditions = make(map[string]string)
	gitUri := instance.Spec.Source.Git.URI
	gitRef := instance.Spec.Source.Git.Ref

	sha1Pattern, _ := regexp.Compile("^[0-9a-f]{40}$")
	now := time.Now().UTC()
	if sha1Pattern.MatchString(gitRef) {
		instance.Status.Phase = "Prepared"
		instance.Status.Conditions["gitCommitHash"] = gitRef
		instance.Status.Conditions["revisionLockedAt"] = now.String()
	} else {
		remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
			Name: "origin",
			URLs: []string{gitUri},
		})
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			return reconcile.Result{}, err
		}
		found := false
		for _, ref := range refs {
			reqLogger.Info("ref", "name", ref.Name(), "value", ref.String())
			if ref.Name().Short() == instance.Spec.Source.Git.Ref {
				instance.Status.Phase = "Prepared"
				instance.Status.Conditions["gitCommitHash"] = ref.Hash().String()
				//instance.Status.Conditions["revisionLockedAt"] = now.String()
				found = true
				break
			}
		}
		if !found {
			instance.Status.Phase = "Failed"
		}
	}

	if err := r.client.Status().Update(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *artv1alpha1.Source) *corev1.Pod {
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
