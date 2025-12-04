.PHONY: secrets-tools scan-secrets secrets-audit secrets-tools-clean

VENV := .venv
VENV_BIN := $(VENV)/bin
ACTIVATE := $(VENV_BIN)/activate

PIP := $(VENV_BIN)/pip
PYTHON := python3

DETECT_SECRETS_GIT ?= https://github.com/ibm/detect-secrets.git@master#egg=detect-secrets

SECRETS_BASELINE := .secrets.baseline

# Create virtual environment if missing
$(ACTIVATE):
	$(PYTHON) -m venv $(VENV)
	$(PIP) install --upgrade pip

# Install detect-secrets into venv
secrets-tools: $(ACTIVATE)
	$(PIP) install "git+$(DETECT_SECRETS_GIT)"

#scan for any potential secrets exposure in the repo.
scan-secrets: secrets-tools
	$(VENV_BIN)/detect-secrets scan --update $(SECRETS_BASELINE)

# Audit the scanned secrets file to resolve accordingly.
secrets-audit: secrets-tools
	$(VENV_BIN)/detect-secrets audit $(SECRETS_BASELINE)
