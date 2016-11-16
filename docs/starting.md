# Prerequisites

* Install the stack-deployment-tool binary.

# Getting Started


1. Create a stacks.yml  ( [sample](/resources/stacks.yml) )
   This configuration will allow you to capture the parameters to be provided to your CloudFormation templates.

   The YAML configruation file supports templating using [Handlebars](http://handlebarsjs.com/) and the following built-in expressions:

   Supported Template Parameters:
   
   
   __Stack Output__ - use the output value from one stack; values must be double-quoted
   
      ```
      output stack="<stack name>" key="<output key to pull value from>" 
      ```
   
   * Example:
   
      ```LogGroupName: '{{output stack="SC-00000000-0000000000000000" key="NagiosLogGroupName"}}'```
   
   
   __env variables__ - use a environment value as a parameter
   
      ```
      env key="key_name" default="optional default value" 
   
      env.<KEY> - this version does not offer a default value
      ```
   
   * Example:
   
      ```
      InstanceType: '{{env key="INSTANCE_TYPE" default="t2.large"}}'
      - or -
      KeyName: '{{env.KEY_NAME}}'
      ```
   
   __timestamp__ - provide the number of seconds elapsed since January 1, 1970 UTC.
   
      ```
      timestamp 
      ```
   
   * Example:
   
      ```
      Name: mystack-{{ timestamp }}
      ```
   
   __yaml values__ - reference a value in the YAML configuration using a json pointer
   
      ```
      from_yaml 
      ```
   
   * Example:
   
      ```
      {{ from_yaml <json pointer to value> }}
      ```
   
   General YAML rules:
   
   All of these should be enclosed in a mustache style template within single-quotes for yaml
   
   i.e.   ```version: '{{env.ARTIFACT_VERSION}}'```
   
   and parenthesis for subexpression
   
   i.e.  ```{{sum 1 (sum 1 1)}}```
   

2. Create a [CloudFormation Template](https://aws.amazon.com/cloudformation/aws-cloudformation-templates/)
   
   stack-deployment-tool allows CloudFormation templates to be in json (default), [hjson](https://hjson.org/), or the [new yaml syntax](https://aws.amazon.com/blogs/aws/aws-cloudformation-update-yaml-cross-stack-references-simplified-substitution/)
   
   The CloudFormation stack file look up example:

  ```
  stacks:
    dev:
      elastic-search:
        stack_name: scytale-es-dev-{{env.STACK_VERSION}}
        
      nagios:
        template: nagios-infrastructure.yaml
  ```
  
  elastic-search would look for: elastic-search.[json | yml | yaml]
  
  nagios stack would use the template value for the Cloudformation template file: nagios-infrastructure.yaml
  

  * Include user data from separate file

   For example, a basic bash script can be stored separately and then included, as show below:

   YAML include:
   
   ```
  UserData:
        "Fn::Base64": !Sub |
            #!/usr/bin/env bash
            apt-get update -y
            ${Local::IncludeFile userdata.sh}
            sleep 10
   ```
  
   JSON include:
  
   ```
  "UserData": { "Fn::Base64": { "Fn::Join": [ "", [
          "#!/usr/bin/env bash\n",
          "apt-get update -y\n",

          {"Fn::Local::IncludeFileLines" : "userdata.sh" },
          
          "sleep 10\n",
        ]]}}
   ```

3. Version you application

  ```
  stack-deployment-tool versions init # sets version to 0.0.0

  # you can bump it to: 0.1.0 when you're ready
  stack-deployment-tool versions bump --minor
  ```

3. Scenarios

  
  * Deploying stack

  ```
  # create the build version (optionally) information
  stack-deployment-tool versions build

  # publish the artifact archive created above to the "sandbox" repo
  stack-deployment-tool artifacts upload <the file> -s resources/stacks.yml -b mybucket

  # deploy the stacks!
  stack-deployment-tool stacks deploy resources/bluegreen/stack_blue.yaml -s dev-techops.blue
  ```

  * Promotion

  The promotion path for artifacts is: sandbox -> snapshot -> staging -> release

  ```
  stack-deployment-tool artifacts promote count.out sandbox -s resources/stacks.yml -b mybucket
  ```

  * Deploying all defined stacks (for dev)

    ```
    stack-deployment-tool versions build
    stack-deployment-tool stacks deploy resources/bluegreen/stack_blue.yaml -s dev
    ```

  * Deploying one application stack  

    ```
    stack-deployment-tool versions build
    stack-deployment-tool stacks deploy resources/bluegreen/stack_blue.yaml -s dev-techops.blue
    ```


## Command Details

### Deploying a stack

This command deploys stack(s) specified via the *--stacks* parameter. This parameter can specify one or more environment/stack combination,
for example:

```
qa.app
qa[app,elb]
```

The deployment updates the stack if one already exists with the same name, or creates a new one otherwise.

``` bash
export AWS_ROLE_ARN=arn:aws:iam::01234:role/Developer # optional
export AWS_PROFILE=Developer

sdt stacks deploy stacks.yml --stacks dev.drone-ecs
```

### Teardown

This command deletes the specified stack(s). Typically this is useful for build/dev environments, where stack only needs to be live for the duration of a test.

``` bash
export AWS_ROLE_ARN=arn:aws:iam::01234:role/Developer # optional
export AWS_PROFILE=Developer

sdt stacks destroy stacks.yml --stacks dev.drone-ecs
```



### Environment Variables

stack-deployment-tool respects the following environment variables:

__HTTP_PROXY__

__AWS_PROFILE__

__AWS_ROLE_ARN__

ex.
```
HTTP_PROXY=http://something AWS_PROFILE=Developer AWS_ROLE_ARN="arn:aws:iam::xxx" ./sdt <stuff>
```

