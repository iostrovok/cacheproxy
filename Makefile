CURDIR := $(shell pwd)
DIR:=TEST_SOURCE_PATH=$(CURDIR)


##
## List of commands:
##

## default:
all: test

test-travis:
	 go test  -race ./...

test: clean
	@echo "======================================================================"
	@echo "Run race test for ./"
	cd $(CURDIR)/ && go test -coverprofile=$(CURDIR)/coverage.main.out -race ./...
	go tool cover -html=$(CURDIR)/coverage.main.out -o $(CURDIR)/coverage.main.html
	@rm -f $(CURDIR)/coverage.main.out
	@rm -f  ./cassettes/*

clean:
	@echo "======================================================================"
	@echo "Clean old tests data..."
	@rm -f  ./coverage.*
	@rm -f  ./cassettes/*

fmt:
	@echo "======================================================================"
	@echo "Run go fmt..."
	@go fmt ./

mod:
	@echo "======================================================================"
	@echo "Run MOD"
	@go mod tidy
	@go mod vendor
	@go mod download
	@go mod verify
