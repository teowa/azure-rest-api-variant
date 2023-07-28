# azure-rest-api-variants

All variants from https://github.com/azure/azure-rest-api-specs

## Usage
```
azure-rest-api-variants                                                               
NAME:
   azure-rest-api-variants - Variants of azure-rest-api-specs

USAGE:
   azure-rest-api-index <command> [option]

VERSION:
   dev

COMMANDS:
   build    Building the variant index
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

```

e.g.,
```sh
azure-rest-api-variants build -o ./variants.json ../azure-rest-api-specs/specification
```

## Result

The result is a json file, which is a map from the parent spec to its variants, see `variants.json` for example.
