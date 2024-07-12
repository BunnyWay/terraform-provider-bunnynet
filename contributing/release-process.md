## Release Process

Bunny aims to release at fortnightly cadence. To determine when the next release is due, you can either:

- Review the latest [releases](https://github.com/BunnyWay/terraform-provider-bunnynet/releases).

If a hotfix is needed, the same process outlined below is used however only the 
semantic versioning patch version is bumped.

- Ensure CI is passing for [`main` branch](https://github.com/BunnyWay/terraform-provider-bunnynet/actions?query=branch%3Amain).
- Remove "(Unreleased)" portion from the header for the version you are intending 
  to release (here, 1.0.0). Create a new H2 above for the next unreleased 
  version (here 1.1.0). Example diff:
  ```diff
  + ## 1.1.0 (Unreleased)

  + ## 1.0.0
  - ## 1.0.0 (Unreleased)

  ```
- Bumping the minor version is fine when adding new features. Major version bump should be done if breaking changes are introduced.  
- Create a new GitHub release with the release title exactly matching the tag 
  (e.g. `v1.0.0`) and copy the entries from the CHANGELOG to the release notes.
- A GitHub Action will now build the binaries, documentation and distribute them 
  to the Terraform registry for the [Bunny provider](https://registry.terraform.io/providers/bunnynet/bunnynet/latest).
- Once this is completed, close off the milestone for the current release and 
  open the next that matches the CHANGELOG additions from earlier. Example: close 
  v1.0.0 but open a v1.1.0.