## 1.10.0 (Unreleased)

NOTES:
* This release is Terraform 0.12-RC1+ compatible. 

## 1.9.0 (March 15, 2019)

IMPROVEMENTS:

* datasource/hcloud_server: Add `with_status` attribute to get images by their status.
* datasource/hcloud_image: Add `with_status` attribute to get servers by their status.
* datasource/hcloud_volume: Add `with_status` attribute to get volumes by their status.

* Added `with_selector` to all datasources that support label selectors.

NOTES:

* **Deprecation**: datasource/hcloud_server: `selector`, will be removed in the near future.
* **Deprecation**: datasource/hcloud_floating_ip: `selector`, will be removed in the near future.
* **Deprecation**: datasource/hcloud_image: `selector`, will be removed in the near future.
* **Deprecation**: datasource/hcloud_ssh_key: `selector`, will be removed in the near future.
* **Deprecation**: datasource/hcloud_volume: `selector`, will be removed in the near future.

## 1.8.1 (March 12, 2019)

BUG FIXES:
* Fix an error on server creation, when a image id was given instead of a image name.
* Fix an missing error on `terraform plan`, when using an image name which does not exists. 

## 1.8.0 (February 06, 2019)

FEATURES:
* **New Datasource**: `hcloud_server` Lookup a server.

IMPROVEMENTS:
* Add API token length validation

## 1.7.0 (December 18, 2018)

FEATURES:
* **New Datasource**: `hcloud_location` Lookup a location.
* **New Datasource**: `hcloud_locations` Lookup all locations.
* **New Datasource**: `hcloud_datacenter` Lookup a datacenter.
* **New Datasource**: `hcloud_datacenters` Lookup all datacenters.
* Volume Automounting is now available for `hcloud_volume` and `hcloud_volume_attachment`

## 1.6.0 (December 03, 2018)

IMPROVEMENTS:
* datasource/hcloud_image: Add `most_recent` attribute to get the latest image when multiple images has the same label.

BUG FIXES:
* Fix an error on volume_attachment creation, when server was locked.

## 1.5.0 (November 16, 2018)

FEATURES:
* **New Resource**: `hcloud_volume_attachment` Manage the attachment between volumes and servers.

IMPROVEMENTS:
* resources/hcloud_server: Add `backups` attribute to enable or disable backups.

NOTES:
* **Read Only**: resources/hcloud_server: `backup_window`, removed the ability to set the attribute. This attribute is now read only.
* Updated hcloud-go to `v1.11.0`

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
