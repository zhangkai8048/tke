/*
 * Tencent is pleased to support the open source community by making TKEStack
 * available.
 *
 * Copyright (C) 2012-2019 Tencent. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use
 * this file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/Apache-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OF ANY KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations under the License.
 */

package chart

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"helm.sh/chartmuseum/pkg/chartmuseum/server/multitenant"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	chartv1 "tkestack.io/tke/api/chart/v1"
	clientset "tkestack.io/tke/api/client/clientset/versioned"
	chartv1informer "tkestack.io/tke/api/client/informers/externalversions/chart/v1"
	chartv1lister "tkestack.io/tke/api/client/listers/chart/v1"
	"tkestack.io/tke/pkg/chart/controller/chart/deletion"
	controllerutil "tkestack.io/tke/pkg/controller"
	helm "tkestack.io/tke/pkg/registry/harbor/helmClient"
	"tkestack.io/tke/pkg/util/log"
)

const (
	// chartDeletionGracePeriod is the time period to wait before processing a received channel event.
	// This allows time for the following to occur:
	// * lifecycle admission plugins on HA apiservers to also observe a channel
	//   deletion and prevent new objects from being created in the terminating channel
	// * non-leader etcd servers to observe last-minute object creations in a channel
	//   so this controller's cleanup can actually clean up all objects
	chartDeletionGracePeriod = 5 * time.Second
)

const (
	controllerName = "chart-controller"
)

// Controller is responsible for performing actions dependent upon an chart phase.
type Controller struct {
	client       clientset.Interface
	cache        *chartCache
	health       *chartHealth
	queue        workqueue.RateLimitingInterface
	lister       chartv1lister.ChartLister
	listerSynced cache.InformerSynced
	stopCh       <-chan struct{}
	// helper to delete all resources in the chart when the chart is deleted.
	chartResourcesDeleter deletion.ChartResourcesDeleterInterface
}

// NewController creates a new Controller object.
func NewController(
	client clientset.Interface, chartInformer chartv1informer.ChartInformer,
	resyncPeriod time.Duration, finalizerToken chartv1.FinalizerName,
	multiTenantServer *multitenant.MultiTenantServer, helmClient *helm.APIClient) *Controller {
	// create the controller so we can inject the enqueue function
	controller := &Controller{
		client:                client,
		cache:                 &chartCache{m: make(map[string]*cachedChart)},
		health:                &chartHealth{charts: sets.NewString()},
		queue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		chartResourcesDeleter: deletion.NewChartResourcesDeleter(client.ChartV1(), multiTenantServer, finalizerToken, true, helmClient),
	}

	// if client != nil &&
	// 	client.ChartV1().RESTClient() != nil &&
	// 	!reflect.ValueOf(client.chartv1().RESTClient()).IsNil() &&
	// 	client.ChartV1().RESTClient().GetRateLimiter() != nil {
	// 	_ = metrics.RegisterMetricAndTrackRateLimiterUsage("chart_controller", client.ChartV1().RESTClient().GetRateLimiter())
	// }

	chartInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: controller.enqueue,
			UpdateFunc: func(oldObj, newObj interface{}) {
				old, ok1 := oldObj.(*chartv1.Chart)
				cur, ok2 := newObj.(*chartv1.Chart)
				if ok1 && ok2 && controller.needsUpdate(old, cur) {
					controller.enqueue(newObj)
				}
			},
			DeleteFunc: controller.enqueue,
		},
		resyncPeriod,
	)
	controller.lister = chartInformer.Lister()
	controller.listerSynced = chartInformer.Informer().HasSynced

	return controller
}

// obj could be an *chartv1.Chart, or a DeletionFinalStateUnknown marker item.
func (c *Controller) enqueue(obj interface{}) {
	key, err := controllerutil.KeyFunc(obj)
	if err != nil {
		log.Error("Couldn't get key for object", log.Any("object", obj), log.Err(err))
		return
	}
	c.queue.AddAfter(key, chartDeletionGracePeriod)
}

func (c *Controller) needsUpdate(old *chartv1.Chart, new *chartv1.Chart) bool {
	if old.UID != new.UID {
		return true
	}

	if !reflect.DeepEqual(old.Spec, new.Spec) {
		return true
	}

	if !reflect.DeepEqual(old.Status, new.Status) {
		return true
	}

	return false
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Info("Starting chart controller")
	defer log.Info("Shutting down chart controller")

	if ok := cache.WaitForCacheSync(stopCh, c.listerSynced); !ok {
		log.Error("Failed to wait for chart caches to sync")
		return
	}

	c.stopCh = stopCh
	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
}

