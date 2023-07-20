








DEPLOYS=`
tke-auth-api tke-auth-controller tke-registry-api tke-platform-api tke-platform-controller tke-business-api tke-business-controller tke-notify-api tke-notify-controller tke-monitor-api tke-monitor-controller tke-audit-api tke-gateway`

COMPOMENT =tke-application-controller
GOARCH=amd64
REGISTRY=tkestack
TAG=v1.9.2.1.g30e412e9.dirty

export OPT=delete
export IMAGE=$(REGISTRY)/$(COMPOMENT)-$(GOARCH):$(TAG)




test: 
	go test -timeout 30s -run ^TestManifest$$ $$(pwd)/manifests/$(COMPOMENT) -v

testapply:
	for i in $$( ls manifests |grep tke);do \
	echo $$i && \
	go test -timeout 30s  $$(pwd)/manifests/$$i -v ; \
	done
	
	
load:
	for i in $$(docker images |grep tke | awk '{print $$1":"$$2}');do  kind load docker-image $$i --name kind; done