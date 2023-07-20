/*
 * Tencent is pleased to support the open source community by making TKEStack
 * available.
 *
 * Copyright (C) 2012-2020 Tencent. All Rights Reserved.
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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	chart "tkestack.io/tke/api/chart"
)

// FakeCharts implements ChartInterface
type FakeCharts struct {
	Fake *FakeChart
	ns   string
}

var chartsResource = schema.GroupVersionResource{Group: "chart.tkestack.io", Version: "", Resource: "charts"}

var chartsKind = schema.GroupVersionKind{Group: "chart.tkestack.io", Version: "", Kind: "Chart"}

// Get takes name of the chart, and returns the corresponding chart object, and an error if there is any.
func (c *FakeCharts) Get(ctx context.Context, name string, options v1.GetOptions) (result *chart.Chart, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(chartsResource, c.ns, name), &chart.Chart{})

	if obj == nil {
		return nil, err
	}
	return obj.(*chart.Chart), err
}

// List takes label and field selectors, and returns the list of Charts that match those selectors.
func (c *FakeCharts) List(ctx context.Context, opts v1.ListOptions) (result *chart.ChartList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(chartsResource, chartsKind, c.ns, opts), &chart.ChartList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &chart.ChartList{ListMeta: obj.(*chart.ChartList).ListMeta}
	for _, item := range obj.(*chart.ChartList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested charts.
func (c *FakeCharts) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(chartsResource, c.ns, opts))

}

// Create takes the representation of a chart and creates it.  Returns the server's representation of the chart, and an error, if there is any.
func (c *FakeCharts) Create(ctx context.Context, chart *chart.Chart, opts v1.CreateOptions) (result *chart.Chart, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(chartsResource, c.ns, chart), &chart.Chart{})

	if obj == nil {
		return nil, err
	}
	return obj.(*chart.Chart), err
}

// Update takes the representation of a chart and updates it. Returns the server's representation of the chart, and an error, if there is any.
func (c *FakeCharts) Update(ctx context.Context, chart *chart.Chart, opts v1.UpdateOptions) (result *chart.Chart, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(chartsResource, c.ns, chart), &chart.Chart{})

	if obj == nil {
		return nil, err
	}
	return obj.(*chart.Chart), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCharts) UpdateStatus(ctx context.Context, chart *chart.Chart, opts v1.UpdateOptions) (*chart.Chart, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(chartsResource, "status", c.ns, chart), &chart.Chart{})

	if obj == nil {
		return nil, err
	}
	return obj.(*chart.Chart), err
}

// Delete takes name of the chart and deletes it. Returns an error if one occurs.
func (c *FakeCharts) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(chartsResource, c.ns, name), &chart.Chart{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCharts) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(chartsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &chart.ChartList{})
	return err
}

// Patch applies the patch and returns the patched chart.
func (c *FakeCharts) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *chart.Chart, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(chartsResource, c.ns, name, pt, data, subresources...), &chart.Chart{})

	if obj == nil {
		return nil, err
	}
	return obj.(*chart.Chart), err
}
