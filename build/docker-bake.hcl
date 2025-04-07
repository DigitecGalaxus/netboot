group "default" {
  targets = ["netboot", "tftp", "http", "cleaner", "sync", "monitoring", "ipxeMenuGenerator"]
}

target "tftp" {
  target     = "tftp"
  tags       = ["planetexpress/netboot-tftp:${IMAGE_TAG}"]
  dockerfile = "./netboot-services/tftp/Dockerfile"
  context    = "./netboot-services/tftp/"
}

target "http" {
  target     = "http"
  tags       = ["planetexpress/netboot-http:${IMAGE_TAG}"]
  dockerfile = "./netboot-services/http/Dockerfile"
  context    = "./netboot-services/http/"
}

target "cleaner" {
  target     = "cleaner"
  tags       = ["planetexpress/netboot-cleaner:${IMAGE_TAG}"]
  dockerfile = "./netboot-services/cleaner/Dockerfile"
  context    = "./netboot-services/cleaner/"
}

target "sync" {
  target     = "monitoring"
  tags       = ["planetexpress/netboot-sync:${IMAGE_TAG}"]
  dockerfile = "./netboot-services/sync/Dockerfile"
}

target "monitoring" {
  target     = "monitoring"
  tags       = ["planetexpress/netboot-monitoring:${IMAGE_TAG}"]
  dockerfile = "./netboot-services/monitoring/Dockerfile"
}

target "ipxeMenuGenerator" {
  target     = "ipxeMenuGenerator"
  tags       = ["planetexpress/netboot-ipxe-menu-generator:${IMAGE_TAG}"]
  dockerfile = "./netboot-services/ipxeMenuGenerator/Dockerfile"
}
