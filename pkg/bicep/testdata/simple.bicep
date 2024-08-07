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
@description('This is a description')
param descriptiontest string
@allowed([
    'foo'
    'bar'
    'baz'
])
@description('This is a new description')
param descriptionenumtest string
@minValue(5)
param minvaluetest int
@maxValue(10)
param maxvaluetest int
@minValue(5)
@maxValue(10)
param minmaxvaluetest int
