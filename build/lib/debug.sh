#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


# cd tke
if [! -d "_debug/certificates"];then
  mkdir -p _debug/certificates
fi

if [! -f "_debug/certificates/ocalhost+2.pem"];then 
  cd _debug/certificates
  mkcert -install
  mkcert localhost 127.0.0.1 ::1
  cd ../../
fi

if [! -f "_debug/token.csv"];then 
  touch _debug/token.csv
  echo 'token,admin,1,"administrator"' > _debug/token.csv
fi

# 
export root_store=~/.local/share

cat > _debug/auth-api.json<< EOF
{
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "etcd": {
      "servers": ["http://127.0.0.1:2379"]
    },
    "authentication": {
      "token_auth_file": "_debug/token.csv",
      "privileged_username": "admin"
    },
    "generic": {
      "external_hostname": "localhost",
      "external_port": 9451
    },
    "auth": {
      "assets_path": "./pkg/auth/web",
      "init_client_id": "client",
      "init_client_secret": "secret",
      "init_client_redirect_uris": [
        "http://localhost:9442/callback",
        "http://127.0.0.1:9442/callback",
        "https://localhost:9441/callback",
        "https://127.0.0.1:9441/callback"
      ]
    }
  }
EOF



### tke-auth-controller
cat > _debug/auth-api-client-config.yaml<< EOF
    apiVersion: v1
    kind: Config
    clusters:
      - name: tke
        cluster:
          certificate-authority: ${root_store}/mkcert/rootCA.pem
          server: https://127.0.0.1:9451
    users:
      - name: admin
        user:
          token: token
    current-context: tke
    contexts:
      - context:
          cluster: tke
          user: admin
        name: tke
EOF



cat >_debug/auth-controller.json<< EOF
    {
      "secure_serving": {
        "tls_cert_file": "_debug/certificates/localhost+2.pem",
        "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
      },
      "client": {
        "auth": {
          "api_server_client_config": "_debug/auth-api-client-config.yaml"
        }
      },
      "features":{
        "category_path": "hack/auth/category.json",
        "policy_path": "hack/auth/policy.json",
        "tenant_admin": "admin",
        "tenant_admin_secret": "secret"
        }
    }
EOF

### tke-platform-api
cat >_debug/platform-api.json<< EOF
  {
    "authentication": {
      "oidc": {
        "client_id": "client",
        "issuer_url": "https://localhost:9451/oidc",
        "ca_file": "${root_store}/mkcert/rootCA.pem",
        "username_prefix": "-",
        "username_claim": "name",
        "tenantid_claim": "federated_claims"
      },
      "token_auth_file": "_debug/token.csv"
    },
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "etcd": {
      "servers": ["http://127.0.0.1:2379"]
    }
  }
EOF
### tke-platform-controller
cat > _debug/platform-api-client-config.yaml<< EOF
  apiVersion: v1
  kind: Config
  clusters:
    - name: tke
      cluster:
        certificate-authority: ${root_store}/mkcert/rootCA.pem
        server: https://127.0.0.1:9443
  users:
    - name: admin
      user:
        token: token
  current-context: tke
  contexts:
    - context:
        cluster: tke
        user: admin
      name: tke
EOF


cat > _debug/platform-controller.json<< EOF

  {
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "client": {
      "platform": {
        "api_server_client_config": "_debug/platform-api-client-config.yaml"
      }
    }
  }
EOF



### tke-registry-api(Optional)
cat > _debug/registry-api.json<< EOF
  {
    "authentication": {
      "oidc": {
        "client_id": "client",
        "issuer_url": "https://localhost:9451/oidc",
        "ca_file": "${root_store}/mkcert/rootCA.pem",
        "token_review_path": "/auth/authn",
        "username_prefix": "-",
        "username_claim": "name",
        "tenantid_claim": "federated_claims"
      },
      "requestheader": {
        "username_headers": "X-Remote-User",
        "group_headers": "X-Remote-Groups",
        "extra_headers_prefix": "X-Remote-Extra-",
        "client_ca_file": "${root_store}/mkcert/rootCA.pem"
      },
      "token_auth_file": "_debug/token.csv"
    },
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "etcd": {
      "servers": [
        "http://127.0.0.1:2379"
      ]
    },
    "registry_config": "_debug/registry-config.yaml"
  }
EOF

cat > _debug/registry-config.yaml<< EOF
  apiVersion: registry.config.tkestack.io/v1
  kind: RegistryConfiguration
  storage:
    fileSystem:
      rootDirectory: _debug/registry
  security:
    # private key for signing registry JWT token, PKCS#1 encoded.
    tokenPrivateKeyFile: keys/private_key.pem
    tokenPublicKeyFile: keys/public.crt
    adminPassword: secret
    adminUsername: admin
    httpSecret: secret
  defaultTenant: default
EOF


### tke-business-api(Optional)

cat > _debug/business-api.json<< EOF
  {
    "authentication": {
      "oidc": {
        "client_id": "client",
        "issuer_url": "https://localhost:9451/oidc",
        "ca_file": "${root_store}/mkcert/rootCA.pem",
        "username_prefix": "-",
        "username_claim": "name",
        "tenantid_claim": "federated_claims"
      },
      "token_auth_file": "_debug/token.csv"
    },
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "etcd": {
      "servers": ["http://127.0.0.1:2379"]
    },
    "client": {
      "platform": {
        "api_server_client_config": "_debug/platform-api-client-config.yaml"
      }
    }
  }
EOF

cat > _debug/business-api-client-config.yaml<< EOF
  apiVersion: v1
  kind: Config
  clusters:
    - name: tke
      cluster:
        certificate-authority: ${root_store}/mkcert/rootCA.pem
        server: https://127.0.0.1:9447
  users:
    - name: admin
      user:
        token: token
  current-context: tke
  contexts:
    - context:
        cluster: tke
        user: admin
      name: tke
