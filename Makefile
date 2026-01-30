PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin

BINARY = redway
GO = go
GOFLAGS = -ldflags="-s -w"

all: build

build:
	@echo "Building Redway..."
	$(GO) build $(GOFLAGS) -o $(BINARY) .
	@echo "Build complete: $(BINARY)"

static:
	@echo "Building static binary..."
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o $(BINARY) .
	@echo "Static build complete: $(BINARY)"

install: build
	@echo "Installing Redway..."
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo ""
	@echo "Installation complete!"
	@echo ""
	@echo "Usage:"
	@echo "  sudo redway init                                    # Initialize with default image"
	@echo "  sudo redway init docker://redroid/redroid:12.0.0_64only-latest  # Custom OCI image"
	@echo "  sudo redway start                                   # Start container"
	@echo "  sudo redway adb-connect                             # Get ADB info"

uninstall:
	@echo "Uninstalling Redway..."
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo "Note: Config and data preserved. Remove manually if needed:"
	@echo "  rm -rf ~/.config/redway"
	@echo "  rm -rf ~/data-redroid"

clean:
	@echo "Cleaning..."
	rm -f $(BINARY)

test:
	$(GO) test -v ./...

.PHONY: all build static install uninstall clean test
