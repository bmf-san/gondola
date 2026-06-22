module github.com/bmf-san/gondola/_examples/standalone

go 1.24.1

require github.com/bmf-san/gondola v0.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/bmf-san/gondola => ../..
