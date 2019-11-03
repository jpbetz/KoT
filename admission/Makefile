SHELL = bash

OUTPUT := manifests.yaml

all: tls.key tls.crt $(OUTPUT)

tls.key tls.crt:
		openssl req -new -x509 -subj "/CN=admission-webhook.deepsea.svc" -nodes -newkey rsa:4096 -keyout tls.key -out tls.crt -days 365

%.yaml: tls.key tls.crt
		sed 's,LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tC.*$$,$(shell base64 tls.crt),;s,LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tC.*$$,$(shell base64 tls.key),' $@ > $@.updated
		mv $@.updated $@
