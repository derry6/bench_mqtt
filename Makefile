
IMAGE_REPO = derrylu/bench_mqtt
IMAGE_TAG = v1.1
.PHONY:all
all:
	go build -o bench_mqtt

.PHONY:image
image:
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .

.PHONY:push 
push:
	docker push $(IMAGE_REPO):$(IMAGE_TAG)

.PHONY:clean
clean:
	rm -f bench_mqtt
	docker rmi -f $(IMAGE_REPO):$(IMAGE_TAG)