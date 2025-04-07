variable "IMAGE_TAG" {}
variable "CONTAINER_REGISTRY" {
  default = "dgpublicimagesprod.azurecr.io"
}

group "default" {
  targets = ["tftp", "http", "cleaner", "sync", "monitoring", "ipxeMenuGenerator"]
}

target "tftp" {
  tags       = ["${CONTAINER_REGISTRY}/planetexpress/netboot-tftp:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context    = "./netboot-services/tftp"
  output     = ["type=registry"]
}

target "http" {
  tags       = ["${CONTAINER_REGISTRY}/planetexpress/netboot-http:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context    = "./netboot-services/http"
  output     = ["type=registry"]
}

target "cleaner" {
  tags       = ["${CONTAINER_REGISTRY}/planetexpress/netboot-cleaner:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context    = "./netboot-services/cleaner"
  output     = ["type=registry"]
}

target "sync" {
  tags       = ["${CONTAINER_REGISTRY}/planetexpress/netboot-sync:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context    = "./netboot-services/sync"
  output     = ["type=registry"]
}

target "monitoring" {
  tags       = ["${CONTAINER_REGISTRY}/planetexpress/netboot-monitoring:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context    = "./netboot-services/monitoring"
  output     = ["type=registry"]
}

target "ipxeMenuGenerator" {
  tags       = ["${CONTAINER_REGISTRY}/planetexpress/netboot-ipxe-menu-generator:${IMAGE_TAG}"]
  dockerfile = "Dockerfile"
  context    = "./netboot-services/ipxeMenuGenerator"
  output     = ["type=registry"]
}
