.PHONY: init dev

all: init dev

.PHONY: init dev

all: init dev

# Initialize dependencies
init:
	go mod tidy
	go mod download

# Run the application in development mode
dev:
	go run ./cmd/timesheet $(ARGS)

# Cleanup
clean:
	go clean
	rm -f coverage.out

# Build the application
build:
	go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o timesheet ./cmd/timesheet

# Run tests
test:
	go test ./... -cover

# Install dependencies and run the application
run: init dev

