data "nango_integrations" "all" {}

output "integration_count" {
  value = length(data.nango_integrations.all.integrations)
}

output "integration_names" {
  value = [for integration in data.nango_integrations.all.integrations : integration.display_name]
}
