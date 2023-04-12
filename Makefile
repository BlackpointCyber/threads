GOBIN=$(shell go env GOPATH)/bin

test: setup
	$(GOBIN)/richgo test -timeout 10s $(args) ./...

lint: setup
	go vet ./...
	$(GOBIN)/staticcheck ./...

# Update adapters to use a new ksql tag
version=
update:
	git tag $(version)
	git push origin $(version)

setup: $(GOBIN)/richgo $(GOBIN)/staticcheck
$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
$(GOBIN)/richgo:
	go install github.com/kyoh86/richgo@latest
