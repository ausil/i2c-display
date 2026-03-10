# License Summary

This project is licensed under the BSD 3-Clause License. All dependencies are compatible with this license.

## Direct Dependencies

### 1. periph.io packages
- **Packages**: `periph.io/x/conn/v3`, `periph.io/x/devices/v3`, `periph.io/x/host/v3`
- **License**: Apache License 2.0
- **Compatibility**: ✅ Compatible with BSD 3-Clause

### 2. github.com/prometheus/client_golang
- **Package**: `github.com/prometheus/client_golang`
- **License**: Apache License 2.0
- **Compatibility**: ✅ Compatible with BSD 3-Clause

### 3. github.com/rs/zerolog
- **Package**: `github.com/rs/zerolog`
- **License**: MIT License
- **Compatibility**: ✅ Compatible with BSD 3-Clause

### 4. golang.org/x/image
- **Package**: `golang.org/x/image`
- **License**: BSD 3-Clause
- **Compatibility**: ✅ Compatible (same license)

### 5. Go Standard Library
- **License**: BSD 3-Clause
- **Compatibility**: ✅ Compatible (same license)

## Indirect Dependencies

| Package | License |
|---------|---------|
| `github.com/beorn7/perks` | MIT |
| `github.com/cespare/xxhash/v2` | MIT |
| `github.com/mattn/go-colorable` | MIT |
| `github.com/mattn/go-isatty` | MIT |
| `github.com/munnerz/goautoneg` | BSD 3-Clause |
| `github.com/prometheus/client_model` | Apache 2.0 |
| `github.com/prometheus/common` | Apache 2.0 |
| `github.com/prometheus/procfs` | Apache 2.0 |
| `go.yaml.in/yaml/v2` | Apache 2.0 |
| `golang.org/x/sys` | BSD 3-Clause |
| `google.golang.org/protobuf` | BSD 3-Clause |

## License Compatibility Overview

The BSD 3-Clause license is a permissive open-source license that is compatible with most other permissive licenses, including:

- ✅ Apache License 2.0
- ✅ MIT License
- ✅ BSD 2-Clause License
- ✅ BSD 3-Clause License

According to the [Apache Software Foundation](https://www.apache.org/legal/resolved.html) and [various open source license guides](https://ghinda.com/blog/opensource/2020/open-source-licenses-apache-mit-bsd.html), permissive licenses like Apache 2.0 and BSD are compatible because they have minimal restrictions on redistribution and use.

## Dependency License Files

License files for all dependencies are included in the `vendor/` directory alongside the vendored source code.

## Compliance

This project complies with all dependency licenses by:

1. Retaining all copyright notices from dependencies
2. Including this license summary document
3. Not modifying the original license terms of dependencies
4. Distributing the software under BSD 3-Clause terms

## References

- [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)
- [BSD 3-Clause License](https://opensource.org/licenses/BSD-3-Clause)
- [License Compatibility Guide](https://www.oreilly.com/library/view/understanding-open-source/0596005814/ch02.html)
- [ASF Third Party License Policy](https://www.apache.org/legal/resolved.html)
