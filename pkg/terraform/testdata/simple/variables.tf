variable "teststring" {
    type = string
    description = "An example string variable"
    default = "string value"
}

variable "testnumber" {
    type = number
    description = "An example number variable"
    default = 20
}

variable "testbool" {
    type = bool
    description = "An example bool variable"
    default = false
}

variable "testobject" {
  type = object({
    name    = string
    address = string
    age     = optional(number)
  })
  description = "An example object variable"
  default = {
    name = "Bob"
    address = "123 Bob St."
  }
  sensitive = true
}

variable "testlist" {
  type = list(string)
  description = "An example list variable"
}

variable "testset" {
  type = set(string)
  description = "An example set variable"
}

variable "testmap" {
  type = map(string)
  description = "An example map variable"
}

variable "nodescription" {
  type = string
}