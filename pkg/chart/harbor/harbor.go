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

package harbor

import (
	"k8s.io/apiserver/pkg/server/mux"

	restclient "k8s.io/client-go/rest"

	"tkestack.io/tke/pkg/chart/harbor/handler"
	"tkestack.io/tke/pkg/chart/harbor/tenant"

	// import filesystem driver to store images
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
	// import in memory driver to store images
	_ "github.com/docker/distribution/registry/storage/driver/inmemory"
	// import s3 object storage driver to store images
	_ "github.com/docker/distribution/registry/storage/driver/s3-aws"
	// import token authentication controller
	_ "tkestack.io/tke/pkg/chart/distribution/auth/token"
)

// RegistryPrefix defines the path prefix for accessing the docker registry v2 server.
// ChartPrefix defines the path prefix for accessing the helm chart server
const RegistryPrefix = "/v2/"
const AuthPrefix = "/service/"
const ChartAPIPrefix = "/api/chartrepo/"
const ChartPrefix = "/chartrepo/"
const CompatibleChartPrefix = "/chart/"

type Options struct {
	chartconfig          *chartconfig.chartconfiguration
	ExternalHost         string
	LoopbackClientConfig *restclient.Config
}

// IgnoredAuthPathPrefixes returns a list of path prefixes that does not need to
// go through the built-in authentication and authorization middleware of apiserver.
func IgnoreAuthPathPrefixes() []string {
	return []string{
		RegistryPrefix,
		AuthPrefix,
		ChartPrefix,
		ChartAPIPrefix,
	}
}

// RegisterRoute to register the docker distribution server path prefix to apiserver.
func RegisterRoute(m *mux.PathRecorderMux, opts *Options) error {

	httpURL := "https://" + opts.chartconfig.DomainSuffix

	harborHandler, err := handler.NewHandler(httpURL, opts.chartconfig.HarborCAFile, opts.LoopbackClientConfig, opts.chartconfig)
	if err != nil {
		return err
	}

	wrappedHandler := tenant.WithTenant(harborHandler, RegistryPrefix, AuthPrefix, ChartPrefix, CompatibleChartPrefix, ChartAPIPrefix, opts.chartconfig.DomainSuffix, opts.chartconfig.DefaultTenant)

	m.HandlePrefix(RegistryPrefix, wrappedHandler)
	m.HandlePrefix(AuthPrefix, wrappedHandler)
	m.HandlePrefix(ChartPrefix, wrappedHandler)
	m.HandlePrefix(ChartAPIPrefix, wrappedHandler)
	m.HandlePrefix(CompatibleChartPrefix, wrappedHandler)

	return nil
}
