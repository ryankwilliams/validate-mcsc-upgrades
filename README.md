# Validate HCP MC & SC Upgrades (WIP)

This repository is a proof of concept to create a test suite that covers
the use case of validating Management Cluster (MC) and Service Cluster (SC)
can be upgraded for either Y/Z Stream and verify hosted control plane (HCP)
workloads are unaffected by the upgrade.

The test suite to cover this use case can be easily run by invoking
`ginkgo run` directly, from a `compiled binary` or from a `container image`.
The container image option makes this image consumable by other CI frameworks
for easy running. With either method used, all that is required is the
necessary input to authenticate with the necessary clusters/infrastructure.

## Test Suite Design

* Before Suite
  * Authenticate with MC/SC
  * Identify existing MC/SC versions
  * Identify MC/SC upgrade version
  * Deploy N HCP clusters as customer workloads
* After Suite
  * Destroy N HCP clusters that were deployed
* Test Cases:
  * SC upgrade
  * SC health checks are passing post upgrade
  * HCP workloads are unaffected post sc upgrade
  * MC upgrade
  * MC health checks are passing post upgrade
  * HCP workloads are unaffected post mc upgrade

The test suite is ordered and has labels that can be applied to tailor what
operations are performed. For example:

* `ginkgo run`
  * Will run through the entire test suite described above
* `ginkgo run --label-filter="ApplyHCPWorkloads || RemoveHCPWorkloads || SCUpgrade"`
  * Will do everything except upgrade the management cluster
* `ginkgo run --label-filter="ApplyHCPWorkloads || RemoveHCPWorkloads || MCUpgrade"`
  * Will do everything except upgrade the service cluster
* `ginkgo run --label-filter="ApplyHCPWorkloads || SCUpgrade || MCUpgrade"`
  * Will do everything except remove the hcp clusters deployed

## TODO's

* Actual implementation as right now the test suite is just a skeleton
