# Changelog

All notable changes to this project will be documented in this file.

## [3.1.0] - 2025-03-25

### 🚀 Features

- *(dagger)* Switch pipeline to dagger

### 🐛 Bug Fixes

- *(deps)* Update docker.io/library/golang docker tag to v1.24.1
- *(deps)* Update module github.com/notaryproject/notation-go to v1.3.1
- *(deps)* Update module github.com/prometheus/client_golang to v1.21.1
- *(deps)* Update module github.com/spf13/cobra to v1.9.1
- *(deps)* Update gitlab.com/act3-ai/asce/go-common digest to 734a59d
- *(deps)* Update module code.cloudfoundry.org/bytefmt to v0.33.0

## [3.0.2] - 2025-02-19

### 🐛 Bug Fixes

- *(deps)* Update module gorm.io/driver/sqlite to v1.5.7
- *(deps)* Update dependency nbconvert to v7.16.6
- *(deps)* Update kubernetes packages to v0.32.2
- *(deps)* Update module code.cloudfoundry.org/bytefmt to v0.29.0
- *(deps)* Update github.com/gomarkdown/markdown digest to 7a1f277
- *(deps)* Update git.act3-ace.com/ace/hub/api/v6 digest to d56eb09
- *(deps)* Update module github.com/notaryproject/notation-go to v1.3.0

### ⚙️ Miscellaneous Tasks

- *(release)* 3.0.2

### Deps

- Switch git.act3-ace.com/ace/data/schema to gitlab.com/act3-ai/asce/data/schema

## [3.0.1] - 2025-02-05

### 🐛 Bug Fixes

- *(ui)* Added background to search bar

### ⚙️ Miscellaneous Tasks

- *(release)* 3.0.1

## [3.0.0] - 2025-01-31

### 🐛 Bug Fixes

- *(linting)* Updated golangci-lint

### 🚜 Refactor

- Remove chi dependency, use public go-common

### ⚙️ Miscellaneous Tasks

- *(release)* 3.0.0

### BREAKING

- [**breaking**] Updated Telemetry module to v3

### Build

- [**breaking**] Correctly trigger breaking semantic release

## [2.1.0] - 2024-12-12

### 🚀 Features

- Added HX-Boost to make navigation faster

### 🐛 Bug Fixes

- *(deps)* Update git.act3-ace.com/ace/hub/api/v6 digest to a3a6971
- *(deps)* Update github.com/gomarkdown/markdown digest to d03b890
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.37
- *(deps)* Update docker.io/library/python docker tag to v3.13
- *(deps)* Update dependency go to v1.23.3
- *(deps)* Update docker.io/library/golang docker tag to v1.23.3
- *(deps)* Update module code.cloudfoundry.org/bytefmt to v0.18.0
- *(deps)* Update module gorm.io/driver/postgres to v1.5.11
- *(deps)* Update dependency devsecops/cicd/pipeline to v20
- Update leaderboard to fix spacing and accessibility issues

### ⚙️ Miscellaneous Tasks

- *(release)* 2.1.0

## [2.0.4] - 2024-12-02

### 🐛 Bug Fixes

- Removed arm64 target for acehub image

### ⚙️ Miscellaneous Tasks

- *(release)* 2.0.4

## [2.0.3] - 2024-12-02

### 🐛 Bug Fixes

- Unpinned acehub pgadmin version

### ⚙️ Miscellaneous Tasks

- *(release)* 2.0.3

## [2.0.2] - 2024-11-27

### 🐛 Bug Fixes

- Pinned pgadmin version to 8.12

### ⚙️ Miscellaneous Tasks

- *(release)* 2.0.2

## [2.0.1] - 2024-11-27

### 🐛 Bug Fixes

- Update modals in various places on bottle details page

### ⚙️ Miscellaneous Tasks

- *(release)* 2.0.1

## [2.0.0] - 2024-11-27

### 🚀 Features

- Add oauth support

### 🐛 Bug Fixes

- *(deps)* Update git.act3-ace.com/ace/hub/api/v6 digest to f535566
- Accessibility issues and add navigation changes
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.35
- *(deps)* Update docker.io/library/golang docker tag to v1.23.2
- *(deps)* Update module github.com/notaryproject/notation-go to v1.2.1
- *(deps)* Update module github.com/prometheus/client_golang to v1.20.5
- *(deps)* Update kubernetes packages to v0.31.2
- *(deps)* Update git.act3-ace.com/ace/hub/api/v6 digest to a3a6971
- *(deps)* Update module code.cloudfoundry.org/bytefmt to v0.16.0
- *(deps)* Update postgres docker tag to v17
- *(deps)* Update git.act3-ace.com/ace/go-common digest to ba34560
- *(deps)* Update module gorm.io/gorm to v1.25.12
- *(deps)* Update github.com/gomarkdown/markdown digest to 72d49d9
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.36
- *(ci)* Gendocs
- *(module)* Moved client module code to v2 package
- *(module)* Restored client v1
- *(module)* Added v1alpha2 version of config
- *(module)* Make v1 client use v1 config
- *(module)* Make v1 client use v1 config
- Search bar fixes and accessibility fixes for catalog page
- Make full card clickable
- Update bottle details page UI
- Versioned package to v2

### ⚙️ Miscellaneous Tasks

