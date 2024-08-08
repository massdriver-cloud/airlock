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
@minLength(5)
param minlengthtest string
@maxLength(10)
param maxlengthtest string
@minLength(5)
@maxLength(10)
param minmaxlengthtest string
param defaulttest string = 'foo'
param defaultintegertest int = 5
