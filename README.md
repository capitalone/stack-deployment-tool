[![Go Report Card](https://goreportcard.com/badge/github.com/capitalone/stack-deployment-tool)](https://goreportcard.com/report/github.com/capitalone/stack-deployment-tool)

# Stack Deployment Tool

Tool that codeifies packaging & delivery best-practices. 
It extends stack configuration with local templating and associating multiple [CloudFormation](https://aws.amazon.com/cloudformation/) stacks for lifecycle management: deployment, updating, and destroying. 


## [stacks.yml](resources/stacks.yml)

Provides Enhanced Capabilities to CloudFormation Stacks:

* Ability to capture parameters for a CloudFormation stack creation into a stacks.yml which can be checked into a vcs
* Ability to reference the output of one CloudFormation stack in the creation of a new CloudFormation stack
* Ability to reference environment variables in a template
* Ability to store user data scripts separately and include them in a CloudFormation template
* Stack configuration (stacks.yml) utilizes [handlebars](http://handlebarsjs.com/) for templatizing 


## Features

* CloudFormation
 * Multiple Stack definitions in a single file (stacks.yml)
  * Holds parameters & tags for stacks for multiple stacks & multiple environments
  * Carry stack outputs as input parameters to new stack creations
 * Stack create, update, destroy
 * Include separate user data script files

* Versioning
 * version tracking for projects by way of a 'version.properties'
 * support bumping major,minor,patch
 * add metadata from git commit
 * add build data from environment vars (BUILD_NUMBER) or user-timestamp

* Artifacts
 * upload, promotion, and download of artifacts to s3 buckets and/or nexus repositories


# [Getting Started](docs/starting.md)

[Guide](docs/starting.md)

[Dev from source](docs/dev.md)

## Command Line Help

```
$ ./build/darwin_amd64/sdt 
Stack Deployment Tool
	that will help with deploying multiple CloudFormation stacks

Usage:
  stack-deployment-tool [command]

Available Commands:
  artifacts   artifacts functions, like finding uploading, downloading, promoting
  stacks      Stack manipulation commands
  version     Print the version number of Hugo
  versions    versioning commands

Flags:
      --config string   config file (default is $HOME/.stack-deployment-tool.yaml)
  -d, --debug           enable debug
  -q, --drymode         enable dry mode
  -h, --help            help for stack-deployment-tool
  -t, --toggle          Help message for toggle

Use "stack-deployment-tool [command] --help" for more information about a command.

```

Example Stack Deploy:

```
$ sdt stacks deploy resources/bluegreen/stack_blue.yaml -s dev.bluegreen-2

INFO[0018] Waiting for stack operation to complete      
+-----------------------------------------------+--------------------------------+--------------------------------+
|                                        Status |                           Type |                      LogicalID |
+-----------------------------------------------+--------------------------------+--------------------------------+
|                            CREATE_IN_PROGRESS |     AWS::CloudFormation::Stack |                      bluegreen |
|                            CREATE_IN_PROGRESS | AWS::CloudFormation::WaitCo... |               DeployWaitHandle |
|                            CREATE_IN_PROGRESS | AWS::ElasticLoadBalancing::... |                   LoadBalancer |
|                            CREATE_IN_PROGRESS | AWS::CloudFormation::WaitCo... |               DeployWaitHandle |
|                               CREATE_COMPLETE | AWS::CloudFormation::WaitCo... |               DeployWaitHandle |
|                            CREATE_IN_PROGRESS | AWS::ElasticLoadBalancing::... |                   LoadBalancer |
|                               CREATE_COMPLETE | AWS::ElasticLoadBalancing::... |                   LoadBalancer |
|                            CREATE_IN_PROGRESS | AWS::CloudFormation::WaitCo... |            DeployWaitCondition |
|                            CREATE_IN_PROGRESS | AWS::CloudFormation::WaitCo... |            DeployWaitCondition |
|                            CREATE_IN_PROGRESS | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                            CREATE_IN_PROGRESS | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                               CREATE_COMPLETE | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                            CREATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                            CREATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                               CREATE_COMPLETE | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                               CREATE_COMPLETE | AWS::CloudFormation::WaitCo... |            DeployWaitCondition |
|                               CREATE_COMPLETE |     AWS::CloudFormation::Stack |                      bluegreen |
|                            UPDATE_IN_PROGRESS |     AWS::CloudFormation::Stack |                      bluegreen |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                                 UPDATE_FAILED | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                   UPDATE_ROLLBACK_IN_PROGRESS |     AWS::CloudFormation::Stack |                      bluegreen |
|                               UPDATE_COMPLETE | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|  UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS |     AWS::CloudFormation::Stack |                      bluegreen |
|                               DELETE_COMPLETE | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                      UPDATE_ROLLBACK_COMPLETE |     AWS::CloudFormation::Stack |                      bluegreen |
|                            UPDATE_IN_PROGRESS |     AWS::CloudFormation::Stack |                      bluegreen |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                               UPDATE_COMPLETE | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                            UPDATE_IN_PROGRESS | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|                               UPDATE_COMPLETE | AWS::AutoScaling::AutoScali... |                   BlueGreenAsg |
|           UPDATE_COMPLETE_CLEANUP_IN_PROGRESS |     AWS::CloudFormation::Stack |                      bluegreen |
|                            DELETE_IN_PROGRESS | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                               DELETE_COMPLETE | AWS::AutoScaling::LaunchCon... |                   LaunchConfig |
|                               UPDATE_COMPLETE |     AWS::CloudFormation::Stack |                      bluegreen |
+-----------------------------------------------+--------------------------------+--------------------------------+
INFO[0600] Stacks Create Complete                       
```

# Roadmap



## Contributors
We welcome your interest in Capital One’s Open Source Projects (the “Project”). Any Contributor to the Project must accept and sign a CLA indicating agreement to the license terms. Except for the license granted in this CLA to Capital One and to recipients of software distributed by Capital One, You reserve all right, title, and interest in and to your Contributions; this CLA does not impact your rights to use your own contributions for any other purpose.

##### [Link to CLA] (https://docs.google.com/forms/d/19LpBBjykHPox18vrZvBbZUcK6gQTj7qv1O5hCduAZFU/viewform)

This project adheres to the [Open Source Code of Conduct][code-of-conduct]. By participating, you are expected to honor this code.

[code-of-conduct]: https://developer.capitalone.com/single/code-of-conduct/

### Contribution Guidelines
We encourage any contributions that align with the intent of this project and add more functionality or languages that other developers can make use of. To contribute to the project, please submit a PR for our review. Before contributing any source code, familiarize yourself with the Apache License 2.0 (license.md), which controls the licensing for this project.