- *(release)* 2.0.0

### Fix

- [**breaking**] Updated some of the spacing on the leaderboard

## [1.0.1] - 2024-09-04

### 🐛 Bug Fixes

- *(deps)* Update module code.cloudfoundry.org/bytefmt to v0.5.0
- *(deps)* Update module code.cloudfoundry.org/bytefmt to v0.6.0
- *(deps)* Update module github.com/notaryproject/notation-go to v1.2.0
- *(deps)* Update kubernetes packages to v0.31.0
- *(deps)* Update module github.com/prometheus/client_golang to v1.20.2
- *(deps)* Update module github.com/masterminds/sprig/v3 to v3.3.0

### ⚙️ Miscellaneous Tasks

- *(release)* 1.0.1

## [1.0.0] - 2024-08-28

### 🐛 Bug Fixes

- *(deps)* Update docker.io/library/golang docker tag to v1.23.0
- *(ci)* [**breaking**] Minor typo

### ⚙️ Miscellaneous Tasks

- *(release)* 1.0.0

## [0.21.4] - 2024-08-28

### 🐛 Bug Fixes

- [**breaking**] Metric sorting query for postgres
- *(deps)* Update git.act3-ace.com/ace/hub/api/v6 digest to 9a83119
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.29

### ⚙️ Miscellaneous Tasks

- *(release)* 0.21.4

## [0.21.3] - 2024-08-20

### 🐛 Bug Fixes

- *(db)* Postgres common label query

### ⚙️ Miscellaneous Tasks

- *(release)* 0.21.3

## [0.21.2] - 2024-08-16

### 🐛 Bug Fixes

- *(helm)* Helm db password value

### ⚙️ Miscellaneous Tasks

- *(release)* 0.21.2

## [0.21.1] - 2024-08-16

### 🐛 Bug Fixes

- *(helm)* Chart values data types and added correct examples

### ⚙️ Miscellaneous Tasks

- *(release)* 0.21.1

## [0.21.0] - 2024-08-14

### 🚀 Features

- Leaderboard v2

### 🐛 Bug Fixes

- *(deps)* Update module github.com/go-chi/chi/v5 to v5.0.14
- *(deps)* Update module github.com/gorilla/schema to v1.4.0
- *(deps)* Update module k8s.io/apimachinery to v0.30.2
- *(deps)* Update module gorm.io/driver/sqlite to v1.5.6
- *(deps)* Update module github.com/spf13/cobra to v1.8.1
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.22
- *(deps)* Update helm release postgresql to v15.5.9
- *(deps)* Update code.cloudfoundry.org/bytefmt digest to 7a5a4f8
- Catalog UI tweaks
- *(deps)* Update module github.com/notaryproject/notation-core-go to v1.0.3
- *(deps)* Update module gorm.io/driver/postgres to v1.5.9
- *(deps)* Update module gorm.io/gorm to v1.25.11
- *(deps)* Update github.com/gomarkdown/markdown digest to 034f12a
- *(deps)* Update code.cloudfoundry.org/bytefmt digest to d61d30b
- *(deps)* Update docker.io/library/ubuntu docker tag to v24
- *(deps)* Update module github.com/go-chi/chi/v5 to v5.1.0
- *(deps)* Update module github.com/notaryproject/notation-go to v1.1.1
- *(deps)* Update module k8s.io/apimachinery to v0.30.3
- *(deps)* Update module github.com/microcosm-cc/bluemonday to v1.0.27
- "open in hub" links
- *(ci)* Add go build vars
- *(helm)* Removed postgresql chart dependency
- *(helm)* Helm chart linting
- *(ci)* Removed unneeded semantic release step override in CI
- *(ci)* Restored important parts of semantic release override

### ⚙️ Miscellaneous Tasks

- *(release)* 0.21.0

### Misc

- Updated generated docs

## [0.20.1] - 2024-05-26

### 🐛 Bug Fixes

- Make /tmp writable in the helm chart deployment
- User in helm chart did not match container
- Lint issues
- More lint issues in GO

### ⚙️ Miscellaneous Tasks

- *(release)* 0.20.1

## [0.20.0] - 2024-05-24

### 🚀 Features

- 193 Metrics should to be displayed in the catalog
- Search by bottle location
- Add IDs to HTML headers to make them linkable
- Added extraPodLabels value to chart

### 🐛 Bug Fixes

- Nil pointer panic when search error occurs
- Update verify.sh
- Add fips arm64 job
- *(deps)* Update dependency go to v1.22.2
- *(deps)* Update github.com/gomarkdown/markdown digest to 642f0ee
- Leaderboard value column matches current search param
- 192 Deprecated Bottles window has invisible text
- *(deps)* Update code.cloudfoundry.org/bytefmt digest to 6038236
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.13
- *(deps)* Update module gorm.io/gorm to v1.25.10
- *(deps)* Update dependency nbconvert to v7.16.4
- *(deps)* Update helm release postgresql to v15.2.9
- Linting issue SearchByRepository comment
- Jupyter notebook convert viewer
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.16
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.17
- *(deps)* Update helm release postgresql to v15.3.3
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.18
- *(deps)* Update helm release postgresql to v15.4.0
- *(deps)* Update module k8s.io/apimachinery to v0.30.1
- *(deps)* Update module github.com/prometheus/client_golang to v1.19.1
- *(deps)* Update module code.cloudfoundry.org/bytefmt to v0.0.0-20240522170716-2951b8ebd80e
- *(deps)* Update docker.io/library/golang docker tag to v1.22.3
- Updated signature mediatype for notary

