# labctl

Control your lab environment from the command line.

## Installation

````
curl https://anthony.portainer.io/bin/labctl | sudo bash
````

## Usage

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

### Recipes

Apply a recipe:

````
labctl recipe apply <recipe-id | recipe-name>
````
