@description('an example string parameter')
param testString string = "foo"
param testInt int = 1
param testBool bool = false
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
    friends = ['steve', 'bob']
}

resource whatever 'foobar' = {
}