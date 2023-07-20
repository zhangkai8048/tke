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

// Code generated by lister-gen. DO NOT EDIT.

package internalversion

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	chart "tkestack.io/tke/api/chart"
)

// ChartLister helps list Charts.
// All objects returned here must be treated as read-only.
type ChartLister interface {
	// List lists all Charts in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*chart.Chart, err error)
	// Charts returns an object that can list and get Charts.
	Charts(namespace string) ChartNamespaceLister
	ChartListerExpansion
}

// chartLister implements the ChartLister interface.
type chartLister struct {
	indexer cache.Indexer
}

// NewChartLister returns a new ChartLister.
func NewChartLister(indexer cache.Indexer) ChartLister {
	return &chartLister{indexer: indexer}
}

// List lists all Charts in the indexer.
func (s *chartLister) List(selector labels.Selector) (ret []*chart.Chart, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*chart.Chart))
	})
	return ret, err
}

// Charts returns an object that can list and get Charts.
func (s *chartLister) Charts(namespace string) ChartNamespaceLister {
	return chartNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ChartNamespaceLister helps list and get Charts.
// All objects returned here must be treated as read-only.
type ChartNamespaceLister interface {
	// List lists all Charts in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*chart.Chart, err error)
	// Get retrieves the Chart from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*chart.Chart, error)
	ChartNamespaceListerExpansion
}

// chartNamespaceLister implements the ChartNamespaceLister
// interface.
type chartNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Charts in the indexer for a given namespace.
func (s chartNamespaceLister) List(selector labels.Selector) (ret []*chart.Chart, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*chart.Chart))
	})
	return ret, err
}

// Get retrieves the Chart from the indexer for a given namespace and name.
func (s chartNamespaceLister) Get(name string) (*chart.Chart, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(chart.Resource("chart"), name)
	}
	return obj.(*chart.Chart), nil
}
