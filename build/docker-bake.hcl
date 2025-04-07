variable "IMAGE_TAG" {}
variable "CONTAINER_REGISTRY" {
  default = "dgpublicimagesprod.azurecr.io"
}

group "default" {
  targets = ["tftp", "http", "cleaner", "sync", "monitoring", "ipxeMenuGenerator"]
}

target "tftp" {
  tags = ["${CONTAINER_REGISTRY}/planetexpress/netboot-tftp:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/tftp"
}

target "http" {
  tags = ["${CONTAINER_REGISTRY}/planetexpress/netboot-http:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/http"
}

target "cleaner" {
  tags = ["${CONTAINER_REGISTRY}/planetexpress/netboot-cleaner:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/cleaner"
}

target "sync" {
  tags = ["${CONTAINER_REGISTRY}/planetexpress/netboot-sync:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/sync"
}

target "monitoring" {
  tags = ["${CONTAINER_REGISTRY}/planetexpress/netboot-monitoring:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/monitoring"
}

target "ipxeMenuGenerator" {
  tags = ["${CONTAINER_REGISTRY}/planetexpress/netboot-ipxe-menu-generator:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context = "./netboot-services/ipxeMenuGenerator"
}
