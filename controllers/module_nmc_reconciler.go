package controllers

import (
	"context"
	"fmt"
	"strings"

	kmmv1beta1 "github.com/kubernetes-sigs/kernel-module-management/api/v1beta1"
	"github.com/kubernetes-sigs/kernel-module-management/internal/api"
	"github.com/kubernetes-sigs/kernel-module-management/internal/filter"
	"github.com/kubernetes-sigs/kernel-module-management/internal/module"
	"github.com/kubernetes-sigs/kernel-module-management/internal/nmc"
	"github.com/kubernetes-sigs/kernel-module-management/internal/registry"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//+kubebuilder:rbac:groups="core",resources=nodes,verbs=get;watch;patch
//+kubebuilder:rbac:groups=kmm.sigs.x-k8s.io,resources=nodemodulestates,verbs=get;list;watch;update;patch;create
//+kubebuilder:rbac:groups=kmm.sigs.x-k8s.io,resources=nodemodulestates/status,verbs=get;update;patch

const (
	ModuleNMCReconcilerName = "ModuleNMCConfig"
)

type ModuleNMCReconciler struct {
	client      client.Client
	kernelAPI   module.KernelMapper
	registryAPI registry.Registry
	filter      *filter.Filter
	reconHelper moduleNMCReconcilerHelperAPI
}

func NewModuleNMCReconciler(client client.Client,
	kernelAPI module.KernelMapper,
	registryAPI registry.Registry,
	nmcHelper nmc.Helper,
	filter *filter.Filter) *ModuleNMCReconciler {
	reconHelper := newModuleNMCReconcilerHelper(client, nmcHelper, registryAPI)
	return &ModuleNMCReconciler{
		client:      client,
		kernelAPI:   kernelAPI,
		filter:      filter,
		reconHelper: reconHelper,
	}
}

func (mnr *ModuleNMCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// get the reconciled module
	logger := log.FromContext(ctx)

	logger.Info("Starting Module-NMS reconcilation", "module name and namespace", req.NamespacedName)

	mod, err := mnr.reconHelper.getRequestedModule(ctx, req.NamespacedName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Module deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get the requested %s KMMO CR: %w", req.NamespacedName, err)
	}

	// get all nodes
	nodes, err := mnr.reconHelper.getNodesList(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get nodes list: %v", err)
	}

	var sumErr error
	for _, node := range nodes {
		kernelVersion := strings.TrimSuffix(node.Status.NodeInfo.KernelVersion, "+")
		mld, _ := mnr.kernelAPI.GetModuleLoaderDataForKernel(mod, kernelVersion)

		shouldBeOnNode, err := mnr.reconHelper.shouldModuleRunOnNode(node, mld)
		if err != nil {
			sumErr = fmt.Errorf("%v: %v", sumErr, err)
			continue
		}

		if shouldBeOnNode {
			err = mnr.reconHelper.enableModuleOnNode(ctx, mld, node)
		} else {
			err = mnr.reconHelper.disableModuleOnNode(ctx, mod.Name, mod.Namespace, node)
		}
		if err != nil {
			sumErr = fmt.Errorf("%v: %v", sumErr, err)
		}
	}

	if sumErr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile module %s/%s with nodes: %v", mod.Namespace, mod.Name, sumErr)
	}
	return ctrl.Result{}, nil
}

type moduleNMCReconcilerHelperAPI interface {
	getRequestedModule(ctx context.Context, namespacedName types.NamespacedName) (*kmmv1beta1.Module, error)
	getNodesList(ctx context.Context) ([]v1.Node, error)
	shouldModuleRunOnNode(node v1.Node, mld *api.ModuleLoaderData) (bool, error)
	enableModuleOnNode(ctx context.Context, mld *api.ModuleLoaderData, node v1.Node) error
	disableModuleOnNode(ctx context.Context, modName, modNamespace string, node v1.Node) error
}

type moduleNMCReconcilerHelper struct {
	client      client.Client
	nmcHelper   nmc.Helper
	registryAPI registry.Registry
}

func newModuleNMCReconcilerHelper(client client.Client,
	nmcHelper nmc.Helper,
	registryAPI registry.Registry) moduleNMCReconcilerHelperAPI {
	return &moduleNMCReconcilerHelper{
		client:      client,
		nmcHelper:   nmcHelper,
		registryAPI: registryAPI,
	}
}

