# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project's packages adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Add tolerations to start on `NotReady` nodes.
- Use host network, so chart-operator can install CNI packaged as an app.

## [2.11.0] - 2021-03-19

### Added

- Updating namespace metadata using namespaceConfig in `Chart` CRs.

## [2.10.0] - 2021-03-17

### Added

- Pause Chart CR reconciliation when it has chart-operator.giantswarm.io/paused=true annotation.

### Changed

- Deploy `giantswarm-critical` PriorityClass when it's not found.

## [2.9.0] - 2021-02-03

### Added

- Use diff key when logging differences between the current and desired release.

### Fixed

- Stop updating Helm release if it has failed the previous 5 attempts.

## [2.8.0] - 2021-01-27

### Added

- Add support for skip CRD flag when installing Helm releases.

## [2.7.1] - 2021-01-13

### Fixed

- Only create VPA if autoscaling API group is present.

## [2.7.0] - 2021-01-07

### Added

- Added last reconciled timestamp as metrics.

## [2.6.0] - 2020-12-21

### Added

- Print difference between current release and desired release.

### Changed

- Updated Helm to v3.4.2.

## [2.5.2] - 2020-12-07

### Added

- Add Vertical Pod Autoscaler support.

## [2.5.1] - 2020-12-01

### Fixed

- Fix comparison of last deployed and revision optional fields in status resource.
- Set memory limit and reduce requests.

## [2.5.0] - 2020-11-09

### Added

- Validate the cache in helmclient to avoid state requests when pulling tarballs.
- Call status webhook with token values.

### Fixed

- Update apiextensions to v3 and replace CAPI with Giant Swarm fork.

## [2.4.0] - 2020-10-29

### Added

- Call status webhook when webhook annotation is present.

### Removed

- Remove chartmigration resource as migration from chartconfig to chart CRs is
complete.

## [2.3.5] - 2020-10-13

### Fixed

- Stop repeating helm upgrade for the failed helm release.

## [2.3.4] - 2020-10-01

### Added

- Added release name as a label into the event count metrics.

## [2.3.3] - 2020-09-29

### Fixed

- Updated Helm to v3.3.4.
- Updated Kubernetes dependencies to v1.18.9.
- Update deployment annotation to use checksum instead of helm revision to
reduce how often pods are rolled.
- Increase wait timeout for accessing Kubernetes API from 10s to 120s.

## [2.3.2] - 2020-09-22

### Added

- Added event count metrics for delete, install, rollback and update of Helm releases.

### Fixed

- Fix structs merging error in helmclient.

### Security

- Updated Helm to v3.3.3.

## [2.3.1] - 2020-09-04

### Added

- Add monitoring labels.

### Changed

- Add namespace to logging message.

### Fixed

- Remove memory limits from deployment.

## [2.3.0] - 2020-08-24

### Changed

- Using default DNS policy for control planes.

## [2.2.1] - 2020-08-19

### Changed

- Fixed the timeout value for the namespace resource.

## [2.2.0] - 2020-08-19

### Changed

- Creating namespace before helm operations.

## [2.1.0] - 2020-08-18

## Changed

- Updated Helm to v3.3.0.

## [2.0.0] - 2020-08-12

## Changed

- Updated backward incompatible Kubernetes dependencies to v1.18.5.
- Updated Helm to v3.2.4.
- Fix the rollback in a loop problem.

## [1.0.7] - 2020-08-05

### Changed

- Rollback the helm release in pending-install, pending-upgrade.

## [1.0.6] - 2020-07-24

### Changed

- Disable force upgrades since recreating resources is not supported.
- Graduate Chart CRD to v1.
- Upgrade to operatorkit 1.2.0.

## [1.0.5] - 2020-07-15

### Changed

- Enable force upgrades when chart CR annotation is present.

## [1.0.4] - 2020-07-08

### Changed

- Update MD5 Hash only if chart-operator upgrade the release successfully.
- Make kubernetes wait timeout configurable when installing and updating
releases.
- Set release revision in CR status.

## [v1.0.3] 2020-06-16

### Changed

- Fixed PodSecurityPolicy compatibility problem.

## [v1.0.2] 2020-06-04

### Changed

- Disabled force-upgrade from helmclient.
- Canceling the release resource when migration is done yet.

## [v1.0.1] 2020-05-26

### Changed

- Using helmclient v1.0.1 for security fix.
- Cancel the release resource when the manifest validation failed.

## [v1.0.0] 2020-05-18

### Changed

- Updated to support Helm 3; To keep using Helm 2, please use version 0.X.X.

## [v0.13.0] 2020-04-21

### Changed

- Deploy as a unique app in app collection in control plane clusters.

## [v0.12.4] 2020-04-15

### Changed

- Always set chart CR annotations so update state calculation is accurate.
- Only update failed Helm releases if the chart values or version has changed.

## [v0.12.3] 2020-04-09

### Changed

- Fix problem pushing chart to default app catalog.

### Changed

## [v0.12.2] 2020-04-09

