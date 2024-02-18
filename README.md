# MIB (Manager image builder)

`mib` build multiple images with dependencies with other images.

A multi-build is detect by change of a commit.

`mib` generate also a document with data specified in each directory.

## Use case

`mib` is very useful when you store all you Dockerfile in the same folder like this :
```
.
├── debian
│   └── v1
│       │── Dockerfile
│       └── bin
├── apline
│   └── v1
│       │── Dockerfile
│       └── config/
│── debian-nginx
│   └── v1
│       │── Dockerfile
│       └── config/
└── debian-nginx-php
    └── v1
        │── Dockerfile      
        └── php/
Or
.
├── debian
│   │── Dockerfile
│   └── bin
├── apline
│   │── Dockerfile
│   └── config/
│── debian-nginx
│   │── Dockerfile
│   └── config/
└── debian-nginx-php
    │── Dockerfile      
    └── php/
```

`mib` will help you to build image in the good order for build first the image parent and after children.
For example, if the image `debian` is parent of `debian-ngix` and `debian-nginx` is parent of `debian-nginx-php`, 
`mib` will build all images in this order :

    1. `debian`
    2. `debian-ngix`
    3. `debian-nginx-php`
    
`mib` works with two modes :
 * `build dirty` : get all files modified in repository (like git status), and will exclude ".md" and ".txt" extensions file and determine dependency between image  
 * `build commit [commit sha]` : get all files modified in commit change in repository, and will exclude ".md" and ".txt" extensions file and determine dependency between image 

You can also generate README.md per all images to describe image like this : 

```yaml
# Image : insanebrain/mib:v1.0.0

## Parents
- [insanebrain/alpine:v1.0.0](../../alpine/v1.0.0/README.md)
- [alpine:3.9](https://hub.docker.com/_/alpine)

## Children
- [insanebrain/mib:v1.0.0-dev](../../mib/v1.0.0-dev/README.md)

## Env Var
| Var Name | Value |
| -------- | ----- |
| GID  | Host group id (id -g)  |
| UID  | Host user id (id -u)  |

## Packages
| Var Name | Value |
| -------- | ----- |
| ca-certificates  |
| curl  |   |
| vim  |   |
| wget  |   |

## Alias
- foo/alpine:v1.0.0

File generated by mib v1.1.0 / February 07, 2019 12:44:17

```

This file will be generate with data your set in [mib.yml](doc/examples/mib.yml). This file is also used for build image.
So, if you want use `mib` you must add this file for all of your images like this :

```
.
├── debian
│   └── v1
│       │── Dockerfile
│       │── mib.yml
│       └── bin
├── apline
│   └── v1
│       │── Dockerfile
│       │── mib.yml
│       └── config/
│── debian-nginx
│   └── v1
│       │── Dockerfile
│       │── mib.yml
│       └── config/
└── debian-nginx-php
    └── v1
        │── Dockerfile
        │── mib.yml   
        └── php/
```
### Global configuration (optional)

`mib` can be configured to set log level or change default template used for README.md image.
For configure `mib` add [config.yml](doc/examples/config.yml) at the root of dir.

Currently, you can configure this parameters :

```yaml
log:
    levelStdout: "info" # trace / debug / info / warn / error / fatal / panic
    pathFile: "mib.log"
    levelFile: "info" # trace / debug / info / warn / error / fatal / panic
build:
    extensionExclude: ".md,.txt" # default extension files that will be exclude when run `build` or `generate`
template:
    imagePath: "my-custom-image.tmpl" # Define a custom template for image
    indexPath: "my-custom-index.tmpl" # Define a custom template for index
```

## Usage

```help
mib --help
usage: mib [<flags>] <command> [<args> ...]

A command-line mib helper.

Flags:
      --help           Show context-sensitive help (also try --help-long and --help-man).
  -c, --config=CONFIG  Define specific configuration file
  -p, --path="."       Define execution path

Commands:
  help [<command>...]
    Show help.

  build dirty
    Build docker images for dirty repo

  build commit <commit>
    Build docker images for specific commit

  list
    List all images of directory

  generate dirty
    Generate readme of images for dirty repo

  generate commit <commit>
    Generate readme of images for specific commit

  generate all
    Generate readme of images for all

  generate index
    Generate a readme index

  version
    Display version
```

## Requirements

* Docker (18.06+)
* Git (2.17+)

## Development

* Install go-bindata:

  ```bash
  go get -u github.com/go-bindata/go-bindata/...
  ```

* Generate assets:

  ```bash
  go generate
  ```

* Build the project:

  ```bash
  go build .
  ```

* Run tests:

  ```bash
  go test ./...
  ```

## Production

* Run mib with docker:

  ```bash
  docker run -it -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:/build insanebrain/mib:${VERSION} mib --help
  ```

* Install mib with deb (may require sudo privileges):

  ```bash
  curl -L "https://github.com/insanebrain/mib/releases/download/${VERSION}/mib_$(uname -m).deb" -o mib_$(uname -m).deb
  dpkg -i mib_$(uname -m).deb
  rm mib_$(uname -m).deb
  ```

* Install binary to custom location:

  ```bash
  curl -L "https://github.com/insanebrain/mib/releases/download/${VERSION}/mib_$(uname -s)_$(uname -m)" -o ${DESTINATION}/mib
  chmod +x ${DESTINATION}/mib
  ```

## Config file mib.yml

See documentation example for image configuration [here](doc/examples/mib.yml).

See documentation example for mib configuration [here](doc/examples/config.yml).

## Bash/ZSH Shell Completion

### Bash

Add in your ~/.bashrc :
```
eval "$(mib --completion-script-bash)"
```

### ZSH

Add in your ~/.zshrc :
```
eval "$(mib --completion-script-zsh)"
```

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md)

## License

MIT License, see [LICENSE](LICENSE.md).
