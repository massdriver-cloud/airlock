@description('an example string parameter')
@minLength(2)
@maxLength(20)
@allowed(['foo','bar'])
param testString string = "foo"

@minValue(0)
@maxValue(10)
@allowed([1,5,7])
param testInt int = 1

param testBool bool = false

@minLength(1)
@maxLength(8)
param testArray array = [1, 2, 3]

param testObject object = {
    name: 'hugh'
    age: 20
    member: true
    nested: {
        foo: 'bar'
        nested2: {
            hello: 'world'
        }
    }
    friends: ['steve', 'bob']
    empty: []
}

param testArrayObject array = [
    {
        foo: 'bar',
        num: 10
    },
    {
        foo: 'baz',
        num: 2
    }
]

param testEmptyObject object = {}
param testEmptyArray array = []

@secure()
param testSecureString string

@secure()
param testSecureObject object

resource whatever 'foobar' = {
}