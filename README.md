# Specter

[![Go](https://github.com/Morebec/specter/actions/workflows/go.yml/badge.svg)](https://github.com/Morebec/specter/actions/workflows/go.yml)

Specter is a Go library designed to help developers easily build declarative DSLs (Domain-Specific Languages) and 
process them through an extensible pipeline. 

It is currently used at [Morébec](https://morebec.com) for generating microservice APIs, code and documentation, 
managing configurations, automating deployments, and so many other fun things. 

Specter provides a simple yet powerful framework to simplify these workflows.

The library also comes with many batteries included for common tasks such as dependency resolution 
and linting, HCL configuration loading and more.

## Key Use Cases

At [Morébec](https://morebec.com) Specter is primarily used to create high-level, syntax-consistent DSLs for tools 
like OpenAPI, Docker/Docker Compose, and Terraform. 

Here are some of the key use cases Specter powers for us internally:

- **Code Generation:** We generate entire code bases in PHP and Go leveraging DDD/CQRS/ES in a low-code manner to focus on business logic and 
reduce plumbing work.
- **Enforce Coding Standards**: We ensure consistency and improve development speed by automating code quality checks and 
  standardization.
- **Configuration Management:** We use it to manage environment-specific configuration files, such as Docker or 
  Kubernetes manifests, using declarative Units.
- **CI/CD Automation:** Automate the generation of CI/CD pipeline definitions (Jenkins, GitHub Actions, etc.) 
  by processing high-level declarative Units.
- **Infrastructure as Code:** Describe infrastructure components declaratively and generate Terraform, 
  scripts, or other IAC artifacts.


## How Specter Works
Specter is based around a simple yet powerful *pipeline* architecture. The core of the pipeline is designed to process 
*Units* — which are declarative components that represent different aspects or concepts — and produce various types 
of outputs based on them called *artifacts*.

In short, specter loads Units which it processes before outputting corresponding artifacts.

For example, in the case of our Go code generator, we first define Microservices with their Commands, Events
and Queries in specification files that are then processed by Specter and transformed into their 
corresponding Go implementation along with a changelog, markdown documentation and OpenAPI specification.

In this example, the Microservice/Command/Event/Query definition files are the "Units", while the
generated code, markdown documentation, changelog, and OpenAPI are the "artifacts".

Units are anything that needs transforming, and artifacts are anything these units can be transformed into.

To illustrate, here's an example of a Unit File that could describe a docker container to be deployed on a 
given host using an HCL syntax:

```
service "web" {
  image = "our-app:latest"
  ports = ["8080:80"]
  volumes = [
    {
      type = "bind"
      source = "./html"
      target = "/usr/share/nginx/html"
    }
  ]
  deploymentHost = "primary-us-east-dev"
}
```

### Pipeline Stages
The pipeline consists of several stages, each responsible for a specific task in the workflow. 
Here's an overview of the stages and the concepts they introduce:

### 1. Source Loading
The very first step is to acquire these units. Depending on the use cases these units could come from files, HTTP resources,
or even Database rows. These different locations are known in Specter as Unit Sources.

As such, the Source Loading stage corresponds to loading these sources so that they can be acquired/fetched 
and read.

- Inputs: Source locations
- Outputs: (Loaded) Sources

### 2. Unit Loading
Units are read and materialized into in-memory data structures. This stage converts raw source data into 
usable Units that can be processed according to your specific needs.

- Inputs: Sources
- Outputs: (Loaded) Units

### 3. Unit Processing
Once the Units are loaded, Specter applies processors which are the core services responsible for generating artifacts
based on these units. These processors can do things like validate the Units, resolve dependencies, or convert them 
into different representations. You can easily add custom processors to extend Specter's behavior. 

The artifacts are in-memory representations of the actual desired outputs. For instance, the FileArtifact represents
a file to be outputted.

- Inputs: Sources
- Outputs: Artifacts

### 4. Artifact Processing
The final stage of the pipeline processes artifacts that were generated during the previous step.
The processing of these artifacts can greatly vary based on the types of artifacts at play.
An artifact could be anything from a file, an API call, to a database insertion or update query, 
to a command or program to be executed.

- Inputs: Artifacts
- Outputs: Final outputs (files, API calls, etc.)

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
