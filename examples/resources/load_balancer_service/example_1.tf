resource "hcloud_load_balancer" "load_balancer" {
  name               = "my-load-balancer"
  load_balancer_type = "lb11"
  location           = "nbg1"
}

resource "hcloud_load_balancer_service" "load_balancer_service" {
  load_balancer_id = hcloud_load_balancer.load_balancer.id
  protocol         = "http"

  http {
    sticky_sessions = true
    cookie_name     = "EXAMPLE_STICKY"
  }

  health_check {
    protocol = "http"
    port     = 80
    interval = 10
    timeout  = 5

    http {
      domain       = "example.com"
      path         = "/healthz"
      response     = "OK"
      tls          = true
      status_codes = ["200"]
    }
  }
}
