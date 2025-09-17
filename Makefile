# PIrateRF Project Makefile - Hack the fucking Planet! 🏴‍☠️
# Add your custom targets here - they will override servicepack defaults

# Override framework variables (optional)
# MIN_TEST_COVERAGE := 95

# Include servicepack framework commands
include Makefile.servicepack

.PHONY: tls pi-setup-deps pi-setup-ap pi-setup-branding \
	pi-reboot ssh pi-setup deploy uninstall install complete

build: ## Build the fucking PIrateRF beast
	@./$(SCRIPTS_DIR)/make/build.sh

test-coverage: ## Run the fucking tests with coverage
	ENV=dev $(MAKE) -f Makefile.servicepack test-coverage

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

pi-setup: pi-setup-deps pi-setup-ap pi-setup-branding pi-reboot ## Setup the bastard Pi for the fkin' mission
	@./$(SCRIPTS_DIR)/make/pi_setup.sh

deploy: build ## Deploy the fucking PIrateRF files to the bastard Pi
	@./$(SCRIPTS_DIR)/make/deploy.sh

run-dev: ## Run PIrateRF in fucking development mode
	@./$(SCRIPTS_DIR)/make/run_dev.sh

uninstall: ## Remove the fucking PIrateRF shit from the Pi completely
	@./$(SCRIPTS_DIR)/make/uninstall.sh

install: ## Install the fucking PIrateRF to the bastard Pi
	@./$(SCRIPTS_DIR)/make/install.sh

complete: pi-setup-deps pi-setup-ap pi-setup-branding deploy install pi-reboot ## Complete setup of the fkin' Pi
	@./$(SCRIPTS_DIR)/make/complete.sh
