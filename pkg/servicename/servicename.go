// TODO: Rename servicename to the name of your service (e.g. package solveallyourproblems)
// See https://go.dev/blog/package-names on package naming guidelines for Golang (Go is very opinionated on this)
package servicename

import (
	"context"
	"encoding/json"
	"io"
	"mssfoobar/aoh-service-template/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ServiceManager struct {
	timeLive   string
	timeReady  string
	logger     *zap.Logger
	gql        *(utils.AohGqlClient)
	httpServer *http.Server
}

func New() *ServiceManager {
	return &ServiceManager{}
}

type errMsg struct {
	Message string `json:"message"`
}

type response struct {
	Data    []interface{} `json:"data,omitempty"`
	Message string        `json:"message,omitempty"`
	Sent_at string        `json:"sent_at,omitempty"`
	Errors  []errMsg      `json:"errors,omitempty"`
}

type User struct {
	Id            string `json:"id,omitempty"`
	Username      string `json:"username,omitempty"`
	Email         string `json:"email,omitempty"`
	Comment       string `json:"comment,omitempty"`
	Created_at    string `json:"created_at,omitempty"`
	First_name    string `json:"first_name,omitempty"`
	Last_name     string `json:"last_name,omitempty"`
	Membership_id string `json:"membership_id"`
	Updated_at    string `json:"updated_at,omitempty"`
}

/**
* Probes:
* Any code greater than or equal to 200 and less than 400 indicates success. Any other code indicates failure.
* https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-http-request
 */

// Liveliness Endpoint of K8s
func (svcmgr *ServiceManager) getLiveness(c *gin.Context) {
	svcmgr.logger.Sugar().Infof("GET %s", c.Request.URL.Path)
	if svcmgr.timeLive == "" {
		c.String(http.StatusServiceUnavailable, "service is down")
		return
	}
	c.String(http.StatusOK, "Live since %s", svcmgr.timeLive)
}

// Readiness Endpoint of K8s
func (svcmgr *ServiceManager) getReadiness(c *gin.Context) {
	svcmgr.logger.Sugar().Infof("GET %s", c.Request.URL.Path)
	if svcmgr.timeReady == "" {
		c.String(http.StatusServiceUnavailable, "service is not yet ready")
		return
	}
	c.String(http.StatusOK, "Ready since %s", svcmgr.timeReady)
}

