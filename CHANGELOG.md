# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Unit tests for `flow` package.
### Changed
- Updated Postman collection.
- Using read/write mutex to Metrics map. Before this change, it was not safe for concurrent use.
- Resolved bug that resulted in requested HTTP header values to not be properly omitted when being reported to Influxdb.
- Travis will now retry the goveralls step if not successful the first time.
- Docker images will now only be built and pushed on pushes to the master branch. This resolves an issue where images were being built and pushed on pull requests where the master branch was the target.

## [1.0.0] - 2017-07-31
### Added
- Initial Project Commit.