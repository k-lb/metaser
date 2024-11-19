# metaser

[![Go Reference](https://pkg.go.dev/badge/github.com/k-lb/metaser.svg)](https://pkg.go.dev/github.com/k-lb/metaser)

Metaser implements serialization and deserialization of structs into/from Kubernetes object's metadata. The primary use-case of this library is to allow easy storing of user-defined data structures in Kubernetes object's annotations and labels.

It's easy to use as it provides simple API and utilizes well-known struct tagging.
Public API contains of just `Encoder` type with `Encode` method and `Decoder` type with `Decode` method.
There are also `Marshal` and `Unmarshal` functions which provide same functionality without need to create
`Encoder` or `Decoder` variables.

## Example
```
type MyData struct {
    MyAnnotationVal int     `k8s:"annotation:myannotation,omitempty"`
    MyLabelVal      float32 `k8s:"label:mylabel,in"`
    MyNameVal       string  `k8s:"name,in"`
}

decoder := metaser.NewDecoder()
encoder := metaser.NewEncoder()

// meta is k8s.io/apimachinery/pkg/apis/meta/v1.Object and data is MyData.
if err := decoder.Decode(meta, &data); err != nil {
    log.Fatalln("unable to deserialize k8s object metadata to MyData struct")
}

// meta is k8s.io/apimachinery/pkg/apis/meta/v1.Object and data is MyData.
if err := encoder.Encode(&data, meta); err != nil {
    log.Fatalln("unable to serialize MyData to kubernetes object metadata")
}
```
