package config

type Config struct {
	Key   string `json:"key" default:"" validate:"required,max=100" label:"配置标识"`
	Value string `json:"value" default:"" validate:"omitempty" label:"配置内容"`
}
