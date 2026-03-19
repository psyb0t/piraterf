# PIrateRF Project Makefile - Hack the fucking Planet! 🏴‍☠️
# Add your custom targets here - they will override servicepack defaults

# Override framework variables (optional)
MIN_TEST_COVERAGE := 70

# Include servicepack framework commands
include Makefile.servicepack

.PHONY: lint lint-fix tls pi-setup-deps pi-setup-ap pi-setup-branding \
	pi-reboot ssh deploy uninstall install pi pi-image

lint: ## Lint Go files
	@echo ""
	@echo "\033[1m=== Linting Go Files ===\033[0m"
	@echo ""
	@echo "\033[0;34m\033[1m[INFO]\033[0m Running modernize analysis..."
	@out=$$(go tool modernize -test ./... 2>&1 \
		| grep -v '\.gen\.go:') || true; \
	if [ -n "$$out" ]; then echo "$$out"; exit 1; fi
	@echo "\033[0;32m\033[1m[SUCCESS]\033[0m modernize passed!"
	@echo "\033[0;34m\033[1m[INFO]\033[0m Running golangci-lint..."
	@go tool golangci-lint run --timeout=30m0s ./...
	@echo "\033[0;32m\033[1m[SUCCESS]\033[0m Linting completed successfully!"

lint-fix: ## Lint and fix Go files
	@echo ""
	@echo "\033[1m=== Linting and Fixing Go Files ===\033[0m"
	@echo ""
	@echo "\033[0;34m\033[1m[INFO]\033[0m Running modernize analysis with fixes..."
	@gen_files=$$(find . -name '*.gen.go' -not -path './vendor/*'); \
	out=$$(go tool modernize -fix -test ./... 2>&1 \
		| grep -v '\.gen\.go:') || true; \
	if [ -n "$$gen_files" ]; then echo "$$gen_files" | xargs git checkout -- 2>/dev/null || true; fi; \
	if [ -n "$$out" ]; then echo "$$out"; exit 1; fi
	@echo "\033[0;32m\033[1m[SUCCESS]\033[0m modernize passed!"
	@echo "\033[0;34m\033[1m[INFO]\033[0m Running golangci-lint with fixes..."
	@go tool golangci-lint run --fix --timeout=30m0s ./...
	@echo "\033[0;32m\033[1m[SUCCESS]\033[0m Linting and fixing completed successfully!"

build: ## Build the fucking PIrateRF beast
	@./$(SCRIPTS_DIR)/make/build.sh

tls: ## Generate fucking TLS certificates for HTTPS
	@./$(SCRIPTS_DIR)/make/tls.sh

pi-setup-deps: ## Setup the fucking dependencies on Pi
	@./$(SCRIPTS_DIR)/make/pi_setup_deps.sh

pi-setup-ap: ## Setup the fucking wifi AP on the bastard PI
	@./$(SCRIPTS_DIR)/make/pi_setup_ap.sh

pi-setup-branding: ## Setup the fucking system branding on the Pi
	@./$(SCRIPTS_DIR)/make/pi_setup_branding.sh

pi-reboot: ## Reboot the fucking Pi
	@./$(SCRIPTS_DIR)/make/pi_reboot.sh

ssh: ## SSH into the fucking Pi
	@./$(SCRIPTS_DIR)/make/ssh.sh


deploy: build ## Deploy the fucking PIrateRF files to the bastard Pi
	@./$(SCRIPTS_DIR)/make/deploy.sh

run-dev: ## Run PIrateRF in fucking development mode
	@./$(SCRIPTS_DIR)/make/run_dev.sh

uninstall: ## Remove the fucking PIrateRF shit from the Pi completely
	@./$(SCRIPTS_DIR)/make/uninstall.sh

install: ## Install the fucking PIrateRF to the bastard Pi
	@./$(SCRIPTS_DIR)/make/install.sh

pi: ## Complete setup of the fkin' Pi
	@./$(SCRIPTS_DIR)/make/pi.sh

pi-image: ## Clone and shrink the fucking Pi SD card images for chaos distribution
	@./$(SCRIPTS_DIR)/make/pi_image.sh
