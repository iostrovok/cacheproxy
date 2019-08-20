CURDIR := $(shell pwd)
DIR:=FILE_DIR=$(CURDIR)/testfiles TEST_SOURCE_PATH=$(CURDIR)


##
## List of commands:
##

## default:
all: test

test:
	@echo "======================================================================"
	@echo "Run race test"
	@$(DIR) GODEBUG=gocacheverify=1 go test -cover -race ./...

test-init: clean test

clean:
	@echo "======================================================================"
	@echo "Clean old tests data..."
	@rm -f  ./cassettes/*

fmt:
	@echo "======================================================================"
	@echo "Run go fmt..."
	@go fmt ./

mod:
	@echo "======================================================================"
	@echo "Run MOD"
	@go mod verify
	@go mod tidy
	@go mod vendor
	@go mod download
	@go mod verify