func (mnrh *moduleNMCReconcilerHelper) getRequestedModule(ctx context.Context, namespacedName types.NamespacedName) (*kmmv1beta1.Module, error) {
	mod := kmmv1beta1.Module{}

	if err := mnrh.client.Get(ctx, namespacedName, &mod); err != nil {
		return nil, fmt.Errorf("failed to get the kmmo module %s: %w", namespacedName, err)
	}
	return &mod, nil
}

func (mnrh *moduleNMCReconcilerHelper) getNodesList(ctx context.Context) ([]v1.Node, error) {
	nodes := v1.NodeList{}
	err := mnrh.client.List(ctx, &nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of nodes: %v", err)
	}
	return nodes.Items, nil
}

func (mnrh *moduleNMCReconcilerHelper) shouldModuleRunOnNode(node v1.Node, mld *api.ModuleLoaderData) (bool, error) {
	if mld == nil {
		return false, nil
	}
	nodeLabelsSet := labels.Set(node.GetLabels())
	sel := labels.NewSelector()

	for k, v := range mld.Selector {
		requirement, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
			return false, fmt.Errorf("failed to create new label requirements for mo %s/%s: %v", mld.Namespace, mld.Name, err)
		}
		sel = sel.Add(*requirement)
	}

	return sel.Matches(nodeLabelsSet), nil
}

func (mnrh *moduleNMCReconcilerHelper) enableModuleOnNode(ctx context.Context, mld *api.ModuleLoaderData, node v1.Node) error {
	logger := log.FromContext(ctx)
	// check the image existence
	exists, err := module.ImageExists(ctx, mnrh.client, mnrh.registryAPI, mld, mld.Namespace, mld.ContainerImage)
	if err != nil {
		return fmt.Errorf("failed to verify is image %s exists: %v", mld.ContainerImage, err)
	}
	if !exists {
		// skip updating NMC, reconciliation will kick in once the build job is completed
		return nil
	}
	moduleConfig := kmmv1beta1.ModuleConfig{
		ContainerImage:       mld.ContainerImage,
		InTreeModuleToRemove: mld.InTreeModuleToRemove,
		Modprobe:             mld.Modprobe,
	}

	nmc := &kmmv1beta1.NodeModulesConfig{
		ObjectMeta: metav1.ObjectMeta{Name: node.Name},
	}

	opRes, err := controllerutil.CreateOrPatch(ctx, mnrh.client, nmc, func() error {
		return mnrh.nmcHelper.SetNMCAsDesired(ctx, nmc, mld.Namespace, mld.Name, &moduleConfig)
	})

	if err == nil {
		logger.Info("enable module in NMC", "name", mld.Name, "namespace", mld.Namespace, "node", node.Name, "result", opRes)
	}
	return err
}

func (mnrh *moduleNMCReconcilerHelper) disableModuleOnNode(ctx context.Context, modName, modNamespace string, node v1.Node) error {
	logger := log.FromContext(ctx)
	nmc, err := mnrh.nmcHelper.Get(ctx, node.Name)
	if err != nil {
		return fmt.Errorf("failed to get the NodeModulesConfig for node %s: %v", node.Name, err)
	}
	if nmc == nil {
		// NodeModulesConfig does not exists, module was never running on the node, we are good
		return nil
	}

	opRes, err := controllerutil.CreateOrPatch(ctx, mnrh.client, nmc, func() error {
		return mnrh.nmcHelper.SetNMCAsDesired(ctx, nmc, modNamespace, modName, nil)
	})

	if err == nil {
		logger.Info("disable module in NMC", "name", modName, "namespace", modNamespace, "node", node.Name, "result", opRes)
	}
	return err
}

func (mnr *ModuleNMCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.
		NewControllerManagedBy(mgr).
		For(&kmmv1beta1.Module{}).
		Owns(&kmmv1beta1.NodeModulesConfig{}).
		Watches(
			&v1.Node{},
			handler.EnqueueRequestsFromMapFunc(mnr.filter.FindModulesForNMCNodeChange),
			builder.WithPredicates(
				filter.NMCReconcilerNodePredicate(),
			),
		).
		Watches(
			&batchv1.Job{},
			handler.EnqueueRequestsFromMapFunc(mnr.filter.FindModuleForJobs),
			builder.WithPredicates(
				filter.NMCReconcileJobPredicate(),
			),
		).
		Named(ModuleNMCReconcilerName).
		Complete(mnr)
}
