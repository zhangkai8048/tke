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

package config

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/client-go/rest"
	versionedclientset "tkestack.io/tke/api/client/clientset/versioned"
	authversionedclient "tkestack.io/tke/api/client/clientset/versioned/typed/auth/v1"
	businessversionedclient "tkestack.io/tke/api/client/clientset/versioned/typed/business/v1"
	platformversionedclient "tkestack.io/tke/api/client/clientset/versioned/typed/platform/v1"
	versionedinformers "tkestack.io/tke/api/client/informers/externalversions"
	generatedopenapi "tkestack.io/tke/api/openapi"
	"tkestack.io/tke/api/registry"
	"tkestack.io/tke/cmd/tke-chart-api/app/options"
	"tkestack.io/tke/pkg/apiserver/authentication"
	"tkestack.io/tke/pkg/apiserver/authorization"
	"tkestack.io/tke/pkg/apiserver/debug"
	"tkestack.io/tke/pkg/apiserver/filter"
	"tkestack.io/tke/pkg/apiserver/handler"
	"tkestack.io/tke/pkg/apiserver/openapi"
	"tkestack.io/tke/pkg/apiserver/storage"
	"tkestack.io/tke/pkg/apiserver/util"
	"tkestack.io/tke/pkg/chart/apiserver"
	"tkestack.io/tke/pkg/chart/chartmuseum"
	controllerconfig "tkestack.io/tke/pkg/controller/config"
	"tkestack.io/tke/pkg/registry/distribution"
)

const (
	license = "Apache 2.0"
	title   = "Tencent Kubernetes Engine Registry API"
)

// Config is the running configuration structure of the TKE registry apiserver.
type Config struct {
	ServerName                     string
	GenericAPIServerConfig         *genericapiserver.Config
	VersionedSharedInformerFactory versionedinformers.SharedInformerFactory
	StorageFactory                 *serverstorage.DefaultStorageFactory
	ExternalScheme                 string
	ExternalHost                   string
	ExternalPort                   int
	ExternalCAFile                 string
	OIDCCAFile                     string
	OIDCTokenReviewPath            string
	OIDCIssuerURL                  string
	AuthClient                     authversionedclient.AuthV1Interface
	BusinessClient                 businessversionedclient.BusinessV1Interface
	PlatformClient                 platformversionedclient.PlatformV1Interface
}

