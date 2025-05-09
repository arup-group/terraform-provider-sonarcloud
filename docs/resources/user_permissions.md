---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sonarcloud_user_permissions Resource - terraform-provider-sonarcloud"
subcategory: ""
description: |-
  This resource manages the permissions of a user for the whole organization or a specific project.
---

# sonarcloud_user_permissions (Resource)

This resource manages the permissions of a user for the whole organization or a specific project.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `login` (String) The login of the user to set the permissions for.
- `permissions` (Set of String) List of permissions to grant. Available global permissions: [`admin`, `profileadmin`, `gateadmin`, `scan`, `provisioning`]. Available project permissions: ['admin`, `scan`, `codeviewer`, `issueadmin`, `securityhotspotadmin`, `user`].

### Optional

- `project_key` (String) The key of the project to restrict the permissions to.

### Read-Only

- `avatar` (String) The avatar ID of the user.
- `id` (String) The implicit ID of the resource.
- `name` (String) The name of the user.

## Import

Import is supported using the following syntax:

```shell
#!/bin/sh
# import user permissions for the whole organization using <login>
terraform import "sonarcloud_user_permissions.example_user" "user@github.com"

# import user permissions for a specific project using <login>,<project_key>
terraform import "sonarcloud_user_permissions.example_user" "user@github.com,example_project"
```
