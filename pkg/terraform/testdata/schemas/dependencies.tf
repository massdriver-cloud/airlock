variable "single" {
  type = object({
    foo = bool
    bar = optional(number)
    baz = optional(string)
  })
}
