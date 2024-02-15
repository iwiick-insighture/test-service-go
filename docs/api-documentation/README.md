# Quickstart </>

This Microservices mainly provides two sets of API endpoints.

- System Secret Manager Routes - Used for registering to use the Secret Service
- Secret Routes - Used for managing secrets once registered to use the Secret Service via the System secret manager routes.

Every route deals with the following Headers. Organization Id is required while Project Id and Scope are optional.

| Header              | Type     | Description                   |
| :------------------ | :------- | :---------------------------- |
| `x-organization-id` | `string` | **Required**. Organization Id |
| `x-project-id`      | `string` | Project Id                    |
| `x-scope`           | `string` | Scope                         |

Supported project level scopes are `OTHERS`, `CONFIGS` and `CREDENTIALS`

Failing to provide the Organization Id in the `x-organization-id` header will result in the following response with `401` status code. Every error responses provided via the APIs are returned in this format.

```json
{
  "success": false,
  "message": "ERROR",
  "error": "Organization Id cannot be Empty. Check Headers !"
}
```

Success responses are provided as follows. The `data` attribute is only returned for success responces,

```json
{
  "success": true,
  "message": "success message",
  "data": "data returned from the API"
}
```

# System Secret Endpoints </>

## `POST` Add System Secret & `PUT` Update/Migrate System Secret

Adding a System Secret is the starting point for using the Secret Service. This endpoint creates a secret containing metadata needed for executing the secret routes in the System secret manager. The Update System secret endpoint will update a secret in the System secret Manager

```http
POST /system
```

```http
PUT /system
```

Both endpoints require a JSON body with a `flow` attribute that can be a string either `SHARED` or `PRIVATE`.

<br>

> ⚠️ **Note**  
> Organization Level Secrets can be stored by registering with only the `OrganizationId` header. Registering project level secrets will create 3 scopes for `OTHERS`, `CREDENTIALS` and `CONFIGS` > <br/>

### SHARED Flow

```json
{
  "flow": "SHARED"
}
```

### PRIVATE Flow

`PRIVATE` flow require a JSON body with additional `arn, region` and `provider` attributes that can be a strings.

```json
{
  "flow": "PRIVATE",
  "arn": "arn:aws:iam::438463683713:role/SMTestRoleChama",
  "region": "ap-southeast-2",
  "provider": "AWS"
}
```

Failing to provide the attributes `arn` , `region`, `provider` for the `PRIVATE` flow in request body will result in the following response with `401` status code for each attribute.

```json
{
  "success": false,
  "message": "ERROR",
  "error": "Missing arn for 'Private' flow"
}
```

Failing to provide values for the `arn` , `region` , `provider`
in request body will result in the following response with `401` status code for each attribute.

```json
{
  "success": false,
  "message": "ERROR",
  "error": "Missing values for attributes in request Body. !"
}
```

<br>

> ⚠️ **Note**  
> The Update System Secret Endpoint is used for one-way migration of secrets from the `SHARED` secret manager to the `PRIVATE` secret Manager
> <br/>

`PUT` endpoints require a JSON body with a flow and required attributes

```json
{
  "success": true,
  "message": "System Secret Updated",
  "data": "For Org1, migrated secret's UUIDS are ca1749f0-a1b6-498b-b245-01b378ed2dee"
}
```

Attempting to do update flow from `SHARED` to `SHARED`
in request body will result in the following response with `401` status code.

```json
{
  "success": false,
  "message": "ERROR",
  "error": "Invalid migartion attempt. Check values for flow !"
}
```

## `DELETE` Delete System Secret

Deletes a System secret in the System secret Manager.

```http
DELETE /system
```

> ⚠️ **Note**  
> If the system secret you are deleting has the `flow` defined as `SHARED`, then all secrets registered to this system secret will also be deleted from the Shared Secret Manager
> <br/>

---

# Secret Endpoints </>

These endpoints return an unauthorized `401` response if the "Organization ID", "Project ID" or `Secret Name` provided via the headers are not registered to use the secret service first using the System secret Routes.

```json
{
  "success": false,
  "message": "UNAUTHORIZED",
  "error": "Provided secret (Organization ID), is not registered to use the Secret Service !"
}
```

These endpoints return an unauthorized `401` response if the "Organization ID", "Project ID" and "Scope" or `Key` provided via the headers are not registered to use the secret service first using the System secret Routes.

```json
{
  "success": false,
  "message": "UNAUTHORIZED",
  "error": "Provided key (Organization + Project + Scope), is not registered to use the Secret Service !"
}
```

## `POST` Add Secret & `PUT` Update Secret

Once registered to use the Secret Service, Add Secret and Update secret route can be used to create a new secret or update that secret using the unique `UUID`

```http
POST /secret
```

```http
PUT /secret/:id
```

Both endpoints require a JSON body with a `secret` attribute that must be a `base64` encoded string.

```json
{
  "secret": "ewogICAgImtleTEiOiAidmFsdWUxIiwKICAgICJrZXkyIjogInZhbHVlMiIKfQ=="
}
```

A successful request will store new secrets in `PRIVATE` or `SHARED` account according to `'flow'` type and return the `UUID` of the secret in the response.

```json
{
  "success": true,
  "message": "New Secret Added",
  "data": "74361e40-b0f9-4d28-97f4-9a0c972e6d64"
}
```

Failing to provide a `base64` encoded string as the value for the `secret` attribute in the request will result in an error response.

```json
{
  "success": false,
  "message": "ERROR",
  "error": "illegal base64 data at input byte 4 or 'secret' attribute value might not be Base64 Encoded !"
}
```

<br/>

## `GET` Get Secret

Retrieves a secret using the unique `UUID`

```http
GET /secret/:id
```

| Params    | Type     | Description                  |
| :-------- | :------- | :--------------------------- |
| `version` | `string` | `uuid` string of the version |

If a secret exists for the provided `UUID`, the `base64` encoded secret will be returned from `PRIVATE` or `SHARED` account according to `'flow'` type .

```json
{
  "success": true,
  "message": "Secret Returned",
  "data": "ewogICAgImtleTEiOiAidmFsdWUxIiwKICAgICJrZXkyIjogInZhbHVlMiIKfQ=="
}
```

<br/>

## `GET` Get Secret Versions

Retrieves the secret versions given the unique `UUID`

```http
GET /secret/versions/:id
```

```json
{
  "success": true,
  "message": "Secret Versions Returned",
  "data": [
    {
      "CreatedDate": "2023-08-21T06:01:34.008Z",
      "KmsKeyIds": ["DefaultEncryptionKey"],
      "LastAccessedDate": "2023-08-21T00:00:00Z",
      "VersionId": "956df742-26b6-4bfb-99d5-3ae3451996cf",
      "VersionStages": ["AWSCURRENT"]
    }
  ]
}
```

<br/>

## `DELETE` Delete Secret

Deletes a secret using the unique `UUID` from `PRIVATE` or `SHARED` account according to `'flow'` type

```http
DELETE /secret/:id
```

```json
{
  "success": true,
  "message": "Secret Deleted",
  "data": "1d913e40-eb3f-4336-aa01-36e508e6d2e0"
}
```

<br/>

## `DELETE` Delete Secret Group

Deletes the entire group of secrets created using the "Organization ID", "Project ID" and "Scope" from `PRIVATE` or `SHARED` account according to `'flow'` type

```http
DELETE /secret/group
```

```json
{
  "success": true,
  "message": "Secret Group deleted with all secrets"
}
```
