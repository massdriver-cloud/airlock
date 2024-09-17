variable "addPropFalse" {
  type = object({
    foo = optional(string)
  })
}
variable "addPropTrue" {
  type = any
}
variable "addPropSchema" {
  type = map(object({
    bar = optional(string)
  }))
}
variable "pattProp" {
  type = map(string)
}
