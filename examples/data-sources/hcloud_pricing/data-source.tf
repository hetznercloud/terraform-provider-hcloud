data "hcloud_pricing" "prices" {}

output "monthly_traffic_price_per_tb" {
  value = data.hcloud_pricing.prices.server_types[0].prices[0].per_tb_traffic.gross
}
