.PHONY: all build test tidy clean

GO := go

default: build

build:
	$(GO) build -o mysql-tester ./src

debug:
	$(GO) build -gcflags="all=-N -l" -o mysql-tester ./src

test: build
	$(GO) test -cover ./...
	#./mysql-tester -check-error

tidy:
	$(GO) mod tidy

clean:
	$(GO) clean -i ./...
	rm -rf mysql-tester
