# Stacks - working with CloudFormation

# Getting Started

## Deployment descriptor format



## Supported commands

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

