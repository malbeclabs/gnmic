package types

type SASL struct {
	User      string `mapstructure:"user,omitempty"`
	Password  string `mapstructure:"password,omitempty"`
	Mechanism string `mapstructure:"mechanism,omitempty"`
	TokenURL  string `mapstructure:"token-url,omitempty"`
	// AWSRegion is the AWS region for MSK IAM authentication.
	// If not specified, uses AWS_REGION or AWS_DEFAULT_REGION environment variables.
	AWSRegion string `mapstructure:"aws-region,omitempty"`
}
