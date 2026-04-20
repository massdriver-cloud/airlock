variable "single" {
  type = string
}
variable "then" {
  type = object({
    double  = string
    another = optional(string)
    nested  = optional(string)
  })
  default = null
}
variable "else" {
  type    = string
  default = null
}
