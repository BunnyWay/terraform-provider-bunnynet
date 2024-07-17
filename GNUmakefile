default: testacc

# Run acceptance tests
.PHONY: testacc
testacc: docs
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

build:
	@go build -o terraform-provider-bunnynet

.PHONY: docs
docs:
	@go generate
