We maintain a terraform provider that is failing to delete a "primary IP" resource in certain circumstances.

Before we can delete a primary IP resource, we have to unassign it from a server instance in our API.

During our tests of the delete call, we noticed that the state of the primary IP resource references an invalid server instance. This lead to try to unassign the primary IP from an instance that was not assigned. Which made the primary IP resource deletion fail.

That bug is only happening in terraform >=1.9, so we ran a bisect from the last known terraform version to work (1.8.5). We found the problematic commit to be 460c7f3933115c3edf670caacd2ffa489ef4eeb8 https://github.com/hashicorp/terraform/pull/35467

I am unsure if the change linked above is about destroying resource or forgetting resource. But we expect the state of the primary IP resource to be refreshed before we can destroy it.


