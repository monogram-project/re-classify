# re-classify Tool

The `re-classify` tool is a command-line utility that classifies tokens using on
a configuration file based on regular expression patterns. It is designed to
implement Monogram's stripped down [external classification protocol](docs/classification-protocol.md), in which
tokens are read, one per line, from the standard input and their 1-line
classification is written to the standard output.

**Useful tips**: The number of output lines will always match the number
of input lines. You will need to provide the whole of the input before any
output is generated.


## Installation

### From Release (Recommended)

Download the latest release for your platform from the [GitHub Releases page](https://github.com/monogram-project/re-classify/releases).

### From Source

The easiest way to install from source is via the `go install` command:
```bash
go install github.com/monogram-project/re-classify/cmd/re-classify@latest
```

Alternatively, you can clone and build locally:
```bash
git clone https://github.com/monogram-project/re-classify.git
cd re-classify
just install
```

### Using Docker

```bash
# Pull the latest image
docker pull sfkleach/re-classify:latest

# Run with a config file
docker run --rm -i sfkleach/re-classify:latest ./test-configs/example-config.yaml < input.txt
```

## Building

This project uses [Just](https://github.com/casey/just) as a command runner. To see available commands:

```bash
just
```

Common commands:
```bash
just build      # Build the binary
just test       # Run tests
just check      # Run all checks (fmt, vet, test)
just clean      # Clean build artifacts
```

## Project Structure

This project follows standard Go conventions:

```
re-classify/
├── .github/
│   └── workflows/            # GitHub Actions workflows
│       ├── build-and-test.yml
│       └── release.yml
├── cmd/
│   └── re-classify/          # Main application
│       └── main.go
├── internal/                 # Private application and library code
│   ├── classifier/           # Token classification logic
│   │   └── classifier.go
│   └── config/               # Configuration handling
│       └── config.go
├── test-configs/             # Example configuration files
│   ├── config.yaml
│   ├── example-config.yaml
│   └── ...
├── .goreleaser.yml           # GoReleaser configuration
├── Dockerfile                # Container build definition
├── Justfile                  # Command runner (replaces Makefile)
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
└── README.md                 # This file
```

## Usage

The `re-classify` command takes a single argument, which is the name of a
configuration file. For details on the configuration file format, see
[`docs/configuration-format.md`](docs/configuration-format.md).

```bash
re-classify [OPTIONS] FILE < STDIN > STDOUT
```

The supported options are `--version` and `--check`. The `--version` option is
self-explanatory. The `--check` option verifies the syntax of the configuration
file and exits.


## Classification Protocol

The tool outputs 1-line classifications for each token. These consist of
a single letter optionally followed by additional, whitespace-separated data.
The 1-letter codes are:

- `S` - Start token (form start, e.g., `def`, `if`, `while`)
- `E` - End token (form end, e.g., `end`, `endif`, `endwhile`)
- `C` - Compound token (multi-part constructs)
- `L` - Label token (identifiers used as labels)
- `P` - Prefix token (operators that come before their operand)
- `O` - Operator token (infix, postfix operators)
- `V` - Variable token (default for unclassified identifiers)

For start tokens, the output may include expected end tokens:
- e.g. `S enddef endfunction` - Start token with possible endings

And operator tokens, the output will include three space separated precedence values:
- e.g. `O 10 100 5`

For more details read [the external classification protocol](docs/classification-protocol.md).