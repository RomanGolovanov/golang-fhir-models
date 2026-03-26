# Golang FHIR Models

This repository contains FHIR® R4 models for Go. The models consist of Go structs for each resource and data type, suitable for JSON encoding/decoding.

## Features

* resources implement the [Marshaler][1] interface
* unmarshal functions are provided for every resource
* enums are provided for every ValueSet used in a [required binding][2], has a computer friendly name and refers only to one CodeSystem
* enums implement `Code()`, `Display()` and `Definition()` methods
* polymorphic (`value[x]`) fields are always generated as optional pointers — only the chosen variant is marshaled, preventing non-compliant JSON with zero-value entries for unused types
* primitive type extensions are fully supported: every primitive field is accompanied by a `{Field}Element *PrimitiveElement` (or `[]*PrimitiveElement` for arrays) that carries the FHIR `_fieldName` companion object (element id and extensions)

## Primitive Type Extensions

The [FHIR JSON spec][3] allows primitive values to carry an `id` and `extension` via an underscore-prefixed companion field. This library represents those companions as `PrimitiveElement` siblings:

```go
// scalar primitive
BirthDate        *string           `json:"birthDate,omitempty"`
BirthDateElement *PrimitiveElement `json:"_birthDate,omitempty"`

// array primitive — parallel slice, nil entries where no extension is present
Given        []string            `json:"given,omitempty"`
GivenElement []*PrimitiveElement `json:"_given,omitempty"`
```

`PrimitiveElement` is defined as:

```go
type PrimitiveElement struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

Companion fields are always `omitempty`, so existing code that does not use primitive extensions is unaffected.

## Polymorphic Fields

FHIR `value[x]` elements are expanded into one field per allowed type, all generated as optional pointers regardless of the element's `min` cardinality. This ensures that only the populated variant is written to JSON:

```go
// Observation.value[x]
ValueQuantity        *Quantity        `json:"valueQuantity,omitempty"`
ValueCodeableConcept *CodeableConcept `json:"valueCodeableConcept,omitempty"`
ValueString          *string          `json:"valueString,omitempty"`
// ...
```

## Usage

In your project, import `github.com/samply/golang-fhir-models/fhir-models/fhir` and you are done.

## TODOs

* [Support ValueSets Referring to Multiple CodeSystems](https://github.com/samply/golang-fhir-models/issues/2)

## Develop

This repository contains two Go modules, the generated models itself and the generator. Both modules use `go generate` to generate the FHIR models. For `go generate` to work, you have to install the generator first. To do that, run `go install` in the `fhir-models-gen` directory. After that, you can regenerate the FHIR Models under `fhir-models` and the subset of FHIR models under `fhir-models-gen`.

## License

Copyright 2019 - 2022 The Samply Community

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

[1]: <https://golang.org/pkg/encoding/json/#Marshaler>
[2]: <https://www.hl7.org/fhir/terminologies.html#strength>
[3]: <https://www.hl7.org/fhir/json.html#primitive>
