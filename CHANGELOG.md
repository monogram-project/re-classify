# Change Log for Monogram-Go

Following the style in https://keepachangelog.com/en/1.0.0/

## v0.2.1, Bracket handling 

### Added

- New command-line option `--echo-to-stderr` for repeating the classification 
  strings to stderr.
- New configuration option `bracket-regexp` for classifying brackets.

## v0.1.3, First stable version

This is the first stable release of re-classify, a token classification tool
that implements the classification protocol for Monogram.

### Added
- Complete implementation of the classification protocol
- Support for surround patterns (form-start/form-end matching)
- Form prefix pattern classification
- Simple and compound label pattern support  
- Operator classification with precedence values
- Variable token classification
- Comprehensive configuration system via YAML files
- Command-line interface for processing token streams
- Documentation for configuration format and classification protocol
