#!/bin/bash
if [ -z "$CLIENT_ID" ]; then
    echo "Enter Service Name (e.g. 'aoh_solve_all_your_problems'):"
    read
    CLIENT_ID=$REPLY
    echo -e
fi

if [ -z "$CLIENT_NAME" ]; then
    echo "Enter Service Label (e.g. 'AGIL Ops Hub - Solve All Your Problems Service'):"
    read
    CLIENT_NAME=$REPLY
    echo -e
fi

if [ -z "$CLIENT_DESC" ]; then
    echo "Enter Service Description (e.g. 'This service will solve all your problems.'):"
    read
    CLIENT_DESC=$REPLY
    echo -e
fi

if [ -z "$KEYCLOAK_USERNAME" ]; then
    echo "Enter Keycloak Username (e.g. user):"
    read
    KEYCLOAK_USERNAME=$REPLY
    echo -e
fi

if [ -z "$KEYCLOAK_PASSWORD" ]; then
    echo "Enter Keycloak Password (e.g. password123):"
    read
    KEYCLOAK_PASSWORD=$REPLY
    echo -e
fi

if [ -z "$KEYCLOAK_URL" ]; then
    echo "Enter Keycloak URL (e.g. https://iam.dev.aoh):"
    read
    KEYCLOAK_URL=$REPLY
    echo -e
fi

if [ -z "$KEYCLOAK_REALM" ]; then
    echo "Enter Keycloak Realm (e.g. aoh):"
    read
    KEYCLOAK_REALM=$REPLY
    echo -e
fi

export CONTINUE=n

while [ "$CONTINUE" = "n" ]; do

    echo "1. Service Name: $CLIENT_ID"
    echo "2. Service Label: $CLIENT_NAME"
    echo "3. Service Description: $CLIENT_DESC"
    echo "4. Keycloak Username: $KEYCLOAK_USERNAME"
    echo "5. Keycloak Password: $KEYCLOAK_PASSWORD"
    echo "6. Keycloak URL: $KEYCLOAK_URL"
    echo "7. Keycloak Realm: $KEYCLOAK_REALM"

    echo "Pick item to re-enter values (1-7 or 'Y/n' to proceed or terminate):"
    read

    if [ "$REPLY" = "Y" ]; then
        CONTINUE=Y
    elif [ "$REPLY" = "y" ]; then
        CONTINUE=Y
    elif [ "$REPLY" = "n" ]; then
        echo "Program Terminated."
        exit
    elif [ "$REPLY" = "1" ]; then
        echo "Enter Service Name (e.g. 'aoh_solve_all_your_problems'):"
        read
        CLIENT_ID=$REPLY
        echo -e
    elif [ "$REPLY" = "2" ]; then
        echo "Enter Service Label (e.g. 'AGIL Ops Hub - Solve All Your Problems Service'):"
        read
        CLIENT_NAME=$REPLY
        echo -e
    elif [ "$REPLY" = "3" ]; then
        echo "Enter Service Description (e.g. 'This service will solve all your problems.'):"
        read
        CLIENT_DESC=$REPLY
        echo -e
    elif [ "$REPLY" = "4" ]; then
        echo "Enter Keycloak Username (e.g. user):"
        read
        KEYCLOAK_USERNAME=$REPLY
        echo -e
    elif [ "$REPLY" = "5" ]; then
        echo "Enter Keycloak Password (e.g. password123):"
        read
        KEYCLOAK_PASSWORD=$REPLY
        echo -e
    elif [ "$REPLY" = "6" ]; then
        echo "Enter Keycloak URL (e.g. https://iam.dev.aoh):"
        read
        KEYCLOAK_URL=$REPLY
        echo -e
    elif [ "$REPLY" = "7" ]; then
        echo "Enter Keycloak Realm (e.g. aoh):"
        read
        KEYCLOAK_REALM=$REPLY
        echo -e
    else
        echo "Unknown command, terminating."
        exit
    fi
done

# Prepare data for new client creation
jq ".clientId = \"$CLIENT_ID\" |
.name = \"$CLIENT_NAME\" |
.description = \"$CLIENT_DESC\"" \
    client.json >new_client.json && mv new_client.json client.json

# Create Keycloak Access Token for the rest of the commands
TOKEN="$(curl -s -X POST -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=$KEYCLOAK_USERNAME&password=$KEYCLOAK_PASSWORD&grant_type=password&client_id=admin-cli" \
    $KEYCLOAK_URL/realms/master/protocol/openid-connect/token | jq -r ".access_token")"

# Create the new Keycloak client with the prepared data
curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d @./client.json \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/clients

# Prepare data for the new role
jq ".name = \"$CLIENT_ID\" |
 .description = \"This is the service account role for '$CLIENT_ID.'\"" \
    role.json >new_role.json && mv new_role.json role.json

# Create the new role
curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d @./role.json \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/roles

# Prepare the data for the new group
jq ".name = \"$CLIENT_ID\" |
 .path = \"/$CLIENT_ID\" |
 .attributes.\"default-role\"[0] = \"$CLIENT_ID\" |
 .realmRoles[0] = \"$CLIENT_ID\"" \
    group.json >new_group.json && mv new_group.json group.json

# Create the new group
curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d @./group.json \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/groups

# Get Client UUID
CLIENT_UUID="$(curl -s -X GET -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/clients/?clientId=$CLIENT_ID | jq -r ".[0].id")"

# Get Role UUID
ROLE_ID=$(curl -s -X GET -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Authorization: Bearer $TOKEN" \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/roles | jq -r ".[] | select(.name == \"$CLIENT_ID\").id")

# Get Group UUID
GROUP_ID=$(curl -s -X GET -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Authorization: Bearer $TOKEN" \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/groups | jq -r ".[] | select(.name == \"$CLIENT_ID\").id")

# Get Service Account UUID
SERVICE_ACC_ID=$(curl -s -X GET -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Authorization: Bearer $TOKEN" \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/clients/$CLIENT_UUID/service-account-user | jq -r ".id")

# Create group-role mapping
curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "[{ \"id\": \"$ROLE_ID\", \"name\": \"$CLIENT_ID\" }]" \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/groups/$GROUP_ID/role-mappings/realm

# Put new service client into the group
curl -s -X PUT -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Authorization: Bearer $TOKEN" \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/users/$SERVICE_ACC_ID/groups/$GROUP_ID

# Get the client secret
CLIENT_SECRET=$(curl -s -X GET -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN " \
    $KEYCLOAK_URL/admin/realms/$KEYCLOAK_REALM/clients/$CLIENT_UUID/client-secret | jq -r ".value")

echo "Client UUID: $CLIENT_UUID"
echo "Client Secret: $CLIENT_SECRET"

# Verify
echo "Created Service Account: "$(curl -s -X POST -H "Content-Type: application/x-www-form-urlencoded" \
    -d "client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET&grant_type=client_credentials&scope=openid" \
    $KEYCLOAK_URL/realms/$KEYCLOAK_REALM/protocol/openid-connect/token |
    jq -r '.access_token | split(".") | .[0],.[1] | @base64d | fromjson | select( .preferred_username != null ).preferred_username')
