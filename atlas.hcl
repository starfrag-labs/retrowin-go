data "composite_schema" "retrowin" {
  schema "public" {
    url = "ent://ent/schema"
  }
}

env "local" {
  src = data.composite_schema.retrowin.url
  dev = "docker://postgres/17/dev?search_path=public"
  url = getenv("DATABASE_URL")
}

env "production" {
  src = data.composite_schema.retrowin.url
  dev = "docker://postgres/17/dev?search_path=public"
  url = getenv("DATABASE_URL")
}
