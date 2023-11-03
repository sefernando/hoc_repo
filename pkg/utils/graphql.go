package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/hasura/go-graphql-client"
	"go.uber.org/zap"
)

type GraphqlConf struct {
	HasuraAddress   string
	GraphqlEndpint  string
	IamUrl          string
	IamClientId     string
	IamClientSecret string
	IamRealm        string
}

type AohGqlClient struct {
	Conf   GraphqlConf
	Client *graphql.Client
	Logger *zap.Logger
}

func (a *AohGqlClient) getToken() (*gocloak.JWT, error) {
	client := gocloak.NewClient(a.Conf.IamUrl)
	ctx := context.Background()

	grant := "client_credentials"
	token, err := client.GetToken(ctx, a.Conf.IamRealm, gocloak.TokenOptions{
		ClientID:     &a.Conf.IamClientId,
		ClientSecret: &a.Conf.IamClientSecret,
		GrantType:    &grant,
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (a *AohGqlClient) CreateGraphqlClient() *AohGqlClient {

	a.Logger.Info("--- Connecting to Hasura GraphQL ---")

	token, err := a.getToken()
	if err != nil {
		a.Logger.Fatal(err.Error())
	}

	graphqlURL := a.Conf.HasuraAddress + "/" + a.Conf.GraphqlEndpint
	a.Client = graphql.NewClient(graphqlURL, nil)
	a.Client = a.Client.WithRequestModifier(func(req *http.Request) {
		req.Header.Set("Authorization", token.TokenType+" "+token.AccessToken)
	})

	// Start a goroutine to periodically refresh the token
	go func() {
		for {

			minimumInterval := 10
			leewayInSeconds := 30
			timeBeforeRefresh := time.Duration(max(token.ExpiresIn-leewayInSeconds, minimumInterval)) * time.Second

			// Wait before refreshing the token again
			a.Logger.Info("Sleep before token refresh", zap.Float64("seconds", timeBeforeRefresh.Seconds()))
			time.Sleep(timeBeforeRefresh)

			token, err := a.getToken()
			if err != nil {
				a.Logger.Fatal(err.Error())
			}

			a.Client = a.Client.WithRequestModifier(func(req *http.Request) {
				req.Header.Set("Authorization", token.TokenType+" "+token.AccessToken)
			})

			a.Logger.Info("Token successfully refreshed.", zap.Int("next_expiry_in_seconds", token.ExpiresIn))
		}
	}()

	a.Logger.Info("Initialized Hasura GraphQL Client",
		zap.String("hasuraEndpoint", a.Conf.HasuraAddress),
		zap.String("graphqlEndpoint", a.Conf.GraphqlEndpint))
	return a
}
