KEYCLOCK_IMAGE_NAME := quay.io/keycloak/keycloak
KEYCLOCK_IMAGE_TAG := $(shell cat KEYCLOAK_VERSION)
KEYCLOCK_IMAGE := $(KEYCLOCK_IMAGE_NAME):$(KEYCLOCK_IMAGE_TAG)
REGISTRY := ghcr.io/pdexchange
IMG := keycloak
VERSION ?= dev

.PHONY: sync-theme
sync-theme: clean
	$(eval UUID=$(shell docker run --entrypoint bash -itd --name keycloak_themes ${KEYCLOCK_IMAGE}))
	docker exec $(UUID) bash -c "mkdir -p /tmp/extract; cp -f /opt/keycloak/lib/lib/main/org.keycloak.keycloak-themes* /tmp/extract/keycloak-themes.jar"
	docker cp $(UUID):/tmp/extract/keycloak-themes.jar .
	tar xvf keycloak-themes.jar
	rm -rf META-INF
	rm -rf keycloak-themes.jar
	cp -r pac/ theme/pac

.PHONY: clean
clean:
	-docker rm -f keycloak_themes
	-rm -rf keycloak-themes.jar
	-rm -rf META-INF

.PHONY: docker-build
docker-build:
	docker build --build-arg KEYCLOAK_BASE_IMAGE=$(KEYCLOCK_IMAGE) -t $(REGISTRY)/$(IMG):$(VERSION)-$(KEYCLOCK_IMAGE_TAG) .

.PHONY: build-prod-image
build-prod-image: clean sync-theme
	docker build --build-arg KEYCLOAK_BASE_IMAGE=$(KEYCLOCK_IMAGE) -t $(REGISTRY)/$(IMG):$(VERSION)-$(KEYCLOCK_IMAGE_TAG) .

.PHONY: clean
clean:
	docker rm -f keycloak_themes
	rm -rf theme