.DEFAULT_GOAL := collector #shu-build

collector:
	go build -o bin/collector/collector cmd/collector/main.go
.PHONY: collector

#seishub-util
shu-build: #sh-fmt sh-vet
	go build -o bin/seishub-util/shu cmd/seishub-util/main.go
.PHONY: shu-build

sh-fmt:
	go fmt ./pkg/seishub
.PHONY: sh-fmt

sh-vet:
	go vet ./pkg/seishub
