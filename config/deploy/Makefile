








DEPLOYS=`tke-auth-api tke-auth-controller tke-registry-api tke-platform-api tke-platform-controller tke-business-api tke-business-controller tke-notify-api tke-notify-controller tke-monitor-api tke-monitor-controller tke-audit-api tke-gateway`

COMPOMENT=tke-chart-controller
GOARCH=amd64
REGISTRY=registry.cn-qingdao.aliyuncs.com/zhangkai8048
TAG=a6f3067d.dirty

export OPT=delete
export IMAGE=$(REGISTRY)/$(COMPOMENT)-$(GOARCH):$(TAG)




test: 
	go test -timeout 30s -run ^TestManifest$$ $$(pwd)/manifests/$(COMPOMENT) -v

testapply:
	for i in $$( ls manifests |grep tke);do \
	echo $$i && \
	go test -timeout 30s  $$(pwd)/manifests/$$i -v ; \
	done
	
push:
	docker tag tkestack/tke-chart-api-amd64:$(TAG) registry.cn-qingdao.aliyuncs.com/zhangkai8048/tke-chart-api-amd64:$(TAG)
	docker push registry.cn-qingdao.aliyuncs.com/zhangkai8048/tke-chart-api-amd64:$(TAG) 
	docker tag tkestack/tke-chart-controller-amd64:$(TAG) registry.cn-qingdao.aliyuncs.com/zhangkai8048/tke-chart-controller-amd64:$(TAG) 
	docker push registry.cn-qingdao.aliyuncs.com/zhangkai8048/tke-chart-controller-amd64:$(TAG)
	
load:
	for i in $$(docker images |grep chart | awk '{print $$1":"$$2}');do  kind load docker-image $$i --name kind; done