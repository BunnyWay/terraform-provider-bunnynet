build:
	@go build -o terraform-provider-bunnynet

.PHONY: docs
docs:
	@go generate

lint:
	@golangci-lint run ./...

test: build unit acc

acc:
	@TESTACC_MC_REGION=DE TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

unit:
	@go list ./... | egrep -v '/internal/provider$$' | xargs go test -v