### ⚙️ Miscellaneous Tasks

- *(release)* 0.20.0

## [0.19.2] - 2024-04-22

### 🐛 Bug Fixes

- Leaderboard metrics optional

### ⚙️ Miscellaneous Tasks

- *(release)* 0.19.2

## [0.19.1] - 2024-04-19

### 🐛 Bug Fixes

- Acehub image rollback base image to fix pgadmin pull

### ⚙️ Miscellaneous Tasks

- *(release)* 0.19.1

## [0.19.0] - 2024-04-19

### 🚀 Features

- Signature count on about page
- Infinite catalog scroll

### 🐛 Bug Fixes

- *(deps)* Updated go and non-go dependencies
- *(ci)* Bump pipeline
- *(deps)* Update dependencies
- *(deps)* Update docker.io/library/golang docker tag to v1.22.2
- *(deps)* Update helm release postgresql to v15.2.4
- *(deps)* Update module git.act3-ace.com/ace/data/schema to v1.2.11
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.7
- *(ci)* Add gorelease
- Fix paths and bump schema
- Add Notary style signature validation to signature processing
- Add conditional to required field public key in signature validation
- Lint issues
- Another cogint lint fix attempt
- Move signature validation to client for api consumers to use
- Import cycle and lint errors
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.9
- Update linters
- Unused log instance
- *(deps)* Update helm release postgresql to v15.2.5
- *(deps)* Update code.cloudfoundry.org/bytefmt digest to 335139c
- *(deps)* Update dependency devsecops/cicd/pipeline to v19.0.11
- *(deps)* Update module k8s.io/apimachinery to v0.30.0
- Misnamed struct member for parsing subject descriptor
- Gorelease verify
- Gorelease script supporting files
- Disable gorelease verify

### 📚 Documentation

- Added .codereportignore and added codereport generated files to .gitignore

### ⚙️ Miscellaneous Tasks

- *(ci)* Renovate update
- *(release)* 0.19.0

## [0.18.7] - 2024-03-25

### 🐛 Bug Fixes

- *(deps)* Postgres 14 update

### ⚙️ Miscellaneous Tasks

- *(release)* 0.18.7

## [0.18.6] - 2024-03-25

### 🐛 Bug Fixes

- Processing manifest includes needed field

### ⚙️ Miscellaneous Tasks

- *(release)* 0.18.6

## [0.18.5] - 2024-03-22

### 🐛 Bug Fixes

- Cleanup unused entries in dockerignore
- Tune GOMAXPROCS
- Catalog links when DefaultBottleSelectors is set
- *(jq)* Corrected the jq filter to work with the new logging tools
- Revert "fix(deps): updated postgresql chart to 14.2.3"

### ⚙️ Miscellaneous Tasks

- *(release)* 0.18.5

## [0.18.4] - 2024-03-20

### 🐛 Bug Fixes

- *(build)* Call `make deps` before building the binary in ci

### ⚙️ Miscellaneous Tasks

- *(release)* 0.18.4

## [0.18.3] - 2024-03-20

### 🐛 Bug Fixes

- *(build)* Prod container image working directory update for proper permissions

### ⚙️ Miscellaneous Tasks

- *(release)* 0.18.3

## [0.18.2] - 2024-03-18

### 🐛 Bug Fixes

- Dockerfile dont copy assets

### ⚙️ Miscellaneous Tasks

- *(release)* 0.18.2

## [0.18.1] - 2024-03-18

### 🐛 Bug Fixes

- *(ko)* Remove kodata directory since we bake them into the executable now
- Remove broken prometheus support from postgres
- Prometheus metrics for HTTP duration
- *(deps)* Updated postgresql chart to 14.2.3
- *(helm)* Fixed empty value for db in chart
- *(docs)* Removed old docs

### ⚙️ Miscellaneous Tasks

- *(release)* 0.18.1

## [0.18.0] - 2024-03-14

### 🚀 Features

- HTMX Search
- Web assets embedded in binary

### 🐛 Bug Fixes

- *(deps)* Bump to debian12 and go 1.22 for building
- Switch to go-common for running the http server
- Remove obsolete TODO
- Update MAINTAINERS file
- Lower our CVEs to a single medium vulnerability
- *(ci)* Bump the pipeline and add a builder for the ipynb image
- *(docs)* Update markdownlint files
- *(deps)* Bump go-common
- *(pipeline)* Fixed embedding docs
- *(deps)* Update postgres docker tag to v16
- *(dep)* Bumped go-common
- *(deps)* Bumped pgx version

### ⚙️ Miscellaneous Tasks

- *(release)* 0.18.0

## [0.17.0] - 2024-01-16

### 🚀 Features

- *(ui)* Added bottle attribute iconography
- *(ui)* Lineage graph improvements using go-echarts
- Bottle signature support

### 🐛 Bug Fixes

