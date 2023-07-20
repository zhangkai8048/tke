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

package rest

import (
	"github.com/docker/libtrust"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	restclient "k8s.io/client-go/rest"
	v1 "tkestack.io/tke/api/chart/v1"
	authversionedclient "tkestack.io/tke/api/client/clientset/versioned/typed/auth/v1"
	businessversionedclient "tkestack.io/tke/api/client/clientset/versioned/typed/business/v1"
	platformversionedclient "tkestack.io/tke/api/client/clientset/versioned/typed/platform/v1"
	"tkestack.io/tke/api/registry"
	"tkestack.io/tke/pkg/apiserver/storage"
	configmapstorage "tkestack.io/tke/pkg/chart/registry/configmap/storage"
)

// StorageProvider is a REST type for core resources storage that implement
// RestStorageProvider interface
type StorageProvider struct {
	LoopbackClientConfig *restclient.Config
	ExternalScheme       string
	ExternalHost         string
	ExternalPort         int
	ExternalCAFile       string
	PrivilegedUsername   string
	AuthClient           authversionedclient.AuthV1Interface
	BusinessClient       businessversionedclient.BusinessV1Interface
	PlatformClient       platformversionedclient.PlatformV1Interface
	Authorizer           authorizer.Authorizer
	TokenPrivateKey      libtrust.PrivateKey
}

// Implement RESTStorageProvider
var _ storage.RESTStorageProvider = &StorageProvider{}

// NewRESTStorage is a factory constructor to creates and returns the APIGroupInfo
func (s *StorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericserver.APIGroupInfo, bool) {
	apiGroupInfo := genericserver.NewDefaultAPIGroupInfo(registry.GroupName, registry.Scheme, registry.ParameterCodec, registry.Codecs)

	apiGroupInfo.VersionedResourcesStorageMap[v1.SchemeGroupVersion.Version] = s.v1Storage(apiResourceConfigSource, restOptionsGetter, s.LoopbackClientConfig)

	return apiGroupInfo, true
}

// GroupName return the api group name
func (*StorageProvider) GroupName() string {
	return registry.GroupName
}

func (s *StorageProvider) v1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter, loopbackClientConfig *restclient.Config) map[string]rest.Storage {
	// registryClient := registryinternalclient.NewForConfigOrDie(loopbackClientConfig)

	// var harborClient *harbor.APIClient = nil
	//var helmClient *helm.APIClient

	// if s.chartconfig.HarborEnabled {
	// 	tr, _ := transport.NewOneWayTLSTransport(s.chartconfig.HarborCAFile, true)
	// 	headers := make(map[string]string)
	// 	headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(
	// 		s.chartconfig.Security.AdminUsername+":"+s.chartconfig.Security.AdminPassword),
	// 	)
	// 	harborCfg := &harbor.Configuration{
	// 		BasePath:      fmt.Sprintf("https://%s/api/v2.0", s.chartconfig.DomainSuffix),
	// 		DefaultHeader: headers,
	// 		UserAgent:     "Swagger-Codegen/1.0.0/go",
	// 		HTTPClient: &http.Client{
	// 			Transport: tr,
	// 		},
	// 	}
	// 	helmCfg := &helm.Configuration{
	// 		BasePath:      fmt.Sprintf("https://%s/api", s.chartconfig.DomainSuffix),
	// 		DefaultHeader: headers,
	// 		UserAgent:     "Swagger-Codegen/1.0.0/go",
	// 		HTTPClient: &http.Client{
	// 			Transport: tr,
	// 		},
	// 	}
	// 	harborClient = harbor.NewAPIClient(harborCfg)
	// 	helmClient = helm.NewAPIClient(helmCfg)
	// }

	storageMap := make(map[string]rest.Storage)
	{

		configMapREST := configmapstorage.NewStorage(restOptionsGetter)
		storageMap["configmaps"] = configMapREST.ConfigMap

		// namespaceREST := namespacestorage.NewStorage(restOptionsGetter, registryClient, s.PrivilegedUsername, harborClient)
		// storageMap["namespaces"] = namespaceREST.Namespace
		// storageMap["namespaces/status"] = namespaceREST.Status

		// repositoryREST := repositorystorage.NewStorage(restOptionsGetter, registryClient, s.PrivilegedUsername, s.LoopbackClientConfig.Host, harborClient, s.TokenPrivateKey)
		// storageMap["repositories"] = repositoryREST.Repository
		// storageMap["repositories/status"] = repositoryREST.Status

		// chartGroupRESTStorage := chartgroupstorage.NewStorage(restOptionsGetter, registryClient, s.AuthClient, s.BusinessClient, s.PrivilegedUsername)
		// chartGroupREST := chartgroupstorage.NewREST(chartGroupRESTStorage.ChartGroup, registryClient, s.AuthClient, harborClient, helmClient)
		// repoUpdateREST := chartgroupstorage.NewRepoUpdateREST(chartGroupRESTStorage.ChartGroup, registryClient,
		// 	s.Authorizer)
		// // storageMap["chartgroups"] = chartGroupREST
		// // storageMap["chartgroups/status"] = chartGroupRESTStorage.Status
		// // storageMap["chartgroups/finalize"] = chartGroupRESTStorage.Finalize
		// // storageMap["chartgroups/repoupdating"] = repoUpdateREST

		// chartREST := chartstorage.NewStorage(restOptionsGetter, registryClient, s.AuthClient, s.BusinessClient, s.PrivilegedUsername)
		// chartVersionREST := chartstorage.NewVersionREST(chartREST.Chart, s.PlatformClient, registryClient,
		// 	s.ExternalScheme,
		// 	s.ExternalHost,
		// 	s.ExternalPort,
		// 	s.ExternalCAFile,
		// 	s.Authorizer,
		// 	helmClient)
		// storageMap["charts"] = chartREST.Chart
		// storageMap["charts/status"] = chartREST.Status
		// storageMap["charts/finalize"] = chartREST.Finalize
		// storageMap["charts/version"] = chartVersionREST
	}

	return storageMap
}