// worker processes the queue of chart objects.
// Each chart can be in the queue at most once.
// The system ensures that no two workers can process
// the same chart at the same time.
func (c *Controller) worker() {
	workFunc := func() bool {
		key, quit := c.queue.Get()
		if quit {
			return true
		}
		defer c.queue.Done(key)

		err := c.syncItem(key.(string))
		if err == nil {
			// no error, forget this entry and return
			c.queue.Forget(key)
			return false
		}

		// rather than wait for a full resync, re-add the chart to the queue to be processed
		c.queue.AddRateLimited(key)
		runtime.HandleError(err)
		return false
	}

	for {
		quit := workFunc()

		if quit {
			return
		}
	}
}

// syncItem will sync the Chart with the given key if it has had
// its expectations fulfilled, meaning it did not expect to see any more of its
// charts created or deleted. This function is not meant to be invoked
// concurrently with the same key.
func (c *Controller) syncItem(key string) error {
	startTime := time.Now()
	defer func() {
		log.Info("Finished syncing chart", log.String("chart", key), log.Duration("processTime", time.Since(startTime)))
	}()

	chartGroupName, chartName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	// chart holds the latest Chart info from apiserver
	chart, err := c.lister.Charts(chartGroupName).Get(chartName)
	switch {
	case errors.IsNotFound(err):
		log.Info("Chart has been deleted. Attempting to cleanup resources",
			log.String("chartGroupName", chartGroupName),
			log.String("chartName", chartName))
		_ = c.processDeletion(key)
		return nil
	case err != nil:
		log.Warn("Unable to retrieve chart from store",
			log.String("chartGroupName", chartGroupName),
			log.String("chartName", chartName), log.Err(err))
		return err
	default:
		cachedChart := c.cache.getOrCreate(key)
		if c.needCompatibleUpgrade(context.Background(), chart) {
			return c.compatibleUpgrade(context.Background(), key, cachedChart, chart)
		}

		if chart.Status.Phase == chartv1.ChartPending ||
			chart.Status.Phase == chartv1.ChartAvailable {
			err = c.processUpdate(context.Background(), cachedChart, chart, key)
		} else if chart.Status.Phase == chartv1.ChartTerminating {
			log.Info("Chart has been terminated. Attempting to cleanup resources",
				log.String("chartGroupName", chartGroupName),
				log.String("chartName", chartName))
			_ = c.processDeletion(key)
			err = c.chartResourcesDeleter.Delete(context.Background(), chartGroupName, chartName)
			if err == nil {
				err = c.updateChartGroup(context.Background(), chart)
			}
			// If err is not nil, do not update object status when phase is Terminating.
			// DeletionTimestamp is not empty and object will be deleted when you request updateStatus
		} else {
			log.Debug(fmt.Sprintf("Chart %s status is %s, not to process", key, chart.Status.Phase))
		}
	}
	return err
}

func (c *Controller) processDeletion(key string) error {
	cachedChart, ok := c.cache.get(key)
	if !ok {
		log.Debug("Chart not in cache even though the watcher thought it was. Ignoring the deletion", log.String("name", key))
		return nil
	}
	return c.processDelete(cachedChart, key)
}

func (c *Controller) processDelete(cachedChart *cachedChart, key string) error {
	log.Info("Chart will be dropped", log.String("name", key))

	if c.cache.Exist(key) {
		log.Info("Delete the chart cache", log.String("name", key))
		c.cache.delete(key)
	}

	if c.health.Exist(key) {
		log.Info("Delete the chart health cache", log.String("name", key))
		c.health.Del(key)
	}

	return nil
}

func (c *Controller) processUpdate(ctx context.Context, cachedChart *cachedChart, chart *chartv1.Chart, key string) error {
	if cachedChart.state != nil {
		// exist and the chart name changed
		if cachedChart.state.UID != chart.UID {
			if err := c.processDelete(cachedChart, key); err != nil {
				return err
			}
		}
	}
	// start update chart if needed
	updated, err := c.handlePhase(ctx, key, cachedChart, chart)
	if err != nil {
		return err
	}
	cachedChart.state = updated
	// Always update the cache upon success.
	c.cache.set(key, cachedChart)
	return nil
}

func (c *Controller) handlePhase(ctx context.Context, key string, cachedChart *cachedChart, chart *chartv1.Chart) (*chartv1.Chart, error) {
	newStatus := chart.Status.DeepCopy()
	switch chart.Status.Phase {
	case chartv1.ChartPending:
		newStatus.Phase = chartv1.ChartAvailable
		newStatus.Message = ""
		newStatus.Reason = ""
		newStatus.LastTransitionTime = metav1.Now()
		return c.updateStatus(ctx, chart, &chart.Status, newStatus)
	case chartv1.ChartAvailable:
		c.startChartHealthCheck(ctx, key)
	}
	return chart, nil
}

