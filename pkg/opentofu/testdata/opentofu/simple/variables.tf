variable "teststring" {
  type        = string
  description = "An example string variable"
  default     = "string value"
}

variable "testnumber" {
  type        = number
  description = "An example number variable"
  default     = 20
}

variable "testbool" {
  type        = bool
  description = "An example bool variable"
  default     = false
}

variable "testemptybool" {
  type        = bool
  description = "An example empty bool variable"
}

variable "testobject" {
  type = object({
    name    = string
    address = optional(string)
    age     = optional(number)
  })
  description = "An example object variable"
  default = {
    name    = "Bob"
    address = "123 Bob St."
  }
  sensitive = true
}

variable "testnestedobject" {
  type = object({
    name    = string
    address = optional(string, "123 Bob St.")
    age     = optional(number, 30)
    dead    = optional(bool, false)
    phones = optional(object({
      home = string
      work = optional(string, "123-456-7891")
      }), {
      home = "987-654-3210"
    })
    children = optional(list(object({
      name = string
      occupation = optional(object({
        company    = string
        experience = optional(number, 0),
        manager    = optional(bool, false)
        }), {
        company    = "Massdriver"
        experience = 1
        manager    = false
      })
      })), [{
      name = "bob"
      occupation = {
        company    = "none",
        experience = 2,
        manager    = true
      }
      }]
    )
  })
  description = "An example nested object variable"
}

variable "testlist" {
  type        = list(string)
  description = "An example list variable"
}

variable "testset" {
  type        = set(string)
  description = "An example set variable"
}

variable "testmap" {
  type        = map(string)
  description = "An example map variable"
}

variable "nodescription" {
  type = string
}
