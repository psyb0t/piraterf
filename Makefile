# PIrateRF Project Makefile - Hack the fucking Planet! 🏴‍☠️
# Add your custom targets here - they will override servicepack defaults

# Override framework variables (optional)
MIN_TEST_COVERAGE := 70

# Include servicepack framework commands
include Makefile.servicepack

.PHONY: tls pi-setup-deps pi-setup-ap pi-setup-branding \
	pi-reboot ssh deploy uninstall install pi pi-image

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