func (c *Controller) updateStatus(ctx context.Context, chart *chartv1.Chart, previousStatus, newStatus *chartv1.ChartStatus) (*chartv1.Chart, error) {
	if reflect.DeepEqual(previousStatus, newStatus) {
		return nil, nil
	}
	// Make a copy so we don't mutate the shared informer cache.
	newObj := chart.DeepCopy()
	newObj.Status = *newStatus

	updated, err := c.client.ChartV1().Charts(newObj.Namespace).UpdateStatus(ctx, newObj, metav1.UpdateOptions{})
	if err == nil {
		return updated, nil
	}
	if errors.IsNotFound(err) {
		log.Info("Not persisting update to chart that no longer exists",
			log.String("chartGroupName", chart.Namespace),
			log.String("chartName", chart.Name),
			log.Err(err))
		return updated, nil
	}
	if errors.IsConflict(err) {
		return nil, fmt.Errorf("not persisting update to chart '%s' that has been changed since we received it: %v",
			chart.Name, err)
	}
	log.Warn(fmt.Sprintf("Failed to persist updated status of chart '%s/%s'",
		chart.Name, chart.Status.Phase),
		log.String("chartGroupName", chart.Namespace),
		log.String("chartName", chart.Name), log.Err(err))
	return nil, err
}

func (c *Controller) updateChartGroup(ctx context.Context, chart *chartv1.Chart) error {
	rcg, err := c.findChartGroup(ctx, chart)
	if err != nil {
		return err
	}

	rcg.Status.ChartCount = rcg.Status.ChartCount - 1
	if _, err := c.client.ChartV1().ChartGroups().UpdateStatus(ctx, rcg, metav1.UpdateOptions{}); err != nil {
		log.Error("Failed to update chartgroup while chart deleted",
			log.String("tenantID", chart.Spec.TenantID),
			log.String("chartGroupName", chart.Spec.ChartGroupName),
			log.Err(err))
		return err
	}
	return nil
}

// If need to upgrade compatibly
func (c *Controller) needCompatibleUpgrade(ctx context.Context, chart *chartv1.Chart) bool {
	switch chart.Spec.Visibility {
	case chartv1.VisibilityPrivate:
		{
			return true
		}
	}
	return false
}

// If need to upgrade compatibly
func (c *Controller) compatibleUpgrade(ctx context.Context, key string, cachedChart *cachedChart, chart *chartv1.Chart) error {
	// cg, err := c.findChartGroup(ctx, chart)
	// if err != nil {
	// 	return err
	// }

	newObj := chart.DeepCopy()
	// switch cg.Spec.Type {
	// case chartv1.RepoType(strings.ToLower(string(chartv1.RepoTypePersonal))):
	// 	{
	// 		if cg.Spec.Visibility == chartv1.VisibilityPrivate {
	// 			newObj.Spec.Visibility = chartv1.VisibilityUser
	// 		}
	// 		break
	// 	}
	// case chartv1.RepoType(strings.ToLower(string(chartv1.RepoTypeProject))):
	// 	{
	// 		if cg.Spec.Visibility == chartv1.VisibilityPrivate {
	// 			newObj.Spec.Visibility = chartv1.VisibilityProject
	// 		}
	// 		break
	// 	}
	// default:
	// 	return nil
	// }

	updated, err := c.client.ChartV1().Charts(newObj.Namespace).Update(ctx, newObj, metav1.UpdateOptions{})
	if errors.IsNotFound(err) {
		log.Info("Not persisting upgrade to chart that no longer exists",
			log.String("chartGroupName", chart.Spec.ChartGroupName),
			log.String("chartName", chart.Name),
			log.Err(err))
		return nil
	}

	cachedChart.state = updated
	// Always update the cache upon success.
	c.cache.set(key, cachedChart)
	return nil
}

func (c *Controller) findChartGroup(ctx context.Context, chart *chartv1.Chart) (*chartv1.ChartGroup, error) {
	chartGroupList, err := c.client.ChartV1().ChartGroups().List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.tenantID=%s,spec.name=%s", chart.Spec.TenantID, chart.Spec.ChartGroupName),
	})
	if err != nil {
		log.Error("Failed to list chart group by tenantID and name",
			log.String("tenantID", chart.Spec.TenantID),
			log.String("name", chart.Spec.ChartGroupName),
			log.Err(err))
		return nil, err
	}
	if len(chartGroupList.Items) == 0 {
		// Chart group must first be created via console
		return nil, fmt.Errorf("chartgroup %s/%s not found", chart.Spec.TenantID, chart.Spec.ChartGroupName)
	}

	return chartGroupList.Items[0].DeepCopy(), nil
}
