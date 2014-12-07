run:
	go run local_server.go

build:
	go build local_server.go

test:
	go test local_server

install: build
	sudo install local_server /usr/local/bin
	sudo local_server -install
