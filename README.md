# Drugstore

Store JSON stuffs in a REST path. Find stuff with [jmespath](http://jmespath.org/).

## Paths

First level is the class. A class has path, ordered collection of attributes.
Values of th path must be unique.

Example class is _product_ with path : _product_, _namespace_, _name_.

An object can be :

```json
{
  "product①": "drugstore",
  "namespace①": "user",
  "name①": "zoe",
  "likes②": ["banana", "apple"],
  "age②": 42
}
```

① This keys are mandatory, it's define the path. The value must be string.

② You can add more keys, with any value type, even object.

Its path is `/product/drugstore/user/zoe`.

## Licence

3 terms BSD Licence, © 2019 Mathieu Lecarme
