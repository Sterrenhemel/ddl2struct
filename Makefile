GO_FILES=$(shell find . -name "*.go")
GO_MOD=$(shell go list)

fmt:
	@gofmt -s -l -w $(GO_FILES)
	@goimports -l -w -local $(GO_MOD) $(GO_FILES)