- Bump deps
- Bump the pipeline to v15
- Bump dependencies
- Testdata cleanup
- *(deps)* Update docker.io/library/python docker tag to v3.12
- *(ci)* Update pipeline version
- *(docs)* Comment out GL issue by email/add Web opt
- Bump pipeline to v16
- *(deps)* Upgraded dependencies
- *(deps)* Update docker images

### 📚 Documentation

- Update Analytics Gateway docs

### ⚙️ Miscellaneous Tasks

- *(release)* 0.17.0

### Fix

- Background color of pills in modal

## [0.16.3] - 2023-09-29

### 🐛 Bug Fixes

- *(ci)* Upgraded the pipelines
- *(lint)* Minor comment additions
- Label text on homepage is white on dark background
- Govulncheck issue
- *(ci)* Update ci pipeline

### 🚜 Refactor

- *(logging)* Switch logging to `log/slog`

### ⚙️ Miscellaneous Tasks

- *(release)* 0.16.3

## [0.16.2] - 2023-07-12

### 🐛 Bug Fixes

- *(sql)* Groupby clause for aggregate function unique pull count

### ⚙️ Miscellaneous Tasks

- *(release)* 0.16.2

## [0.16.1] - 2023-06-29

### 🐛 Bug Fixes

- *(ci)* Increase memory limit for acehub job
- Postgres fix

### ⚙️ Miscellaneous Tasks

- *(release)* 0.16.1

## [0.16.0] - 2023-06-15

### 🚀 Features

- SkipInvalid flag added to client uploads

### 🐛 Bug Fixes

- *(deps)* Update dependency devsecops/cicd/pipeline to v9.0.36
- Update deps
- Release process upgrades to align with ace/data/tool
- Regression in gorm sqlite driver
- Improve search errors and results feedback
- *(deps)* Upgraded nbconvert
- *(doc)* Update REST API and Web UI links
- *(doc)* Update config API doc link
- *(ci)* Bump CI version
- Update pipeline
- Moving swagger out of release.sh
- Added yq to the "generate docs" job
- Using v4 of yq
- Makefile deps
- Auth in "generate docs" job was broken
- PATH in job
- Trying again
- And again

### ⚙️ Miscellaneous Tasks

- *(release)* 0.16.0

## [0.15.13] - 2023-03-27

### 🐛 Bug Fixes

- Aggregate pull score query

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.13

## [0.15.12] - 2023-03-24

### 🐛 Bug Fixes

- Bump pipeline version

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.12

## [0.15.11] - 2023-03-20

### 🐛 Bug Fixes

- Removed extraneous print

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.11

## [0.15.10] - 2023-03-16

### 🐛 Bug Fixes

- Ace hub image

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.10

## [0.15.9] - 2023-03-16

### 🐛 Bug Fixes

- Hover states and other UI issues
- Bump ci to v9.0.0
- *(deps)* Update docker.io/busybox docker tag to v1.36
- *(deps)* Update docker.io/library/python docker tag to v3.11
- *(deps)* Update dependency devsecops/cicd/pipeline to v9.0.8
- Updated schema and fixed apidocs target in Makefile
- Release script

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.9

## [0.15.8] - 2023-01-23

### 🐛 Bug Fixes

- Label hover

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.8

## [0.15.7] - 2023-01-04

### 🐛 Bug Fixes

- Change pill hover-state color
- Upgrade schema to the latest release
- Text wrapping on the annotations popover
- Upgrade schema and skaffold and ci
- Upgrade go deps and nbconvert
- Update the message about using the telemetry server
- Updated dependencies and bumped to the new pipeline
- Also remove empty part selectors when removing a selector

### 🚜 Refactor

- Moved more functionality into internal packages

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.7

## [0.15.6] - 2022-12-12

### 🐛 Bug Fixes

- Bug in selector matching
- Bumped schema again

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.6

## [0.15.5] - 2022-12-12

### 🐛 Bug Fixes

- Update Styling to Match Ace Hub
- Minor fix for the makefile
- *(ci)* Bump CI pipeline
- Test script
- Added the "v" prefix to tag in values.yaml
- Issues w/ color schemes on BootStrap classes
- Removing fs watcher

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.2
- *(release)* 0.15.3
- *(release)* 0.15.4
- *(release)* 0.15.5

## [0.15.4] - 2022-12-05

### 🐛 Bug Fixes

- Test script

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.4

## [0.15.3] - 2022-12-02

### 🐛 Bug Fixes

- *(ci)* Bump CI pipeline

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.3

## [0.15.2] - 2022-12-02

### 🐛 Bug Fixes

- Minor fix for the makefile

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.2

## [0.15.1] - 2022-11-24

### 🐛 Bug Fixes

- Bumped to support FIPS again
- *(ci)* Fix artifact path
- *(ci)* Bump again
- Added a test case for v1 bottles

### 🚜 Refactor

- Updated schema

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.1

## [0.15.0] - 2022-11-23

### 🚀 Features

- Upgrade to bottle v1

### 🐛 Bug Fixes

- Bumped schema version
- *(deps)* Update docker.io/library/golang docker tag to v1.19.3
- *(deps)* Update helm release postgresql to v11.9.13
- Test bottle dependencies

### 🚜 Refactor

- Moved ParseSourceURI to schema

### ⚙️ Miscellaneous Tasks

- *(release)* 0.15.0

## [0.14.2] - 2022-11-15

