# Translate from a Terraform module to JSON Schema

This command will read all files in a directory with `*.tf` suffix, find all variable declaration blocks, and generate a corresponding JSON Schema.

## Examples

```shell
airlock terraform input path/to/terraform/module/
```
