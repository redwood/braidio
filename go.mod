module github.com/redwood/bradio

replace github.com/redwood/liquid-dsp => ../liquid-dsp

replace github.com/redwood/hackrf => ../hackrf

go 1.17

require (
	github.com/pkg/errors v0.9.1
	github.com/redwood/hackrf v0.0.0-00010101000000-000000000000
	github.com/redwood/liquid-dsp v0.0.0-20211119033507-63d812b9add0
	go.uber.org/multierr v1.7.0
)

require go.uber.org/atomic v1.7.0 // indirect
