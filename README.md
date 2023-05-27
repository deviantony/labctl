# labctl

Control your lab environment from the command line.

## Installation

````
curl https://anthony.portainer.io/bin/labctl | sudo bash
````

## Usage

### Configuration

Override default configuration:

```
export LABCTL_CONFIG=/root/workspace/labctl/data/config.yml
```

### Lab environments (flasks)

A flask represent a lab environment.

Create a new flask:

````
labctl flask create
````

List existing flasks:

````
labctl flask ls
````

Remove a flask:

````
labctl flask rm <flask-id>
````

Or remove all flasks:

````
labctl flask rm
````

Copy files into a flask:

````
labctl flask cp <flask-id> <source> <destination>
````

Exec (SSH) into a flask:

````
labctl flask exec <flask-id>
````

### Access keys and tokens (keyring)

A keyring is a collection of access keys and tokens.

*NOTE*: Keyring is not implemented yet.

Create a new keyring:

````
labctl keyring create
````

List existing keyrings:

````
labctl keyring ls
````

Remove a keyring:

````
labctl keyring rm <keyring-id>
````

### Recipes

*NOTE*: Recipes are not implemented yet.

Apply a recipe:

````
labctl recipe apply <recipe-id | recipe-name>
````
