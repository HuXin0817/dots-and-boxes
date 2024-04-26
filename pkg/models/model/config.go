package model

type Config bool

const (
	On  Config = true
	Off Config = false
)

var configName = map[string]Config{
	"ON": On,
	"On": On,
	"on": On,
	"1":  On,

	"OFF": Off,
	"Off": Off,
	"off": Off,
	"0":   Off,
}

func NewConfig(s string) Config {
	return configName[s]
}
