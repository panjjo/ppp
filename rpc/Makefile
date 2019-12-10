REGISTRY:=harbor.yunss.com:5000
GROUP:=micro
PROJECT:=ppp
TAG:=k8s
NAMESPACE:=micro

all: build docker push

build:
	GOOS=linux go build -v -o srv

docker:
	docker image rm -f ${GROUP}/${PROJECT}:${TAG}
	docker build -t ${GROUP}/${PROJECT}:${TAG} .

tag:
	docker image rm -f ${GROUP}/${PROJECT}:${TAG}
	docker tag ${GROUP}/${PROJECT}:latest ${GROUP}/${PROJECT}:${TAG}

push:
	docker image rm -f ${REGISTRY}/${GROUP}/${PROJECT}:${TAG}
	docker tag ${GROUP}/${PROJECT}:${TAG} ${REGISTRY}/${GROUP}/${PROJECT}:${TAG}
	docker push ${REGISTRY}/${GROUP}/${PROJECT}:${TAG}

redeploy:
	kubectl delete -n ${NAMESPACE} -f k8s/deploy.yaml
	kubectl apply -n ${NAMESPACE} -f k8s/deploy.yaml

deploy:
	kubectl apply -n ${NAMESPACE} -f k8s/deploy.yaml
