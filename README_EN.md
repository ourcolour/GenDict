# Multi-DB Data Dictionary Generator

[中文](README.md) [ENG](README_EN.md)

[![Go Version](https://img.shields.io/badge/Go-1.19%2B-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/your-username/your-repo-name/pulls)

<img src="./Icon.png" alt="Icon" style="zoom: 25%;" />

A visualized data dictionary generation tool developed using Go language, which supports multiple relational databases. This tool can automatically scan database structures, generate detailed data dictionary documents, and support multiple output formats.

![screen](screen.png)

## Functional Properties

- **Multi-Database Support**: Supports various mainstream relational databases such as MySQL, PostgreSQL, Oracle, SQLite, and more
- **Automatic Scanning**: Automatically parses database structures and extracts metadata information including tables, views, and fields (indexes are currently not supported)
- **Multiple Output Formats**: Supports exporting the data dictionary to Excel, Markdown, and other formats for easy reference and integration.

## Installation Instructions

### Prerequisites

-   **Go 1.21** or later
-   Access to the target database (e.g., MySQL, PostgresSQL)

### Installation Methods

1. **From Source**

```shell
# Clone source code
git clone https://github.com/ourcolour/GenDict.git
cd GenDict

# Compile and install
go mod download
go build -o gen-dict main.go

# Move the executable to a directory included in the PATH environment variable, for example:
sudo mv gen-dict /usr/local/bin/
```

2. **Directly download the binary**

You can download pre-compiled binaries for Windows, Linux, and macOS from the project's Release page.

3. **Supported Databases**

The current tool supports the following database types. Its metadata extraction is based on each database's information_schemaor system catalog tables.

| Database   | Supported Status | Memo    |
|------------|------------------|---------|
| MySQL      | ✅ Supported      | Default |
| PostgreSQL | ✅ Supported      | Default      |
| Oracle     | ✅ Supported      | Default      |
| SQLite     | ✅ Supported      | Default      |

## Contribution Guide

We welcome contributions of all forms! Including but not limited to:

- Reporting bugs
- Proposing new features or improvements
- Submitting pull requests

1. Fork the project
2. Create your feature branch (git checkout -b feature/AmazingFeature)
3. Commit your changes (git commit -m 'Add some AmazingFeature')
4. Push to the branch (git push origin feature/AmazingFeature)
5. Open a pull request

## License

This project is licensed under the Apache License 2.0. For more details, please refer to the [LICENSE](LICENSE) file。

## Acknowledgments

- Thanks to all the developers who have contributed to this project.

---

**Note**: When using this tool, please ensure you have legal permission to access the target database and comply with relevant data security and privacy regulations.