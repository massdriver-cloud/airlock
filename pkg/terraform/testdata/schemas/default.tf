variable "emptytest" {
  type    = string
  default = null
}
variable "stringtest" {
  type    = string
  default = "foo"
}
variable "integertest" {
  type    = number
  default = 3
}
variable "numbertest" {
  type    = number
  default = 3.4
}
variable "booltest" {
  type    = bool
  default = true
}
variable "arraytest" {
  type    = list(string)
  default = ["foo"]
}
variable "objecttest" {
  type = object({
    foo = string
  })
  default = {"foo":"bar"}
}
variable "requiredtest" {
  type = string
}
