.DEFAULT_GOAL := collector #shu

OUT = ./bin
CMD = ./cmd
PKG = .

VET_FLG = -composites=false

TEST_FLG = #-tags outtest

#############
# Collector #
#############
collector: provider collector-fmt collector-vet collector-test
	go build -o $(OUT)/collector/collector $(CMD)/collector/main.go 
.PHONY: collector

collector-fmt: 
	go fmt $(CMD)/collector/... $(PKG)/collector/...
.PHONY: collector-fmt

collector-vet: 
	go vet $(VET_FLG) $(CMD)/collector/... $(PKG)/collector/...
.PHONY: collector-vet

collector-test:
	go test $(TEST_FLG) $(CMD)/collector/... $(PKG)/collector/...
.PHONY: collector-test


#############
# Provider  #
#############
provider: provider-fmt provider-vet provider-test
.PHONY: provider

provider-fmt: 
	go fmt $(PKG)/provider/...
.PHONY: provider-fmt

provider-vet:
	go vet $(VET_FLG)  $(PKG)/provider/...
.PHONY: provider-vet

provider-test:
	go test $(TEST_FLG) $(PKG)/provider/...
.PHONY: provider-test

# provider-outtest:
# 	go test -tags outtest $(PKG)/provider/...
# .PHONY: provider-outtest

################
# seishub-util #
################
shu: seishub shu-fmt shu-vet shu-test
	go build -o $(OUT)/seishub-util/shu $(CMD)/seishub-util/main.go
.PHONY: shu

shu-fmt: 
	go fmt $(CMD)/seishub-util/...
.PHONY: shu-fmt

shu-vet: 
	go vet $(VET_FLG) $(CMD)/seishub-util/...
.PHONY: shu-vet

shu-test:
	go test $(TEST_FLG) $(CMD)/seishub-util/...
.PHONY: shu-test


###########
# seishub #
###########
seishub: seishub-fmt seishub-vet seishub-test
.PHONY: seishub

seishub-fmt:
	go fmt $(PKG)/provider/seishub/...
.PHONY: seishub-fmt

seishub-vet:
	go vet $(VET_FLG) $(PKG)/provider/seishub/...
.PHONY: seishub-vet

seishub-test:
	go test $(TEST_FLG) $(PKG)/provider/seishub/...
.PHONY: seishub-test