// CreateConfigFromOptions creates a running configuration instance based
// on a given TKE registry apiserver command line or configuration file option.
func CreateConfigFromOptions(serverName string, opts *options.Options) (*Config, error) {

	genericAPIServerConfig := genericapiserver.NewConfig(registry.Codecs)
	var ignoredAuthPathPrefixes []string
	ignoredAuthPathPrefixes = append(ignoredAuthPathPrefixes, distribution.IgnoredAuthPathPrefixes()...)
	ignoredAuthPathPrefixes = append(ignoredAuthPathPrefixes, chartmuseum.IgnoredAuthPathPrefixes()...)
	genericAPIServerConfig.BuildHandlerChainFunc = handler.BuildHandlerChain(ignoredAuthPathPrefixes, nil, nil)
	// long running function for distribution and chartmuseum path.
	genericAPIServerConfig.LongRunningFunc = filter.LongRunningRequestCheck(sets.NewString("watch"), sets.NewString(), ignoredAuthPathPrefixes)
	// increase default max post payload for distribution and chartmuseum.
	genericAPIServerConfig.MaxRequestBodyBytes = chartmuseum.MaxUploadSize
	genericAPIServerConfig.MergedResourceConfig = apiserver.DefaultAPIResourceConfigSource()
	genericAPIServerConfig.EnableIndex = false
	genericAPIServerConfig.EnableProfiling = false

	if err := util.SetupAuditConfig(genericAPIServerConfig, opts.Audit); err != nil {
		return nil, err
	}
	if err := opts.Generic.ApplyTo(genericAPIServerConfig); err != nil {
		return nil, err
	}
	if err := opts.SecureServing.ApplyTo(&genericAPIServerConfig.SecureServing, &genericAPIServerConfig.LoopbackClientConfig); err != nil {
		return nil, err
	}

	openapi.SetupOpenAPI(genericAPIServerConfig, generatedopenapi.GetOpenAPIDefinitions, title, license, opts.Generic.ExternalHost, opts.Generic.ExternalPort)

	// storageFactory
	storageFactoryConfig := storage.NewFactoryConfig(registry.Codecs, registry.Scheme)
	storageFactoryConfig.APIResourceConfig = genericAPIServerConfig.MergedResourceConfig
	completedStorageFactoryConfig, err := storageFactoryConfig.Complete(opts.ETCD)
	if err != nil {
		return nil, err
	}
	storageFactory, err := completedStorageFactoryConfig.New()
	if err != nil {
		return nil, err
	}
	if err := opts.ETCD.ApplyWithStorageFactoryTo(storageFactory, genericAPIServerConfig); err != nil {
		return nil, err
	}

	// client config
	genericAPIServerConfig.LoopbackClientConfig.ContentConfig.ContentType = "application/vnd.kubernetes.protobuf"

	kubeClientConfig := genericAPIServerConfig.LoopbackClientConfig
	clientgoExternalClient, err := versionedclientset.NewForConfig(kubeClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create real external clientset: %v", err)
	}
	versionedInformers := versionedinformers.NewSharedInformerFactory(clientgoExternalClient, 10*time.Minute)

	debug.SetupDebug(genericAPIServerConfig, opts.Debug)

	if err := authentication.SetupAuthentication(genericAPIServerConfig, opts.Authentication); err != nil {
		return nil, err
	}

	if err := authorization.SetupAuthorization(genericAPIServerConfig, opts.Authorization); err != nil {
		return nil, err
	}

	cfg := &Config{
		ServerName:                     serverName,
		GenericAPIServerConfig:         genericAPIServerConfig,
		VersionedSharedInformerFactory: versionedInformers,
		StorageFactory:                 storageFactory,
		ExternalScheme:                 opts.Generic.ExternalScheme,
		ExternalHost:                   opts.Generic.ExternalHost,
		ExternalPort:                   opts.Generic.ExternalPort,
		ExternalCAFile:                 opts.Generic.ExternalCAFile,
		OIDCIssuerURL:                  opts.Authentication.OIDC.IssuerURL,
		OIDCCAFile:                     opts.Authentication.OIDC.CAFile,
		OIDCTokenReviewPath:            opts.Authentication.OIDC.TokenReviewPath,
	}

	// client config for auth apiserver
	authAPIServerClientConfig, ok, err := controllerconfig.BuildClientConfig(opts.AuthAPIClient)
	if err != nil {
		return nil, err
	}
	if ok && authAPIServerClientConfig != nil {
		authClient, err := versionedclientset.NewForConfig(rest.AddUserAgent(authAPIServerClientConfig, "tke-chart-api"))
		if err != nil {
			return nil, err
		}
		cfg.AuthClient = authClient.AuthV1()
	}

	// client config for business apiserver
	businessAPIServerClientConfig, ok, err := controllerconfig.BuildClientConfig(opts.BusinessAPIClient)
	if err != nil {
		return nil, err
	}
	if ok && businessAPIServerClientConfig != nil {
		businessClient, err := versionedclientset.NewForConfig(rest.AddUserAgent(businessAPIServerClientConfig, "tke-chart-api"))
		if err != nil {
			return nil, err
		}
		cfg.BusinessClient = businessClient.BusinessV1()
	}

	// client config for platform apiserver
	platformAPIServerClientConfig, ok, err := controllerconfig.BuildClientConfig(opts.PlatformAPIClient)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("failed to initialize client config of platform API server")
	}
	platformClient, err := versionedclientset.NewForConfig(rest.AddUserAgent(platformAPIServerClientConfig, "tke-chart-api"))
	if err != nil {
		return nil, err
	}
	cfg.PlatformClient = platformClient.PlatformV1()

	return cfg, nil
}
