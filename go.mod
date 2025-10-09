module github.com/go-gandi/terraform-provider-gandi/v2

go 1.16

require (
	github.com/fatih/color v1.9.0 // indirect
	github.com/go-gandi/go-gandi v0.7.0
	github.com/google/uuid v1.1.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.16.0
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/oklog/run v1.1.0 // indirect
)

replace github.com/go-gandi/go-gandi => github.com/landure/go-gandi v0.3.0
