terraform-provider-kubernetes-awaiter
=====================================

Waits until resource appearing in the K8S cluster. Resource should be described with URI.

![8006789bb0cc3734ca56a33a79d2660023d66fd71ea1755948161b32292801bf](https://user-images.githubusercontent.com/418868/107657861-a095d400-6c96-11eb-8b79-df7e07c84f8e.jpg)

Provider
--------

https://registry.terraform.io/providers/a0s/kubernetes-awaiter

```hcl
terraform {
  required_providers {
    kubernetes-awaiter = {
      version = "~> v0.1.0"
      source = "a0s/kubernetes-awaiter"
    }
  }
}

provider "kubernetes-awaiter" {}

resource "kubernetes_resource_awaiter" "waiter" {
  provider = kubernetes-awaiter
  
  cacert = "CACERT"
  token = "TOKEN"
  timeout = "5m"
  poll = "1s"

  uri = "RESOURCE_URI"
}
```

Usage
-----

Input data:

```hcl
locals {
  // Path to the kube config file
  kube_config_path = "../kube-conig.yml"
  
  // URI path to the Kubernetes API resource which we want to wait
  kubeapi_resource_path = "/apis/apiextensions.k8s.io/v1/customresourcedefinitions/issuers.cert-manager.io"
}
```

Create service account (with token) which able to check resource existing:

```hcl
resource "kubernetes_service_account" "resource_checker" {
  metadata {
    name = "terraform-resource-checker"
  }
}

data "kubernetes_secret" "resource_checker" {
  metadata {
    name = kubernetes_service_account.resource_checker.default_secret_name
  }
}

resource "kubernetes_cluster_role" "resource_checker" {
  metadata {
    name = "terraform-resource-checker"
  }
  rule {
    api_groups = ["apiextensions.k8s.io"]
    resources = ["customresourcedefinitions"]
    verbs = ["get"]
  }
}

resource "kubernetes_cluster_role_binding" "resource_checker" {
  metadata {
    name = "terraform-resource-checker"
  }
  subject {
    kind = "ServiceAccount"
    name = kubernetes_service_account.resource_checker.metadata[0].name
    api_group = ""
  }
  role_ref {
    kind = "ClusterRole"
    name = kubernetes_cluster_role.resource_checker.metadata[0].name
    api_group = "rbac.authorization.k8s.io"
  }
}
```

Waiting to appear of Issuer custom resource definition from `helm_release`:

```hcl
locals {
  server = yamldecode(file(local.kube_config_path)).clusters[0].cluster.server
  cacert = base64decode(yamldecode(file(local.kube_config_path)).clusters[0].cluster["certificate-authority-data"])
  token = lookup(data.kubernetes_secret.resource_checker.data, "token")
}

resource "helm_release" "cert_manager" {
  chart = "cert-manager"
  name = "cert-manager"
  repository = "https://charts.jetstack.io"
  namespace = "cert-manager"
  create_namespace = true

  set {
    name = "installCRDs"
    value = true
  }
}

resource "kubernetes_resource_awaiter" "wait_cert_manager" {
  provider = kubernetes-awaiter
  depends_on = [helm_release.cert_manager]

  cacert = local.cacert
  token = local.token
  timeout = "5m"
  poll = "1s"

  uri = "${local.server}${local.kubeapi_resource_path}"
}
```

Known issues
------------

- ~~`kubernetes_manifest` from [terraform-provider-kubernetes-alpha](https://github.com/hashicorp/terraform-provider-kubernetes-alpha) try to access CRD at the _planning_ phase, so `kubernetes-awaiter` can't helps here.~~ fixed in [v0.3.2](https://github.com/hashicorp/terraform-provider-kubernetes-alpha/releases/tag/v0.3.2)

TODO
----
- [ ] add ValidateDiagFunc
- [ ] add tests
