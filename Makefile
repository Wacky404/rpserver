CERT_DIR=certs
CERT_FILE=$(CERT_DIR)/localhost.pem
KEY_FILE=$(CERT_DIR)/localhost-key.pem
BINARY=bin/rpserver
SRC=cmd/rpserver/main.go

.PHONY: all cert build run clean ensure-deps

all: ensure-deps cert build run

# Ensure mkcert and nss are installed
ensure-deps:
	@echo "Checking mkcert and NSS dependencies..."
	@if ! command -v mkcert > /dev/null; then \
		echo "mkcert not found. Installing..."; \
		if command -v brew > /dev/null; then \
			brew install mkcert nss; \
		elif command -v apt > /dev/null; then \
			sudo apt update && sudo apt install -y mkcert libnss3-tools; \
		else \
			echo "Unsupported OS or missing package manager. Please install mkcert and nss manually."; exit 1; \
		fi \
	fi
	@mkcert -install

cert: ensure-deps
	mkdir -p $(CERT_DIR)
	mkcert -cert-file $(CERT_FILE) -key-file $(KEY_FILE) localhost

build: $(BINARY)
	chmod +x $(BINARY)

$(BINARY): $(SRC)
	go build -o $(BINARY) $(SRC)

run:
	$(BINARY) --cert $(CERT_FILE) --key $(KEY_FILE)

clean:
	rm -rf $(CERT_DIR) $(BINARY)
