KEYCLOCK_IMAGE=quay.io/keycloak/keycloak:18.0.2


sync-theme: cleanup
	$(eval UUID=$(shell docker run --entrypoint bash -itd --name keycloak_themes ${KEYCLOCK_IMAGE}))
	docker exec $(UUID) bash -c "mkdir -p /tmp/extract; cp -f /opt/keycloak/lib/lib/main/org.keycloak.keycloak-themes* /tmp/extract/keycloak-themes.jar"
	docker cp $(UUID):/tmp/extract/keycloak-themes.jar .
	tar xvf keycloak-themes.jar
	rm -rf META-INF
	rm -rf keycloak-themes.jar

cleanup:
	-docker rm -f keycloak_themes
	-rm -rf keycloak-themes.jar
	-rm -rf META-INF