### 🐛 Bug Fixes

- Typo in README
- Updated deps again

### ⚙️ Miscellaneous Tasks

- *(release)* 0.14.2

## [0.14.1] - 2022-11-15

### 🐛 Bug Fixes

- Switch to ci-bin
- *(deps)* Update dependency postgres to v15
- *(deps)* Update dependency nbconvert to v7.2.1
- *(deps)* Update dependency docker.io/library/golang to v1.19.2
- *(deps)* Update helm release postgresql to v11.9.1
- Bump CI pipeline
- Bump go-common

### ⚙️ Miscellaneous Tasks

- *(release)* 0.14.1

## [0.14.0] - 2022-10-24

### 🚀 Features

- Added the filter subcommand
- Stricter manifest validation

### 🐛 Bug Fixes

- Output the jq filter to stdout
- Upgrade to the latest schema

### 🚜 Refactor

- Moved manifest validation to schema

### ⚙️ Miscellaneous Tasks

- *(release)* 0.14.0

## [0.13.4] - 2022-10-12

### 🐛 Bug Fixes

- Bumped bottle processor version
- Added a BottleDetailURL function

### 🚜 Refactor

- Switched to the cli being in the cli folder

### ⚙️ Miscellaneous Tasks

- *(release)* 0.13.4

## [0.13.3] - 2022-10-03

### 🐛 Bug Fixes

- EnvPathOr

### ⚙️ Miscellaneous Tasks

- *(release)* 0.13.3

## [0.13.2] - 2022-10-03

### 🐛 Bug Fixes

- Added EnvPathOr

### ⚙️ Miscellaneous Tasks

- *(release)* 0.13.2

## [0.13.1] - 2022-09-30

### 🐛 Bug Fixes

- Improve auto generated docs

### ⚙️ Miscellaneous Tasks

- *(release)* 0.13.1

## [0.13.0] - 2022-09-30

### 🚀 Features

- Added a MatchAny function for selectors

### 🐛 Bug Fixes

- Error handling

### ⚙️ Miscellaneous Tasks

- *(release)* 0.13.0

## [0.12.16] - 2022-09-18

### 🐛 Bug Fixes

- Improved handling of /api in the URL paths

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.16

## [0.12.15] - 2022-09-17

### 🐛 Bug Fixes

- Added documentation for plogs

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.15

## [0.12.14] - 2022-09-15

### 🐛 Bug Fixes

- *(deps)* Update module go.uber.org/zap to v1.23.0
- *(deps)* Update helm release postgresql to v11.7.6

### 🚜 Refactor

- Moved code out of the telemetry command

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.14

### Resolves

- Part labels are clickable (they should not be)

## [0.12.13] - 2022-08-27

### 🐛 Bug Fixes

- More resources for jobs
- Added memory to KO

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.13

## [0.12.12] - 2022-08-27

### 🐛 Bug Fixes

- Bumped KO jobs resources

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.12

## [0.12.11] - 2022-08-26

### 🐛 Bug Fixes

- Improved version handling based on Justen's feedback
- Bumped pipelines to hopefully make my pipeline green
- More lint issues
- Removed a dead route
- Dockerfile entrypoint was wrong

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.11

## [0.12.10] - 2022-08-23

### 🐛 Bug Fixes

- CI pipelines upgraded to fix buildkit

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.10

## [0.12.9] - 2022-08-23

### 🐛 Bug Fixes

- Version handling when built with "go build"

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.9

## [0.12.8] - 2022-08-22

### 🐛 Bug Fixes

- Bump to go 1.19

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.8

## [0.12.7] - 2022-08-22

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.7

## [0.12.6] - 2022-08-22

### 🐛 Bug Fixes

- *(deps)* Update dependency nbconvert to v6.5.3
- Trying to fix version

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.6

## [0.12.5] - 2022-08-19

### 🐛 Bug Fixes

- Updated the template

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.5

## [0.12.4-alpha.1] - 2022-08-19

### 🐛 Bug Fixes

- Added a comment

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.4-alpha.1

## [0.12.4] - 2022-08-19

### 🐛 Bug Fixes

- Added beta and alpha branch support to semantic release

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.4

## [0.12.3] - 2022-08-19

### 🐛 Bug Fixes

- Trying again with the pipeline

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.3

## [0.12.2] - 2022-08-19

### 🐛 Bug Fixes

- Trying CI again

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.2

## [0.12.1] - 2022-08-18

### 🐛 Bug Fixes

- *(hub)* Use ACT3_TOKEN to integrate with the pipeline

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.1

## [0.12.0] - 2022-08-18

### 🚀 Features

- Upgraded postgres chart

### 🐛 Bug Fixes

- Redaction
- *(deps)* Update dependency devsecops/cicd/pipeline to v7
- Switch back to xdg
- Direct uploading was not working
- Propogate errors
- *(deps)* Update dependency devsecops/cicd/pipeline to v7.0.4
- *(deps)* Update helm release postgresql to v11.6.20
- *(deps)* Revert template "go-cli" to v1.0.13 (checkpoint commit made by act3-pt) (go-cli:v1.0.13)
- *(deps)* Revert template "go-cli" to v1.0.14 (checkpoint commit made by act3-pt) (go-cli:v1.0.14)
- *(deps)* Update dependency devsecops/cicd/pipeline to v7.0.8
- *(deps)* Update helm release postgresql to v11.6.25
- *(deps)* Update helm release postgresql to v11.7.1
- Skaffold now work
- *(ci)* Bumped to fix helm chart issues
- *(ci)* Bumped CI again to fix the chart deps

