# Kubernetes Admission Webhook

Admission webhooks are HTTP callbacks that receive admission requests and do something with them.
You can define two types of admission webhooks, validating admission Webhook and mutating admission webhook.
With validating admission Webhooks, you may reject requests to enforce custom admission policies.
With mutating admission Webhooks, you may change requests to enforce custom defaults.

This repo currently aims to provide an example for a validating admission Webhook.

## Overview

The Admission webhook requires a `ValidatingWebhookConfiguration` to be created. Once created the Kubernetes API server will send requests to the Webhook based upon the configuration created. 
The configuration specifies a `namespace` & `service` to call back to which will process the webhook, then send a `Allowed` or `Disallowed` to the server. 
In the event the webhook is disallowed, a `Status` response will be added to the request so it is clear why the request was denied.

The example code will allow any pod to be created except if one of the containers is named `steve`. 

## Generate certs

The webhook requires the service to be running TLS.
The following example will show a simple way to generate a self-signed cert. 

Generate CA
```
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -days 100000 -out ca.crt -subj "/CN=admission_ca"
```
Generate server
```
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=webhook.default.svc" -config server.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 100000 -extensions v3_req -extfile server.conf
```

## Deploy

### Certs

Create secret to pass the certs created in previous step:

```bash
$ kubectl create secret generic webhookcerts --from-file=server.crt --from-file=server.key
```

### Webhook

Deploy the webhook deployment, service, and config:

```bash
$ kubectl apply -f deployment
```