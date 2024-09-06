param stringtest string
param integertest int
param numbertest int
param booltest bool
param arraytest array
param objecttest object
param nestedtest object
@allowed([
  'foo'
  'bar'
])
param enumtest string
@allowed([
  1
  2
])
param enumtestints int
@allowed([
  true
  false
])
param enumtestbools bool
@allowed([
  [
    'foo'
    'bar'
  ]
  [
    'baz'
    'qux'
  ]
])
param enumtestarrays array
@allowed([
  {
    foo: 'bar'
  }
  {
    baz: 'qux'
  }
])
param enumobjecttest object
@sys.description('This is a description')
param descriptiontest string
@sys.description('This is a new description')
@allowed([
  'foo'
  'bar'
  'baz'
])
param descriptionenumtest string
@minValue(5)
param minvaluetest int
@maxValue(10)
param maxvaluetest int
@minValue(5)
@maxValue(10)
param minmaxvaluetest int
@minLength(5)
param minlengthstringtest string
@maxLength(10)
param maxlengthstringtest string
@minLength(5)
@maxLength(10)
param minmaxlengthstringtest string
@minLength(2)
param minlengtharraytest array
@maxLength(5)
param maxlengtharraytest array
@minLength(2)
@maxLength(5)
param minmaxlengtharraytest array
param defaultstringtest string = 'foo'
param defaultintegertest int = 5
param defaultbooltest bool = true
param defaultarraytest array = [
  'foo'
  'bar'
]
param defaultobjecttest object = {
  bar: 'baz'
  foo: 5
}
param defaultspaceobjecttest object = {
  foo: 'bar baz'
  lorem: 'ipsum'
}
param defaultarrayobjecttest array = [
  {
    bar: 'baz'
    foo: 5
  }
  {
    bar: 'qux'
    foo: 10
  }
]
param defaultnestedarraytest array = [
  [
    [
      'foo'
    ]
    [
      'bar'
    ]
  ]
  [
    [
      'baz'
    ]
    [
      'qux'
    ]
  ]
]
param defaultnestedobjecttest object = {
  foo: {
    bar: {
      baz: 'qux'
    }
  }
  quid: {
    pro: 'quo'
  }
}
@secure()
param securestringtest string