### ⚙️ Miscellaneous Tasks

- *(release)* 0.12.0

### Refact

- Use the default os.UserConfigDir() function instead of the package
- Switch to go-chi
- Added act3-pt and synced with the template

## [0.11.0] - 2022-07-11

### 🚀 Features

- Added act3-pt
- Added a new doc generation approach

### 🐛 Bug Fixes

- Add file ".version.yml" (created by act3-pt)
- *(deps)* Update helm values docker.io/busybox to v1.35
- *(deps)* Update dependency nbconvert to v6.5.0
- *(deps)* Update dependency jinja2 to v3.1.2
- Removed old doc approach
- Added the api gen docs config file
- We can now output logs to stdout or stderr via a flag

### ⚙️ Miscellaneous Tasks

- *(release)* 0.11.0

## [0.10.14] - 2022-07-08

### 🐛 Bug Fixes

- Expose redaction for the Location struct

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.14

## [0.10.13] - 2022-07-08

### 🐛 Bug Fixes

- Removed logging from the telemetry action
- Missed a sample config
- We are supposed to call sync before we are done with the log in case anything is buffered
- Fixed a lint issue and separated out boiler plate code
- Made the redaction code reusable

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.13

### Refact

- Moved sample configs into the versioned directory
- Added a ConfigSpec to make MarshalLog easier to implement

## [0.10.12] - 2022-06-28

### 🐛 Bug Fixes

- Updated some comments

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.12

### Refact

- Command line interface into composable actions
- More refactoring to make the root command not special
- Types no longer depends on httputils
- Moved the scheme out of Load

## [0.10.11] - 2022-06-16

### 🐛 Bug Fixes

- Updated the docs

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.11

### Refact

- Improved the way we handle versioning

## [0.10.10] - 2022-06-16

### 🐛 Bug Fixes

- Removed version.go

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.10

## [0.10.9] - 2022-06-16

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.9

## [0.10.8] - 2022-06-16

### 🐛 Bug Fixes

- Use filter-coverage.sh in Makefile
- *(ci)* Update to gitlab 15
- Added a new source of version information

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.8

## [0.10.7] - 2022-06-09

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.7

## [0.10.6] - 2022-06-09

### 🐛 Bug Fixes

- *(ci)* Get build arg for gitlab credentials
- *(hub)* Working to get the dockerfile building again

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.6

## [0.10.5] - 2022-06-08

### 🐛 Bug Fixes

- Moved the interface checking to compile time
- Updated act3-dev-tools version

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.5

### Refact

- Better isolated client from the database dependencies

## [0.10.4] - 2022-06-06

### 🐛 Bug Fixes

- Updated acehub to not use s3.lynx for apt repo

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.4

## [0.10.3] - 2022-06-06

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.3

## [0.10.2] - 2022-05-27

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.2

## [0.10.1] - 2022-05-27

### 🐛 Bug Fixes

- *(ci)* Ensure the executables are available for the build process
- *(ci)* Added mroe needs statements for the tags.txt

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.1

## [0.10.0] - 2022-05-27

### 🚀 Features

- Upgraded schema to v1.0.0 to include KRM

### 🐛 Bug Fixes

- Upgraded golangci-lint
- *(ci)* Fixed unit test job name
- Adding dependency for the linux binary
- Trying needs
- Build issues from the MR
- *(ci)* Bumped the pipeline to fix a failure

### ⚙️ Miscellaneous Tasks

- *(release)* 0.10.0

## [0.9.0] - 2022-05-19

### 🚀 Features

- Added part selectors
- YAML view of bottle definition

### 🐛 Bug Fixes

- Bump pipeline version
- Added back in redaction
- Typo in flag name
- Defaulting for config
- Removed signals (we now import it)
- Removed logger global
- Missing authors are fine
- Unit tests
- Add file ".version.yml" (created by act3-pt)
- Add file "config.js" (created by act3-pt)
- Add file "renovate.json" (created by act3-pt)
- CI build and lint issue
- Do not produce duplicate records when re-processing
- Semantic-release needs branches now that we are using "main"
- Added validation for descriptors in manifests
- Sort keys in preload to display consistently

### 🚜 Refactor

- Moved client command into a new package
- Applied changes back to manifests

### ⚙️ Miscellaneous Tasks

- *(release)* 0.9.0

## [0.8.19] - 2022-04-06

### 🐛 Bug Fixes

- Name of env for ace-dt

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.19

## [0.8.18] - 2022-03-28

### 🐛 Bug Fixes

- Added a ServiceMonitor

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.18

## [0.8.17] - 2022-03-28

### 🐛 Bug Fixes

- Added commit version to the version string

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.17

## [0.8.16] - 2022-03-28

### 🐛 Bug Fixes

- Minor webapp formatting changes
- Added logging for artifacts
- Nbconvert

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.16

## [0.8.15] - 2022-03-25

### 🐛 Bug Fixes

- Added emphasis to make the viewer list
- Removed redundant text in templates and added helper text
- Bump bottle processor version

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.15

