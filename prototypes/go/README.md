# Config prototype with Go

## Regenerate model

The model is generated from the json schema using [gojsonschema](https://github.com/xeipuuv/gojsonschema).

```bash
gojsonschema -p otel ../../json_schema/schema/schema.json > internal/otel/model.go
```

## Run

```bash
go run main.go
```

## Issues

- codegen checks for null value even if field is nullable. The following code must be edited manually:

```go
if _, ok := raw["always_off"]; !ok {
    return fmt.Errorf("field always_off: required")
}
if _, ok := raw["always_on"]; !ok {
    return fmt.Errorf("field always_on: required")
}
```

- provider shutdown causes panic
