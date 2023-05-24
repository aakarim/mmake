# API service commands
show-vars:
	@echo "Root directory: $(MM_ROOT)"
	@echo "Current directory: $(MM_PATH)"
	@echo "Build directory: $(MM_OUT_ROOT)"
	@echo "Target directory: $(MM_OUT_PATH)"

build:
	cue export $(MM_PATH)/config.cue -o $(MM_OUT_PATH)/config.yaml

deploy: build
	@echo "Deploying api service with config..."
	@echo "Deployed api service!"
	cat $(MM_OUT_PATH)/config.yaml

.PHONY: show-vars build deploy

svc:
	go run ./services/api/cmd