EOF

cat >_debug/business-controller.json<< EOF
  {
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "client": {
      "platform": {
        "api_server_client_config": "_debug/platform-api-client-config.yaml"
      },
      "business": {
        "api_server_client_config": "_debug/business-api-client-config.yaml"
      }
    }
  }
EOF


### tke-monitor-api(Optional)


cat > _debug/monitor-config.yaml<< EOF
  apiVersion: monitor.config.tkestack.io/v1
  kind: MonitorConfiguration
  storage:
    influxDB:
      servers:
        - address: http://localhost:8086
EOF


cat > _debug/monitor-api-client-config.yaml<< EOF
  apiVersion: v1
  kind: Config
  clusters:
    - name: tke
      cluster:
        certificate-authority: ${root_store}/mkcert/rootCA.pem
        server: https://127.0.0.1:9455
  users:
    - name: admin
      user:
        token: token
  current-context: tke
  contexts:
    - context:
        cluster: tke
        user: admin
      name: tke
EOF


cat > _debug/monitor-api.json<< EOF
  {
    "authentication": {
      "oidc": {
        "client_id": "client",
        "issuer_url": "https://localhost:9451/oidc",
        "ca_file": "${root_store}/mkcert/rootCA.pem",
        "username_prefix": "-",
        "username_claim": "name",
        "tenantid_claim": "federated_claims"
      },
      "token_auth_file": "_debug/token.csv"
    },
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "etcd": {
      "servers": ["http://127.0.0.1:2379"]
    },
    "client": {
      "platform": {
        "api_server_client_config": "_debug/platform-api-client-config.yaml"
      }
    },
    "monitor_config": "_debug/monitor-config.yaml"
  }

EOF

### tke-monitor-controller(Optional)

cat > _debug/monitor-controller.json<< EOF
  {
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "client": {
      "monitor": {
        "api_server_client_config": "_debug/monitor-api-client-config.yaml"
      },
      "business": {
        "api_server_client_config": "_debug/business-api-client-config.yaml"
      }
    },
    "monitor_config": "_debug/monitor-config.yaml"
  }

EOF



### tke-notify-api(Optional)

cat > _debug/notify-api.json<< EOF
  {
    "authentication": {
      "oidc": {
        "client_id": "client",
        "issuer_url": "https://localhost:9451/oidc",
        "ca_file": "${root_store}/mkcert/rootCA.pem",
        "username_prefix": "-",
        "username_claim": "name",
        "tenantid_claim": "federated_claims"
      },
      "requestheader": {
        "username_headers": "X-Remote-User",
        "group_headers": "X-Remote-Groups",
        "extra_headers_prefix": "X-Remote-Extra-",
        "client_ca_file": "${root_store}/mkcert/rootCA.pem"
      },
      "token_auth_file": "_debug/token.csv"
    },
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "etcd": {
      "servers": ["http://127.0.0.1:2379"]
    },
    "client": {
      "platform": {
        "api_server_client_config": "_debug/platform-api-client-config.yaml"
      }
    }
  }
EOF


### tke-notify-controller(Optional)
cat > _debug/notify-api-client-config.yaml<< EOF
  apiVersion: v1
  kind: Config
  clusters:
    - name: tke
      cluster:
        certificate-authority: ${root_store}/mkcert/rootCA.pem
        server: https://127.0.0.1:9457
  users:
    - name: admin
      user:
        token: token
  current-context: tke
  contexts:
    - context:
        cluster: tke
        user: admin
      name: tke
EOF

cat > _debug/notify-controller.json<< EOF
  {
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "client": {
      "notify": {
        "api_server_client_config": "_debug/notify-api-client-config.yaml"
      }
    }
  }
EOF




### tke-gateway

cat > _debug/gateway-config.yaml<< EOF
  apiVersion: gateway.config.tkestack.io/v1
  kind: GatewayConfiguration
  components:
    auth:
      address: https://127.0.0.1:9451
      passthrough:
        caFile: ${root_store}/mkcert/rootCA.pem
    platform:
      address: https://127.0.0.1:9443
      passthrough:
        caFile: ${root_store}/mkcert/rootCA.pem
    ### Optional Services ###
    # TKE Registry
    # registry:
    #   address: https://127.0.0.1:9453
    #   passthrough:
    #     caFile: ${root_store}/mkcert/rootCA.pem
    # TKE Business
    # business:
    #   address: https://127.0.0.1:9447
    #   frontProxy:
    #     caFile: ${root_store}/mkcert/rootCA.pem
    #     clientCertFile: certificates/localhost+2-client.pem
    #     clientKeyFile: certificates/localhost+2-client-key.pem
    # TKE Monitor
    # monitor:
    #   address: https://127.0.0.1:9455
    #   passthrough:
    #     caFile: ${root_store}/mkcert/rootCA.pem
    # TKE Notify
    # notify:
    #   address: https://127.0.0.1:9457
    #   passthrough:
    #         caFile: ${root_store}/mkcert/rootCA.pem

EOF


cat > _debug/gateway.json << EOF
{
    "authentication": {
      "oidc": {
        "client_secret": "secret",
        "client_id": "client",
        "issuer_url": "https://localhost:9451/oidc",
        "ca_file": "${root_store}/mkcert/rootCA.pem",
        "username_prefix": "-",
        "username_claim": "name",
        "tenantid_claim": "federated_claims"
      }
    },
    "secure_serving": {
      "tls_cert_file": "_debug/certificates/localhost+2.pem",
      "tls_private_key_file": "_debug/certificates/localhost+2-key.pem"
    },
    "gateway_config": "_debug/gateway-config.yaml"
}
EOF


