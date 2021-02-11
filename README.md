terraform-provider-kubernetes-awaiter
=====================================

Waits for kubernetes API resource described by uri

Usage
-----

Input data:

```hcl
locals {
  kube_config_path = "PATH TO KUBE CONFIG FILE"
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

Waiting to appear of Issuer and ClusterIssuer custom resource definition from `helm_release`:

```hcl
locals {
  server = yamldecode(file(local.kube_config_path)).clusters[0].cluster.server
  cacert = base64decode(yamldecode(file(local.kube_config_path)).clusters[0].cluster["certificate-authority-data"])
  token = lookup(data.kubernetes_secret.resource_checker.data, "token")
  
  resource_path = "/apis/apiextensions.k8s.io/v1/customresourcedefinitions/issuers.cert-manager.io" 
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

  uri = "${local.server}${resource_path}"
}
```

Known issues
------------

- `kubernetes_manifest` from 
  [terraform-provider-kubernetes-alpha](https://github.com/hashicorp/terraform-provider-kubernetes-alpha) is trying 
  to access CRD at the _planning_ phase, so `kubernetes-awaiter` can't helps here.

TODO
----
- [ ] add ValidateDiagFunc
- [ ] add tests
