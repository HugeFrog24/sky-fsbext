# Changelog

All notable changes to this project will be documented in this file.

## [1.0.10] - _(2025-05-24)_

- ### Changed
  - Updated Go dependencies and tooling
  
- ### Fixed
  - Improved string handling in Windows
  - Enhanced error handling for file operations


## [1.0.3] - _(2024-02-03)_

- ### Added
  - Enhanced `check_disk_space` to support mock testing and added string conversion in subprocess calls.

- ### Changed
  - Refactored disk space check to target the output directory and improved path resolution for `vgmstream-cli`.

- ### Fixed
  - Improved error handling in file extraction and directory management for increased stability.

## [1.0.2] - _(2023-09-30)_

- ### Added
  - Implemented dynamic disk space check based on a compression ratio.
  - Improved code documentation with inline comments for better readability.

- ### Fixed
  - Corrected path handling to ensure proper file and directory management, resolving potential double directory issues (e.g., './in/in/').

## [1.0.1] - _(2023-03-17)_

- ### Added
  - More flexibility, advanced logging and command-line arguments

- ### Changed
  - Suppress the verbose output of vgmstream-cli by default

- ### Fixed
  - Made the script more robust in general

## [1.0.0] - _(2023-03-17)_

- Initial release