func (svcmgr *ServiceManager) postUsers(c *gin.Context) {
	var resp response
	svcmgr.logger.Sugar().Infof("POST %s", c.Request.URL.Path)
	if svcmgr.timeReady == "" {
		resp.Errors = append(resp.Errors, errMsg{Message: "service is not yet ready"})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		svcmgr.logger.Error(err.Error())
		resp.Errors = append(resp.Errors, errMsg{Message: err.Error()})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var data User
	err = json.Unmarshal(raw, &data)
	if err != nil {
		svcmgr.logger.Error(err.Error())
		resp.Errors = append(resp.Errors, errMsg{Message: err.Error()})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// TODO: Match the GraphQL mutation you are using
	type aoh_ums_user_insert_input User

	// TODO: Create the GraphQL mutation string
	var m struct {
		User `graphql:"insert_aoh_ums_user_one(object: $Obj)"`
	}

	v := map[string]interface{}{
		"Obj": aoh_ums_user_insert_input(data),
	}

	err = svcmgr.gql.Client.Mutate(context.Background(), &m, v)

	if err != nil {
		svcmgr.logger.Error(err.Error())
		resp.Errors = append(resp.Errors, errMsg{Message: err.Error()})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	svcmgr.logger.Info("Created new user", zap.String("id", m.User.Id))
	resp.Data = append(resp.Data, m.User)
	resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
	c.JSON(http.StatusOK, resp)
}

func (svcmgr *ServiceManager) getUsers(c *gin.Context) {
	var resp response

	svcmgr.logger.Sugar().Infof("GET %s", c.Request.URL.Path)
	if svcmgr.timeReady == "" {
		resp.Errors = append(resp.Errors, errMsg{Message: "service is not yet ready"})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	type Users []User
	var q struct {
		Users `graphql:"aoh_ums_user"`
	}

	err := svcmgr.gql.Client.Query(context.Background(), &q, nil)
	if err != nil {
		svcmgr.logger.Error(err.Error())
		resp.Errors = append(resp.Errors, errMsg{Message: err.Error()})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	for _, v := range q.Users {
		resp.Data = append(resp.Data, v)
	}

	resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
	c.JSON(http.StatusOK, resp)
}

func (svcmgr *ServiceManager) getUserById(c *gin.Context) {
	var resp response
	userId := c.Param("user_id")

	svcmgr.logger.Sugar().Infof("GET %s", c.Request.URL.Path)
	if svcmgr.timeReady == "" {
		resp.Errors = append(resp.Errors, errMsg{Message: "service is not yet ready"})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	var q struct {
		User `graphql:"aoh_ums_user_by_pk(id: $ID)"`
	}

	v := map[string]interface{}{
		"ID": userId,
	}

	err := svcmgr.gql.Client.Query(context.Background(), &q, v)
	if err != nil {
		svcmgr.logger.Error(err.Error())
		resp.Errors = append(resp.Errors, errMsg{Message: err.Error()})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Data = append(resp.Data, q.User)
	resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
	c.JSON(http.StatusOK, resp)
}

func (svcmgr *ServiceManager) patchUserById(c *gin.Context) {
	var resp response
	incidentId := c.Param("user_id")

	svcmgr.logger.Sugar().Infof("PATCH %s", c.Request.URL.Path)
	if svcmgr.timeReady == "" {
		resp.Errors = append(resp.Errors, errMsg{Message: "service is not yet ready"})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		svcmgr.logger.Error(err.Error())
		resp.Errors = append(resp.Errors, errMsg{Message: err.Error()})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var data User
	err = json.Unmarshal(raw, &data)
	if err != nil {
		svcmgr.logger.Error(err.Error())
		resp.Errors = append(resp.Errors, errMsg{Message: err.Error()})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	type aoh_ums_user_set_input User

	var m struct {
		User `graphql:"update_aoh_ums_user_by_pk(pk_columns: {id: $ID}, _set: $Obj)"`
	}

	v := map[string]interface{}{
		"ID":  incidentId,
		"Obj": aoh_ums_user_set_input(data),
	}

	err = svcmgr.gql.Client.Mutate(context.Background(), &m, v)
	if err != nil {
		svcmgr.logger.Error(err.Error())
		resp.Errors = append(resp.Errors, errMsg{Message: err.Error()})
		resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Data = append(resp.Data, m.User)
	resp.Sent_at = time.Now().UTC().Format(time.RFC3339)
	c.JSON(http.StatusOK, resp)
}

func (svcmgr *ServiceManager) Start(conf utils.Config) {
	logger := utils.InitZapLogger(conf.LogLevel)
	agql := (utils.AohGqlClient{
		Conf:   conf.Graphql,
		Logger: logger,
	})

	svcmgr.timeLive = time.Now().Format(time.RFC3339)
	svcmgr.timeReady = time.Now().Format(time.RFC3339)
	svcmgr.logger = logger
	svcmgr.gql = agql.CreateGraphqlClient()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Note: In AGIL Ops Hub, we handle CORS via Traefik. However, we recommend you use
	// https://github.com/gin-contrib/cors as a middleware for CORS if your architecture requires you to do this within
	// the service itself
	//
	// For local development, run Chrome with web-security disabled for testing
	router.Use(gin.Recovery())

	// Use Router Groups to segregate endpoints by what middlewares should be run on them,
	// for example, the info, liveliness, and rediness endpoints do not require a connection to the database, and thus
	// do not require ensuring the JWT is still valid
	v1Group := router.Group("/v1")
	usersGroup := v1Group.Group("/users")

	// Note: Liveliness and Readiness endpoints are mandatory for Kubernetes
	// Ensure your readiness endpoint only publishes success when your service
	// is actually ready to serve data (e.g. after it has fetched all required tokens).
	v1Group.GET("/info/liveness", svcmgr.getLiveness)
	v1Group.GET("/info/readiness", svcmgr.getReadiness)

	// TODO: Create whatever endpoints you need
	usersGroup.POST("", svcmgr.postUsers)
	usersGroup.GET("", svcmgr.getUsers)
	usersGroup.GET("/:user_id", svcmgr.getUserById)
	usersGroup.PATCH("/:user_id", svcmgr.patchUserById)

	port := ":" + conf.Port
	svcmgr.httpServer = &http.Server{
		Addr:    port,
		Handler: router,
	}

	logger.Info("HTTP Server Starting", zap.String("port", conf.Port))
	go func() {
		if err := svcmgr.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			svcmgr.logger.Error(err.Error())
		}
		svcmgr.logger.Info("HTTP Server Shutdown")
	}()
}

func (svcmgr *ServiceManager) Stop() {
	svcmgr.logger.Info("Service exiting...")
	err := svcmgr.httpServer.Shutdown(context.Background())
	if err != nil {
		svcmgr.logger.Error(err.Error())
	}
	svcmgr.logger.Info("Service exited successfully.")
}
