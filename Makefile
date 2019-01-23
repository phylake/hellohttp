default: build
	docker push phylake/hellohttp

build:
	gofmt -s -w .
	docker build -t phylake/hellohttp .

run: build
	docker run -it --rm -p 3000:3000 phylake/hellohttp
