# AGIL Ops Hub Service Template

This template can be used to quickly create a service that is ready to serve endpoints for other services to utilize,
and can establish a connection to the database.

Refer to the docs for a step-by-step guide on setting up a new service with this template:
https://mssfoobar.github.io/aoh-docs/docs/development/app/guides/services

## Project Structure

Below is a summary of our project structure, which follows Golang's standards. For more information, see: https://github.com/golang-standards/project-layout

### cmd

This is where the application code sits. This typicall calls reusable packages from the `pkg` folder. If you have types, classes, etc. that can be reused in other applications, it should
reside in `pkg` instead.

It's common for `cmd` to have a small `main` function that simply imports code from `pkg`.

### pkg

This should contain library code that can be used by other applications. Ideally, your package code could function as an SDK which can be imported and used in other Golang projects to access and process data in `AOH`.

### docker

Your service may contain multiple packages in your `pkg` folder, each of these packages should have its own Dockerfile.
In this sample, we'll just have one called `service.Dockerfile`, you should rename this accordingly.
(e.g. `solve-all-your-problems.Dockerfile`)

### docker-compose

The `docker-compose` file contains all the variables

## Create a Keycloak Client for your Service

This is required for your service to access the system with a 'service account'. A sample
`client.json` is included in this repo under the config folder - these settings can be imported into Keycloak to
quickly create your client.

There are 3 fields you need to change:

1. clientId
2. name
3. description

Then, you can import the `.json` file as-is into Keycloak, either by calling the
[Admin API's create client endpoint](https://www.keycloak.org/docs-api/22.0.1/rest-api/#_clients):

```
POST /admin/realms/{realm}/clients

```

Or by using Keycloak's console UI:

1. Log in to the console
2. Choose the appropriate realm from the drop-down list (e.g. ar2 or aoh or aocs etc.)
3. In the "Clients list" tab, click "Import client" and upload the `client.json` file

After creating your new service client, you will be able to view the client secret from
Clients > Select your client > Credentials > Client Secret