## [0.8.14] - 2022-03-24

### 🐛 Bug Fixes

- A few more changes of x.bottle

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.14

## [0.8.13] - 2022-03-24

### 🐛 Bug Fixes

- Added more environment variables for viewing artifacts in ACE Hub

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.13

## [0.8.12] - 2022-03-23

### 🐛 Bug Fixes

- Removed FTS from searching for now

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.12

## [0.8.11] - 2022-03-23

### 🐛 Bug Fixes

- Fixed bottle name again to be compatible

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.11

## [0.8.10] - 2022-03-23

### 🐛 Bug Fixes

- Changed the volume name to not conflit with ACE Hub

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.10

## [0.8.9] - 2022-03-23

### 🐛 Bug Fixes

- Configmap change annotation was missing
- Chart lint issue

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.9

## [0.8.8] - 2022-03-23

### 🐛 Bug Fixes

- Markdownlint fixes
- Upgrade dockerfiles to GO 1.18
- Nbconvert version

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.8

## [0.8.7] - 2022-03-22

### 🐛 Bug Fixes

- Changed some log levels in migration
- Moved the bottle pulls to the sidebar
- Moved aliases next to the digest
- Markdown lint that I missed
- Upgraded the pipeline and the code to GO 1.18
- Auth in unit test before_script was missing

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.7

## [0.8.6] - 2022-03-14

### 🐛 Bug Fixes

- Changed references from lynx to lion

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.6

## [0.8.5] - 2022-03-12

### 🐛 Bug Fixes

- Acehub image

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.5

## [0.8.4] - 2022-03-12

### 🐛 Bug Fixes

- Lint issue
- Made the global config private

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.4

## [0.8.3] - 2022-03-08

### 🐛 Bug Fixes

- Jupyter field was not set
- Build error

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.3

## [0.8.2] - 2022-02-24

### 🐛 Bug Fixes

- Added VERSION

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.2

## [0.8.1] - 2022-02-24

### 🐛 Bug Fixes

- Updated the docs to reflect the easier install procedure
- Bake the version into source code

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.1

## [0.8.0] - 2022-02-24

### 🚀 Features

- Added a search by Author name and email functionality

### 🐛 Bug Fixes

- Added a check when uploading to make sure a header is set
- Removed replace in go.mod

### ⚙️ Miscellaneous Tasks

- *(release)* 0.8.0

## [0.7.1] - 2022-02-10

### 🐛 Bug Fixes

- Bumped CI to enable CGO by default

### ⚙️ Miscellaneous Tasks

- *(release)* 0.7.1

## [0.7.0] - 2022-02-09

### 🚀 Features

- Enabled client-side caching

### 🐛 Bug Fixes

- Moved building out of the dockerfile so we can support private repositories easily.

### ⚙️ Miscellaneous Tasks

- *(release)* 0.7.0

## [0.6.0] - 2022-02-09

### 🚀 Features

- Added cookie support to upload and download

### 🐛 Bug Fixes

- Docker lint issues
- Functional test now builds separate from running the telemetry server to avoid a timeout

### ⚙️ Miscellaneous Tasks

- *(release)* 0.6.0

## [0.5.2] - 2022-01-25

### 🐛 Bug Fixes

- Run make deps in the Dockerfile

### ⚙️ Miscellaneous Tasks

- *(release)* 0.5.2

## [0.5.1] - 2022-01-25

### 🐛 Bug Fixes

- Removed debug image

### ⚙️ Miscellaneous Tasks

- *(release)* 0.5.1

## [0.5.0] - 2022-01-25

### 🚀 Features

- Added a "from-latest" to better support mirroring of all types
- Added an about page with a version
- Markdown support for artifacts
- Initial cut at "Open in ACE Hub"
- Catalog now supports searching for parents and children
- Added support for non-bottles and unknown bottles
- Added parent and children of to the leaderboard

### 🐛 Bug Fixes

- Postgres now seems to work
- *(chart)* Set priorityClassName: service-critical as default in the deployment template
- Functional test

### ⚙️ Miscellaneous Tasks

- *(release)* 0.5.0

### Assets/static/css/bootstrap-icons.css,assets/static/css/bootstrap-icons.svg,assets/static/css/bootstrap.min.css,assets/static/css/bootstrap.min.css.map,assets/static/js/bootstrap.bundle.min.js,assets/static/js/bootstrap.bundle.min.js.map,assets/static/js/leader-line.min.js

- Convert to Git LFS

## [0.4.0] - 2021-12-23

### 🚀 Features

- Added selectors to source references

### 🐛 Bug Fixes

- Added EXPOSE to dockerfile to support gitlab CI

### ⚙️ Miscellaneous Tasks

- *(release)* 0.4.0

## [0.3.2] - 2021-12-22

### 🐛 Bug Fixes

- Bump CI to fix image digest in the chart

### ⚙️ Miscellaneous Tasks

- *(release)* 0.3.2

## [0.3.1] - 2021-12-21

### 🐛 Bug Fixes

- Logging bug
- Removed duplicate manifestations

### ⚙️ Miscellaneous Tasks

- *(release)* 0.3.1

## [0.3.0] - 2021-12-17

### 🚀 Features

- Added a handler for artifacts
- Switch to bandwidth in events

