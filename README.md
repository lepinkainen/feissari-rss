# feissari-rss

Make the feissarimokat.com RSS feed usable in a feed reader

## Description

feissari-rss is a tool that enhances the RSS feed from feissarimokat.com by fetching and embedding images directly into the feed. This makes the feed more usable in standard RSS readers by displaying images that would otherwise be missing.

## Installation

### Building from source

This project uses [Task](https://taskfile.dev/) for build automation. Make sure you have Go and Task installed.

```bash
# Build for your current platform
task build

# Cross-compile for Linux AMD64
task build-linux
```

The compiled binary will be available in the `build` directory.

## Usage

```bash
# Basic usage (outputs to current directory)
./build/feissari-rss

# Specify an output directory
./build/feissari-rss -outdir /path/to/output
```
