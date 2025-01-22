all: build

build:
	go run -race connectionHandler.go init.go hasher.go setter.go getter.go engine.go resizer.go ShardUtils.go notificationHandler.go

test:
	go test