### 🐛 Bug Fixes

- Add more linters
- Unit tests

### ⚙️ Miscellaneous Tasks

- *(release)* 0.3.0

## [0.2.3] - 2021-12-16

### 🐛 Bug Fixes

- Update CI to fix sub paths
- Build

### ⚙️ Miscellaneous Tasks

- *(release)* 0.2.3

## [0.2.2-alpha.1] - 2021-12-15

### 🐛 Bug Fixes

- Trying --cache=false

### ⚙️ Miscellaneous Tasks

- *(release)* 0.2.2-alpha.1

## [0.2.2] - 2021-12-15

### 🐛 Bug Fixes

- Add jq highlighting properly
- Added a hub target for the Makefile

### ⚙️ Miscellaneous Tasks

- *(release)* 0.2.2

## [0.2.1] - 2021-12-15

### 🐛 Bug Fixes

- Lint
- CI build

### ⚙️ Miscellaneous Tasks

- *(release)* 0.2.1

## [0.2.0] - 2021-12-15

### 🚀 Features

- Added bottle validation

### 🐛 Bug Fixes

- Time handling in "additional locations"
- Lint by moving validation code out of bottle processor
- Unit test
- Docs

### ⚙️ Miscellaneous Tasks

- *(release)* 0.2.0

## [0.1.4] - 2021-12-07

### 🐛 Bug Fixes

- Postgres regression
- Another postgres fix

### ⚙️ Miscellaneous Tasks

- *(release)* 0.1.4

## [0.1.3] - 2021-12-04

### 🐛 Bug Fixes

- CLI flag parsing
- Download
- Switched to SHA256 for the Canonical digest for now.

### ⚙️ Miscellaneous Tasks

- *(release)* 0.1.3

## [0.1.2] - 2021-11-18

### 🐛 Bug Fixes

- Coverage

### ⚙️ Miscellaneous Tasks

- *(release)* 0.1.2

## [0.1.1] - 2021-11-17

### 🐛 Bug Fixes

- Location handler
- For FTS to work with postgres we need a newer version of postgres than the chart.

### ⚙️ Miscellaneous Tasks

- *(release)* 0.1.1

## [0.1.0] - 2021-11-16

### 🚀 Features

- Reprocessing of manifests and events
- Numeric values for labels

### 🐛 Bug Fixes

- Sort order dropdown
- Bumped CI version
- Alert if clipboard is unavaillable
- *(chart)* Use docker.io in image reference
- Readme
- Unit tests
- Better error handling

### ⚙️ Miscellaneous Tasks

- *(release)* 0.1.0

## [0.0.12] - 2021-11-08

### 🐛 Bug Fixes

- Made postgres a local chart
- Added helm dependency back in but removed the repository

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.12

## [0.0.11] - 2021-11-08

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.11

## [0.0.10] - 2021-11-08

### 🐛 Bug Fixes

- Trying again

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.10

## [0.0.9] - 2021-11-08

### 🐛 Bug Fixes

- Trying 1.6.0 again
- Ace hub image

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.9

## [0.0.8] - 2021-11-08

### 🐛 Bug Fixes

- Bumped CI
- Context for kaniko

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.8

## [0.0.7] - 2021-11-07

### 🐛 Bug Fixes

- Hopefully this works
- Added webapp unit tests

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.7

## [0.0.6] - 2021-11-07

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.6

## [0.0.5] - 2021-11-07

### 🐛 Bug Fixes

- Improved a unit test
- More logging to the CI

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.5

## [0.0.4] - 2021-11-07

### 🐛 Bug Fixes

- *(ci)* Trying again

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.4

## [0.0.3] - 2021-11-07

### 🐛 Bug Fixes

- CI

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.3

## [0.0.2] - 2021-11-07

### 🐛 Bug Fixes

- Added version file
- Ace hub image

### ⚙️ Miscellaneous Tasks

- *(release)* 0.0.2

## [0.0.1] - 2021-11-07

### 🚀 Features

- Added a stub of the chart with postgres as a subchart
- [**breaking**] Added support for any hash function
- Added reprocessing logic
- *(load-cmd)* Upload is now optional
- Added manifest table and handler
- Added manifestations to the bottle page
- Added a basic data catalog
- *(handlers)* Added GetBottle handler
- *(handler)* Added a GetManifest handler
- *(handlers)* Trying to get the if statements around the write statements working.
- Added ace hub image to the pipeline

### 🐛 Bug Fixes

- *(make)* Found out the sqlite extention works better than the dump command.
- *(handlers)* Cleaned up location handler's join
- *(handlers)* Added verification of data for location handler
- Lint
- More lint issues
- *(handlers)* Increase security of GetBlob
- Lint
- Bottles.html was broken because we changed the field name
- *(handlers)* Added some extra error handling into GetBlob handler
- Switch to StatusFound instead of StatusMovedPermenately
- AutoMigrate and switched to URL dsns
- Upsert for postgres and removed the hack
- Pretty format bottle JSON
- Fixed label link and metrics view
- Helm and skaffold now work

### 🧪 Testing

- *(test.sql)* Added an example of how LENGTH works.

### Co-authored-by

- Jon Roeber <jroeber@users.noreply.github.com>

### Refact

- Moved templating and upload code.

