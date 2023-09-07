package kafka

// Config kafka config
type Config struct {
	Brokers    []string `mapstructure:"brokers" validate:"required"`
	GroupID    string   `mapstructure:"groupID" validate:"required,gte=0"`
	InitTopics bool     `mapstructure:"initTopics"`
}
