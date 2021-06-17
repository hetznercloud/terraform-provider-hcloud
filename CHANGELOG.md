Every new changelog can be found under the Github Release page: https://github.com/hetznercloud/terraform-provider-hcloud/releases

## 1.27.0 (June 17, 2021)

FEATURES:
* `hcloud_firewall` resource & datasource: Support GRE & ESP protocol in firewall rules

## 1.26.2 (May 28, 2021)

BUG FIXES:

* Fix invalid checksum for release 1.26.1

## 1.26.1 (May 28, 2021)

BUG FIXES:
* `hcloud_firewall` datasource: `destination_ips` missed in definition
* `hcloud_certificate` resource: panic when parsing certificate chains
  (#359)

## 1.26.0 (March 30, 2021)

* **New Resource** `hcloud_managed_certificate`
* **New Resource** `hcloud_uploaded_certificate`
* **Deprecated Resource** `hcloud_certificate`

## 1.25.2 (March 16, 2021)

BUG FIXES:
* `hcloud_firewall` resource: plugin normalized CIDRs silently.

## 1.25.1 (March 10, 2021)

BUG FIXES:
* `hcloud_firewall` documentation: fix name of `firewall_ids` property.

## 1.25.0 (March 10, 2021)

FEATURES:
* **New Resource**: `hcloud_snapshot`
* **New Resource**: `hcloud_firewall`
* **New Data Source**: `hcloud_firewall`

BUG FIXES:
* `hcloud_server` resource: image had a wrong type (int instead of string) when a server was created from a snapshot
* `hcloud_load_balancer_target` resource: force recreation when changing a target attribute (server_id, ip or label_selector)

NOTES:
* The provider is now built with Go 1.16

## 1.24.1 (February 04, 2021)

BUG FIXES:
* `hcloud_volume` datasource: id is now marked as computed to allow more setups where the id is unknown
* `hcloud_ssh_key` datasource: id is now marked as computed to allow more setups where the id is unknown
* `hcloud_network` datasource: id is now marked as computed to allow more setups where the id is unknown
* `hcloud_image` datasource: id is now marked as computed to allow more setups where the id is unknown
* `hcloud_certificate` datasource: id is now marked as computed to allow more setups where the id is unknown
* `hcloud_volume` resource: Automount is now working when you attach an already existing volume to a server.

## 1.24.0 (January 12, 2021)

FEATURES:
* **New Datasource**: `hcloud_server_type`
* **New Datasource**: `hcloud_server_types`
* New `network` property for `hcloud_server` resource.

BUG FIXES:
* `hcloud_volume` resource: A race condition was fixed, that was called when you tried to create multiple volumes for a single server
* `hcloud_locations` datasource: Use a stable value as IDs instead of a timestamp. We now use a hash of the concatenation of all location IDs as ID
* `hcloud_datacenters` datasource: Use a stable value as IDs instead of a timestamp. We now use a hash of the concatenation of all datacenters IDs as ID

Notes:
* This release is tested against Terraform 0.13.x and 0.14.x. Testing on 0.12.x was removed, therefore Terraform 0.12.x is no longer officially supported

## 1.23.0 (November 03, 2020)

FEATURES:
* `hcloud_network_subnet` supports vSwitch Subnets

Notes:
* The provider was updated to use the Terraform Plugin SDK v2.

## 1.22.0 (October 05, 2020)

FEATURES:

* All `hcloud_*` resources are now importable.

BUG FIXES:
* `hcloud_rdns` resource: It is now possible to import the resource as documented.

## 1.21.0 (September 09, 2020)

CHANGED:

* Un-deprecate `network_id` property of `hcloud_load_balancer_network` and
  `hcloud_server_network` resources.
* Change module path from
  `github.com/terraform-providers/terraform-provider-hcloud` to
  `github.com/hetznercloud/terraform-provider-hcloud`

## 1.20.1 (August 18, 2020)
BUG FIXES:

* `hcloud_certificate` resource: Updating the certificate needs to recreate the certificate.

NOTES:
* The provider is now build with Go 1.15
* We overhauled parts of the underlying test suite

## 1.20.0 (August 10, 2020)

FEATURES:

* Allow updating/resizing a Load Balancer through the
  `load_balancer_type` of `hcloud_load_balancer` resource
* Add support for Load Balancer Label Selector and IP targets.

## 1.19.2 (July 28, 2020)

CHANGED:

* Deprecate `network_id` property of `hcloud_server_network` and
  `hcloud_load_balancer_network` resources. Introduce a `subnet_id`
  property as replacement.

  Both resources require a subnet to be created. Since `network_id`
  references the network and not the subnet there is no explicit
  dependency between those resources. This leads to Terraform creating
  those resources in parallel, which creates a race condition. Users
  stuck with the `network_id` property can create an explicit dependency
  on the subnet using `depends_on` to work around this issue.

BUG FIXES:
* Enable and Disable `proxyprotocol` on a Load Balancer didn't work after creation
* Deleted all Load Balancer services when you changed the `listen_port` of one service
* `hcloud_load_balancer_target` was not idempotent when you add a target that was already defined

NOTES:
* Update to hcloud-go v1.19.0 to fix the bad request issue

## 1.19.1 (July 16, 2020)

NOTES:

* First release under new terraform registry
* Provider was moved to https://github.com/hetznercloud/terraform-provider-hcloud

## 1.19.0 (July 10, 2020)

BUG FIXES:

* Update to hcloud-go v1.18.2 to fix a conflict issue
* Ensure `alias_ip` retain the same order.

NOTES:

* This release uses Terraform Plugin SDK v1.15.0.

## 1.18.1 (July 02, 2020)

BUG FIXES

* Set correct defaults for `cookie_name` and `cookie_lifetime`
  properties of `hcloud_load_balancer_service`.
* Remove unsupported `https` protocol from health check documentation.
* Force recreate of `hcloud_network` if `ip_range` changes.

## 1.18.0 (June 30, 2020)

FEATURES:

* **New Resource**: `hcloud_load_balancer_target` which allows to add a
  target to a load balancer. This resource extends the `target` property
  of the `hcloud_load_balancer` resource.  `hcloud_load_balancer_target`
  should be preferred over the `target` property of
  `hcloud_load_balancer`.

## 1.17.0 (June 22, 2020)

FEATURES:

* **New Datasource**: `hcloud_load_balancer`
* **New Resource**: `hcloud_load_balancer`
* **New Resource**: `hcloud_load_balancer_service`
* **New Resource**: `hcloud_load_balancer_network`

BUG FIXES:

* resources/hcloud_network_route: Fix panic when trying to lookup an already deleted Network route

## 1.16.0 (March 24, 2020)

BUG FIXES:
* resource/hcloud_ssh_key: Fix panic when we update labels in SSH keys
* resource/hcloud_server_network: Fix alias ips ignored on creation of server network
* resource/hcloud_server: Use first assigned `ipv6_address` as value instead of the network address. **Attention: This can be a breaking change**

NOTES:
* This release uses Terraform Plugin SDK v1.8.0.

## 1.15.0 (November 11, 2019)

IMPROVEMENTS:

* resources/hcloud_server: Add retry mechanism for enabling the rescue mode.

NOTES:
* This release uses Terraform Plugin SDK v1.3.0.

## 1.14.0 (October 01, 2019)

NOTES:
* This release uses the Terraform Plugin SDK v1.1.0.

## 1.13.0 (September 19, 2019)

IMPROVEMENTS:

* resources/hcloud_floating_ip: Add `name` attribute to get or set the name of a Floating IP.
* datasource/hcloud_floating_ip: Add `name` attribute to get Floating IPs by their name.

NOTES:

* This release is Terraform 0.12.9+ compatible.
* Updated hcloud-go to `v1.16.0`
* The provider is now tested and build with  Go `1.13`

## 1.12.0 (July 29, 2019)

FEATURES:

* **New Datasource**: `hcloud_ssh_keys` Lookup all SSH keys.

IMPROVEMENTS:

* resources/hcloud_server_network: Add `mac_address` attribute to get the mac address of the Network interface.

BUG FIXES:

* Fix an error on server creation, when an iso id was given instead of an iso name.

NOTES:

* This release is Terraform 0.12.5+ compatible.
* Updated hcloud-go to `v1.15.1`
* Added hcloud-go request debugging when using `TF_LOG`.

## 1.11.0 (July 10, 2019)

FEATURES:

* **New Resource**: `hcloud_network` Manage Networks.
* **New Resource**: `hcloud_network_subnet` Manage Networks Subnets.
* **New Resource**: `hcloud_network_route` Manage Networks Routes.
* **New Resource**: `hcloud_server_network` Manage attachment between servers and Networks.
* **New Datasource**: `hcloud_network` Lookup Networks.

## 1.10.0 (May 14, 2019)

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
