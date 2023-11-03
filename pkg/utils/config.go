package utils

import (
	"os"

	"github.com/joho/godotenv"
	// TODO: Rename package appropriate (e.g. mssfoobar/aoh-solveallyourproblems/pkg/constants)
	"mssfoobar/aoh-service-template/pkg/constants"
)

type Config struct {
	Port     string
	LogLevel string
	Graphql  GraphqlConf
}

// TODO: Review what additional configuration parameters you need here - add more environment variables etc.
func (c *Config) Load(confFile string) {
	godotenv.Load(confFile)
	c.Port = os.Getenv(constants.ENV_APP_PORT)

	c.LogLevel = os.Getenv(constants.ENV_LOG_LEVEL)

	c.Graphql = GraphqlConf{
		HasuraAddress:   os.Getenv(constants.ENV_HASURA_HOST) + ":" + os.Getenv(constants.ENV_HASURA_PORT),
		GraphqlEndpint:  os.Getenv(constants.ENV_GQL_ENDPOINT),
		IamUrl:          os.Getenv(constants.ENV_IAM_URL),
		IamClientId:     os.Getenv(constants.ENV_IAM_CLIENT_ID),
		IamClientSecret: os.Getenv(constants.ENV_IAM_CLIENT_SECRET),
		IamRealm:        os.Getenv(constants.ENV_IAM_REALM),
	}
}
