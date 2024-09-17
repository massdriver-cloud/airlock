variable "stringtest" {
  type = string
}
variable "integertest" {
  type = number
}
variable "numbertest" {
  type = number
}
variable "booltest" {
  type = bool
}
variable "arraytest" {
  type = list(string)
}
variable "objecttest" {
  type = object({
    foo = string
    bar = optional(number)
  })
}
variable "nestedtest" {
  type = object({
    top = object({
      nested = string
    })
  })
}
