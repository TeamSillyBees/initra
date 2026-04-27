variable "local_url" {
  type    = string
  default = "postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable"
}

variable "dev_url" {
  type    = string
  default = "postgres://postgres:postgres@127.0.0.1:5432/initra_dev?sslmode=disable"
}

variable "test_url" {
  type    = string
  default = "postgres://postgres:postgres@127.0.0.1:5432/initra_test?sslmode=disable"
}

variable "prod_url" {
  type    = string
  default = "postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable"
}

env "local" {
  url = var.local_url
  dev = "docker://postgres/16/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }

  schema {
    src = "file://schema"
  }
}

env "dev" {
  url = var.dev_url
  dev = "docker://postgres/16/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }

  schema {
    src = "file://schema"
  }
}

env "test" {
  url = var.test_url
  dev = "docker://postgres/16/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }

  schema {
    src = "file://schema"
  }
}

env "prod" {
  url = var.prod_url

  migration {
    dir = "file://migrations"
  }

  schema {
    src = "file://schema"
  }
}
