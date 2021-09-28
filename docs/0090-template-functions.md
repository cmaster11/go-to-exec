# Template functions

`go-to-exec` lets you use all function from the [sprig](https://github.com/Masterminds/sprig) library, and the following additional functions:

| Function name | Args | Output | Example |
|---|---|---|---|
| `cleanNewLines` | `text` | Replace all sequences of more than 2 newlines, in `text`, with 2 newlines | `cleanNewLines "Hello\n\n\n\nworld"` |
| `dump` | `value` | Prints a human-readable (YAML) representation of `value` | `dump "Hello"`, `dump .` |
| `fileReadToString` | `path` | Reads the file at `path` and returns its content as string | `fileReadToString "hello.txt"` |
| `yamlDecode` | `text` | Decodes `text` into a usable map (works only with YAML maps!) | `(yamlDecode "name: Mr. Anderson").name` |
| `yamlToJson` | `text` | Decodes `text` as YAML and re-encodes it as JSON (works only with YAML maps!) | `yamlToJson "name: Mr. Anderson"` |