variable "IMAGE_TAG" {}
variable "CONTAINER_REGISTRY" {
  default = "dgpublicimagesprod.azurecr.io"
}

group "default" {
  targets = ["tftp", "http", "cleaner", "sync", "monitoring", "ipxeMenuGenerator"]
}

target "tftp" {
  tags = ["planetexpress/netboot-tftp:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/tftp"
}

target "http" {
  tags = ["planetexpress/netboot-http:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/http"
}

target "cleaner" {
  tags = ["planetexpress/netboot-cleaner:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/cleaner"
}

target "sync" {
  tags = ["planetexpress/netboot-sync:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/sync"
}

target "monitoring" {
  tags = ["planetexpress/netboot-monitoring:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/monitoring"
}

target "ipxeMenuGenerator" {
  tags = ["planetexpress/netboot-ipxe-menu-generator:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/ipxeMenuGenerator"
}
