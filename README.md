<img src="images/logo.png" width="400" height="400">

# AIRLOCK

[![Go Report Card](https://goreportcard.com/badge/github.com/massdriver-cloud/airlock)](https://goreportcard.com/report/github.com/massdriver-cloud/airlock)
[![License](https://img.shields.io/github/license/massdriver-cloud/airlock)](https://github.com/massdriver-cloud/airlock/blob/master/LICENSE)

## Overview

Translate between JSON Schema and common IaC languages (opentofu, helm, bicep)

## Getting Started

### Prerequisites

- Go version 1.22

### Installation

To install this package, simply run:

```bash
go get -u github.com/massdriver-cloud/airlock
```

### Usage

#### OpenTofu

OpenTofu -> JSON Schema:

```bash
airlock opentofu input /path/to/module
```

<details>
    <summary>Example</summary>

`main.tf`:

```terraform
provider "aws" {
  region = var.region
}

variable "bucket_name" {
  description = "The name of the S3 bucket."
  type        = string
}

variable "region" {
  description = "The AWS region to create the S3 bucket in."
  type        = string
  default     = "us-east-1"
}

variable "enable_versioning" {
  description = "Enable versioning on the S3 bucket."
  type        = bool
  default     = false
}

variable "acl" {
  description = "The access control list for the S3 bucket."
  type        = string
  default     = "private"
}

resource "aws_s3_bucket" "example" {
  bucket = var.bucket_name
  acl    = var.acl

  versioning {
    enabled = var.enable_versioning
  }

  lifecycle_rule {
    id      = "delete-old-versions"
    enabled = true

    noncurrent_version_expiration {
      days = 30
    }
  }

  tags = {
    Name        = var.bucket_name
    Environment = "Dev"
  }
}

output "bucket_id" {
  value = aws_s3_bucket.example.id
}

output "bucket_arn" {
  value = aws_s3_bucket.example.arn
}
```

JSON output:

```json
{
  "properties": {
    "bucket_name": {
      "type": "string",
      "title": "bucket_name",
      "description": "The name of the S3 bucket."
    },
    "region": {
      "type": "string",
      "title": "region",
      "description": "The AWS region to create the S3 bucket in.",
      "default": "us-east-1"
    },
    "enable_versioning": {
      "type": "boolean",
      "title": "enable_versioning",
      "description": "Enable versioning on the S3 bucket.",
      "default": false
    },
    "acl": {
      "type": "string",
      "title": "acl",
      "description": "The access control list for the S3 bucket.",
      "default": "private"
    }
  },
  "required": [
    "acl",
    "bucket_name",
    "enable_versioning",
    "region"
  ]
}
```

</details>

JSON Schema -> OpenTofu:

```bash
airlock opentofu output /path/to/schema.json
```

<details>
    <summary>Example</summary>

`schema.json`:

```json
{
  "properties": {
    "form": {
      "title": "Form",
      "type": "object",
      "required": [
        "firstName",
        "lastName"
      ],
      "properties": {
        "firstName": {
          "type": "string",
          "title": "First name"
        },
        "lastName": {
          "type": "string",
          "title": "Last name"
        },
        "age": {
          "type": "integer",
          "title": "Age"
        },
        "bio": {
          "type": "string",
          "title": "Bio"
        },
        "password": {
          "type": "string",
          "title": "Password",
          "minLength": 3
        },
        "telephone": {
          "type": "string",
          "title": "Telephone",
          "minLength": 10
        }
      }
    }
  }
}
```

OpenTofu output:

```terraform
variable "form" {
  type = object({
    firstName = string
    lastName  = string
    age       = optional(number)
    bio       = optional(string)
    password  = optional(string)
    telephone = optional(string)
  })
  default = null
}
```

</details>

#### Helm

Helm -> JSON Schema:

```bash
airlock helm input /path/to/values.yaml
```

<details>
  <summary>Example</summary>

`values.yaml`:

```yaml
name: my-app

image:
  repository: my-app-repo/my-app
  tag: latest
  pullPolicy: IfNotPresent

service:
  enabled: true
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: false
  path: /
  hosts:
    - host: my-app.local
      paths:
        - /
  tls: []

replicaCount: 1

resources:
  requests:
    cpu: "100m"
    memory: "256Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"

env:
  - name: DATABASE_URL
    value: postgres://user:password@postgres:5432/mydb
  - name: APP_SECRET
    value: supersecretkey

logging:
  level: info

persistence:
  enabled: false
  storageClass: "standard"
  accessModes:
    - ReadWriteOnce
  size: 1Gi

annotations: {}

nodeSelector: {}

tolerations: []

affinity: {}
```

JSON Schema output:

```json
{
  "properties": {
    "name": {
      "type": "string",
      "title": "name",
      "description": "Application name",
      "default": "my-app"
    },
    "image": {
      "properties": {
        "repository": {
          "type": "string",
          "title": "repository",
          "default": "my-app-repo/my-app"
        },
        "tag": {
          "type": "string",
          "title": "tag",
          "default": "latest"
        },
        "pullPolicy": {
          "type": "string",
          "title": "pullPolicy",
          "default": "IfNotPresent"
        }
      },
      "type": "object",
      "required": [
        "repository",
        "tag",
        "pullPolicy"
      ],
      "title": "image",
      "description": "Image configuration"
    },
    "service": {
      "properties": {
        "enabled": {
          "type": "boolean",
          "title": "enabled",
          "default": true
        },
        "type": {
          "type": "string",
          "title": "type",
          "default": "ClusterIP"
        },
        "port": {
          "type": "integer",
          "title": "port",
          "default": 80
        },
        "targetPort": {
          "type": "integer",
          "title": "targetPort",
          "default": 8080
        }
      },
      "type": "object",
      "required": [
        "enabled",
        "type",
        "port",
        "targetPort"
      ],
      "title": "service",
      "description": "Service configuration"
    },
    "ingress": {
      "properties": {
        "enabled": {
          "type": "boolean",
          "title": "enabled",
          "default": false
        },
        "path": {
          "type": "string",
          "title": "path",
          "default": "/"
        },
        "hosts": {
          "items": {
            "properties": {
              "host": {
                "type": "string",
                "title": "host",
                "default": "my-app.local"
              },
              "paths": {
                "items": {
                  "type": "string",
                  "default": "/"
                },
                "type": "array",
                "title": "paths",
                "default": [
                  "/"
                ]
              }
            },
            "type": "object",
            "required": [
              "host",
              "paths"
            ]
          },
          "type": "array",
          "title": "hosts",
          "default": [
            {
              "host": "my-app.local",
              "paths": [
                "/"
              ]
            }
          ]
        },
        "tls": {
          "items": {
            "properties": {
              "secretName": {
                "type": "string",
                "title": "secretName",
                "default": "my-app-tls"
              },
              "hosts": {
                "items": {
                  "type": "string",
                  "default": "my-app.local"
                },
                "type": "array",
                "title": "hosts",
                "default": [
                  "my-app.local"
                ]
              }
            },
            "type": "object",
            "required": [
              "secretName",
              "hosts"
            ]
          },
          "type": "array",
          "title": "tls",
          "default": [
            {
              "hosts": [
                "my-app.local"
              ],
              "secretName": "my-app-tls"
            }
          ]
        }
      },
      "type": "object",
      "required": [
        "enabled",
        "path",
        "hosts",
        "tls"
      ],
      "title": "ingress",
      "description": "Ingress configuration"
    },
    "replicaCount": {
      "type": "integer",
      "title": "replicaCount",
      "description": "Replicas configuration",
      "default": 1
    },
    "resources": {
      "properties": {
        "requests": {
          "properties": {
            "cpu": {
              "type": "string",
              "title": "cpu",
              "default": "100m"
            },
            "memory": {
              "type": "string",
              "title": "memory",
              "default": "256Mi"
            }
          },
          "type": "object",
          "required": [
            "cpu",
            "memory"
          ],
          "title": "requests"
        },
        "limits": {
          "properties": {
            "cpu": {
              "type": "string",
              "title": "cpu",
              "default": "500m"
            },
            "memory": {
              "type": "string",
              "title": "memory",
              "default": "512Mi"
            }
          },
          "type": "object",
          "required": [
            "cpu",
            "memory"
          ],
          "title": "limits"
        }
      },
      "type": "object",
      "required": [
        "requests",
        "limits"
      ],
      "title": "resources",
      "description": "Resource requests and limits"
    },
    "env": {
      "items": {
        "properties": {
          "name": {
            "type": "string",
            "title": "name",
            "default": "DATABASE_URL"
          },
          "value": {
            "type": "string",
            "title": "value",
            "default": "postgres://user:password@postgres:5432/mydb"
          }
        },
        "type": "object",
        "required": [
          "name",
          "value"
        ]
      },
      "type": "array",
      "title": "env",
      "description": "Environment variables",
      "default": [
        {
          "name": "DATABASE_URL",
          "value": "postgres://user:password@postgres:5432/mydb"
        },
        {
          "name": "APP_SECRET",
          "value": "supersecretkey"
        }
      ]
    },
    "logging": {
      "properties": {
        "level": {
          "type": "string",
          "title": "level",
          "default": "info"
        }
      },
      "type": "object",
      "required": [
        "level"
      ],
      "title": "logging",
      "description": "Logging configuration"
    },
    "persistence": {
      "properties": {
        "enabled": {
          "type": "boolean",
          "title": "enabled",
          "default": false
        },
        "storageClass": {
          "type": "string",
          "title": "storageClass",
          "default": "standard"
        },
        "accessModes": {
          "items": {
            "type": "string",
            "default": "ReadWriteOnce"
          },
          "type": "array",
          "title": "accessModes",
          "default": [
            "ReadWriteOnce"
          ]
        },
        "size": {
          "type": "string",
          "title": "size",
          "default": "1Gi"
        }
      },
      "type": "object",
      "required": [
        "enabled",
        "storageClass",
        "accessModes",
        "size"
      ],
      "title": "persistence",
      "description": "Persistent storage configuration"
    },
    "annotations": {
      "properties": {},
      "type": "object",
      "title": "annotations",
      "description": "Custom annotations"
    },
    "nodeSelector": {
      "properties": {},
      "type": "object",
      "title": "nodeSelector",
      "description": "Node selector"
    },
    "tolerations": {
      "items": {
        "properties": {
          "key": {
            "type": "string",
            "title": "key",
            "default": "key1"
          },
          "operator": {
            "type": "string",
            "title": "operator",
            "default": "Equal"
          },
          "value": {
            "type": "string",
            "title": "value",
            "default": "value1"
          },
          "effect": {
            "type": "string",
            "title": "effect",
            "default": "NoSchedule"
          }
        },
        "type": "object",
        "required": [
          "key",
          "operator",
          "value",
          "effect"
        ]
      },
      "type": "array",
      "title": "tolerations",
      "description": "Tolerations",
      "default": [
        {
          "effect": "NoSchedule",
          "key": "key1",
          "operator": "Equal",
          "value": "value1"
        },
        {
          "effect": "NoExecute",
          "key": "key2",
          "operator": "Exists"
        }
      ]
    },
    "affinity": {
      "properties": {},
      "type": "object",
      "title": "affinity",
      "description": "Affinity settings"
    }
  },
  "type": "object",
  "required": [
    "name",
    "image",
    "service",
    "ingress",
    "replicaCount",
    "resources",
    "env",
    "logging",
    "persistence",
    "annotations",
    "nodeSelector",
    "tolerations",
    "affinity"
  ]
}
```

</details>

#### Bicep

Bicep -> JSON Schema:

```bash
airlock bicep input /path/to/template.bicep
```

<details>
  <summary>Example</summary>

`template.bicep`:

```bicep
@description('The name of the resource group.')
param resourceGroupName string

@description('The location where the storage account will be deployed.')
param location string = resourceGroup().location

@description('The name of the storage account.')
@secure()
param storageAccountName string

@description('The SKU for the storage account.')
@allowed([
  'Standard_LRS'
  'Standard_GRS'
  'Standard_RAGRS'
  'Standard_ZRS'
  'Premium_LRS'
])
param sku string = 'Standard_LRS'

@description('The kind of storage account.')
@allowed([
  'StorageV2'
  'Storage'
  'BlobStorage'
  'FileStorage'
  'BlockBlobStorage'
])
param kind string = 'StorageV2'

resource storageAccount 'Microsoft.Storage/storageAccounts@2021-09-01' = {
  name: storageAccountName
  location: location
  sku: {
    name: sku
  }
  kind: kind
  properties: {
    supportsHttpsTrafficOnly: true
  }
}

output storageAccountId string = storageAccount.id
output storageAccountPrimaryEndpoints object = storageAccount.properties.primaryEndpoints
```

JSON Schema output:

```json
{
  "properties": {
    "storageAccountName": {
      "type": "string",
      "format": "password",
      "title": "storageAccountName",
      "description": "The name of the storage account."
    },
    "kind": {
      "type": "string",
      "enum": [
        "StorageV2",
        "Storage",
        "BlobStorage",
        "FileStorage",
        "BlockBlobStorage"
      ],
      "title": "kind",
      "description": "The kind of storage account.",
      "default": "StorageV2"
    },
    "location": {
      "type": "string",
      "title": "location",
      "description": "The location where the storage account will be deployed.",
      "default": "[resourceGroup().location]"
    },
    "resourceGroupName": {
      "type": "string",
      "title": "resourceGroupName",
      "description": "The name of the resource group."
    },
    "sku": {
      "type": "string",
      "enum": [
        "Standard_LRS",
        "Standard_GRS",
        "Standard_RAGRS",
        "Standard_ZRS",
        "Premium_LRS"
      ],
      "title": "sku",
      "description": "The SKU for the storage account.",
      "default": "Standard_LRS"
    }
  },
  "type": "object",
  "required": [
    "kind",
    "location",
    "resourceGroupName",
    "sku",
    "storageAccountName"
  ]
}
```

</details>

JSON Schema -> Bicep:

```bash
airlock bicep output /path/to/schema.json
```

<details>
  <summary>Example</summary>

`schema.json`:

```json
{
  "properties": {
    "firstName": {
      "title": "First name",
      "type": "string"
    },
    "lastName": {
      "title": "Last name",
      "type": "string"
    },
    "phoneNumber": {
      "title": "Phone number",
      "type": "string",
      "minLength": 9,
      "maxLength": 12
    },
    "email": {
      "title": "Email",
      "type": "string",
      "minLength": 3
    },
    "age": {
      "title": "Age",
      "type": "integer",
      "minimum": 1
    },
    "ssn": {
      "title": "SSN",
      "type": "string",
      "format": "password",
      "minLength": 9,
      "maxLength": 9
    },
    "color": {
      "title": "Favorite color",
      "type": "string",
      "enum": [
        "Blue",
        "Red",
        "Yellow",
        "Other"
      ]
    },
    "active": {
      "title": "User is active",
      "description": "Is the user currently active?",
      "type": "boolean",
      "default": false
    }
  }
}
```

Bicep output:

```bicep
param firstName string
param lastName string
@minLength(9)
@maxLength(12)
param phoneNumber string
@minLength(3)
param email string
@minValue(1)
param age int
@minLength(9)
@maxLength(9)
@secure()
param ssn string
@allowed([
  'Blue'
  'Red'
  'Yellow'
  'Other'
])
param color string
@sys.description('Is the user currently active?')
param active bool = false
```

</details>