### Changed

- Fix update state calculation and status resource for long running deployments.
- Handle 503 responses when GitHub Pages is unavailable.
- Make HTTP client timeout configurable for pulling chart tarballs in AWS China.
- Switch from dep to go modules.

## [v0.12.1] 2020-03-10

### Removed

- Remove usage of legacy chartconfig CRs in Tiller metrics.

## [v0.12.0] 2020-03-09

### Added

- Add chartmigration resource to allow legacy chartconfig controller to be
removed. ([#358](https://github.com/giantswarm/chart-operator/pull/358))

### Changed

- Improve reason field in chart CR status when installing a chart fails. ([#359](https://github.com/giantswarm/chart-operator/pull/359))
- Use version from chart CR when calculating desired state to reduce number of
HTTP requests to pull chart tarballs. ([#351](https://github.com/giantswarm/chart-operator/pull/353))
- Wait for deleted Helm release before removing finalizer. ([#360](https://github.com/giantswarm/chart-operator/pull/360))
- Do not wait when installing or updating a Helm release takes over 3 seconds.
We check progress in the next reconciliation loop. ([#362](https://github.com/giantswarm/chart-operator/pull/362))

### Removed

- Remove legacy chartconfig controller. ([#365](https://github.com/giantswarm/chart-operator/pull/365))

## [v0.8.0]

### Added

- Separate network policy.

## [v0.7.0]

### Added

- Separate podsecuritypolicy.
- Security context in deployment spec with non-root user.

[Unreleased]: https://github.com/giantswarm/chart-operator/compare/v2.11.0...HEAD
[2.11.0]: https://github.com/giantswarm/chart-operator/compare/v2.10.0...v2.11.0
[2.10.0]: https://github.com/giantswarm/chart-operator/compare/v2.9.0...v2.10.0
[2.9.0]: https://github.com/giantswarm/chart-operator/compare/v2.8.0...v2.9.0
[2.8.0]: https://github.com/giantswarm/chart-operator/compare/v2.7.1...v2.8.0
[2.7.1]: https://github.com/giantswarm/chart-operator/compare/v2.7.0...v2.7.1
[2.7.0]: https://github.com/giantswarm/chart-operator/compare/v2.6.0...v2.7.0
[2.6.0]: https://github.com/giantswarm/chart-operator/compare/v2.5.2...v2.6.0
[2.5.2]: https://github.com/giantswarm/chart-operator/compare/v2.5.1...v2.5.2
[2.5.1]: https://github.com/giantswarm/chart-operator/compare/v2.5.0...v2.5.1
[2.5.0]: https://github.com/giantswarm/chart-operator/compare/v2.4.0...v2.5.0
[2.4.0]: https://github.com/giantswarm/chart-operator/compare/v2.3.5...v2.4.0
[2.3.5]: https://github.com/giantswarm/chart-operator/compare/v2.3.4...v2.3.5
[2.3.4]: https://github.com/giantswarm/chart-operator/compare/v2.3.3...v2.3.4
[2.3.3]: https://github.com/giantswarm/chart-operator/compare/v2.3.2...v2.3.3
[2.3.2]: https://github.com/giantswarm/chart-operator/compare/v2.3.1...v2.3.2
[2.3.1]: https://github.com/giantswarm/chart-operator/compare/v2.3.0...v2.3.1
[2.3.0]: https://github.com/giantswarm/chart-operator/compare/v2.2.1...v2.3.0
[2.2.1]: https://github.com/giantswarm/chart-operator/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/giantswarm/chart-operator/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/giantswarm/chart-operator/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/giantswarm/chart-operator/compare/v1.0.7...v2.0.0
[1.0.7]: https://github.com/giantswarm/chart-operator/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/giantswarm/chart-operator/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/giantswarm/chart-operator/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/giantswarm/chart-operator/compare/v1.0.3...v1.0.4
[v1.0.3]: https://github.com/giantswarm/chart-operator/compare/v1.0.2...v1.0.3
[v1.0.2]: https://github.com/giantswarm/chart-operator/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/giantswarm/chart-operator/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/giantswarm/chart-operator/compare/v0.13.0...v1.0.0
[v0.13.0]: https://github.com/giantswarm/chart-operator/compare/v0.12.4...v0.13.0
[v0.12.4]: https://github.com/giantswarm/chart-operator/compare/v0.12.3...v0.12.4
[v0.12.3]: https://github.com/giantswarm/chart-operator/compare/v0.12.2...v0.12.3
[v0.12.2]: https://github.com/giantswarm/chart-operator/compare/v0.12.1...v0.12.2
[v0.12.1]: https://github.com/giantswarm/chart-operator/compare/v0.12.0...v0.12.1
[v0.12.0]: https://github.com/giantswarm/chart-operator/compare/v0.8.0...v0.12.0
[v0.8.0]: https://github.com/giantswarm/chart-operator/compare/v0.7.0...v0.8.0

[v0.7.0]: https://github.com/giantswarm/chart-operator/releases/tag/v0.7.0
