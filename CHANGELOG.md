## 1.5.0 (Unreleased)
## 1.4.0 (October 18, 2018)

FEATURES:

* **New Resource**: `hcloud_volume` Manage volumes.
* **New Datasource**: `hcloud_volume` Lookup volumes.

NOTES:

* **Deprecation**: resource/hcloud_server: `backup_window`, will be removed in the near future.

## 1.3.0 (September 12, 2018)

FEATURES:

- **New Resource**: `hcloud_rnds` Manage reverse DNS entries for servers and Floating IPs.
* **New Resource**: `hcloud_floating_ip_assignment` Manage the association between Floating IPs and servers.
- **New Datasource**: `hcloud_floating_ip` Lookup Floating ips.
- **New Datasource**: `hcloud_image` Lookup images.
- **New Datasource**: `hcloud_ssh_key` Lookup SSH Keys.
- **New Provider Config**: `poll_interval`  Configures the interval in which actions are polled by the client. Default `500ms`. Increase this interval if you run into rate limiting errors.

IMPROVEMENTS:

* resource/hcloud_server: Add `ipv6_network` attribute.

NOTES:

* Updated hcloud-go to `v1.9.0`

## 1.2.0 (June 07, 2018)

NOTES:

* Switched from MIT licence to MPL2
* removed `reverse_dns` property of `hcloud_floating_ip`, because it was not useable, see https://github.com/hetznercloud/terraform-provider-hcloud/issues/32
* improved test coverage
* updated terraform to `v0.11.7`
* updated hcloud-go to `v1.6.0`
* added log when waiting for an action to complete

BUG FIXES:

* delete records from state that are invalid or are not found by the server
* resource update methods return the result of the read method

## 1.1.0 (March 2, 2018)

* Save hashsum of `user_data`, existing state is migrated
* update hcloud-go to v1.4.0
* update terraform from v0.11.2 to v0.11.3

## 1.0.0 (January 30, 2018)

* Initial release
