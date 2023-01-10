# Specter

Specter is a development toolkit in Go that allows you to develop configuration file processors based on 
HashiCorp Configuration Language (HCL). With Specter, you can define your own Domain-Specific Language (DSL) 
using HCL and create a processing pipeline to validate, lint, resolve dependencies, and generate code or output 
files from these DSL configuration files.

## Features

- Develop your own DSL using HCL
- Validate and lint configuration files
- Resolve dependencies between configuration files
- Generate code or output files from configuration files

## Getting Started

To start using Specter, you need to install Go and set up your Go workspace. 
Then, you can install Specter using the following command:

```bash
go get github.com/morebec/specter
```

Next, you can create a new configuration file processor by defining your DSL in HCL and implementing the processing 
pipeline. You can find more detailed instructions and examples in the [documentation](https://morebec.github.io/specter).

## Examples

Here are some examples of what you can do with Specter:

- [Configuration file generator](https://github.com/morebec/specter-example-config-generator)
- [Code generator](https://github.com/morebec/specter-example-code-generator)

## Contributions

We welcome contributions to Specter! If you have an idea for a new feature or have found a bug, please open an issue to 
discuss it. If you want to contribute code, please follow 
our [contribution guidelines](https://morebec.github.io/specter/contributing) and open a pull request.

## License

Specter is licensed under the [MIT License](LICENSE).