terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "2.15.0"
    }
  }
}

provider "docker" {
  host    = "npipe:////.//pipe//docker_engine"
}

resource "docker_image" "frontend" {
  name = "plex-local-dl-ui"
  build {
    path = "../."
    tag = ["plex-local-dl-ui:local"]
  }
}

resource "docker_container" "frontend" {
  image = docker_image.frontend.latest
  name  = docker_image.frontend.name
  ports {
    internal = 8080
    external = 8080
  }
  ports {
    internal = 443
    external = 8443
  }
}