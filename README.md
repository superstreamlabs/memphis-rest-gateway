<div align="center">
Please pay attention that Memphis.dev is no longer supported officially by the Superstream team (formerly Memphis.dev ) and was released to the public.
  <div align="center">
 
<a href="![Github (2)](https://github.com/memphisdev/memphis.js/assets/107035359/731a59be-0f46-4a94-84c3-c0b2a07fe01c)">![Github (2)](https://github.com/memphisdev/memphis.js/assets/107035359/281222f9-8f93-4a20-9de8-7c26541bded7)</a>
<p align="center">
<a href="https://memphis.dev/discord"><img src="https://img.shields.io/discord/963333392844328961?color=6557ff&label=discord" alt="Discord"></a>
<a href="https://github.com/memphisdev/memphis/issues?q=is%3Aissue+is%3Aclosed"><img src="https://img.shields.io/github/issues-closed/memphisdev/memphis?color=6557ff"></a> 
  <img src="https://img.shields.io/npm/dw/memphis-dev?color=ffc633&label=installations">
<a href="https://github.com/memphisdev/memphis/blob/master/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Code%20of%20Conduct-v1.0-ff69b4.svg?color=ffc633" alt="Code Of Conduct"></a> 
<img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/memphisdev/memphis?color=61dfc6">
<img src="https://img.shields.io/github/last-commit/memphisdev/memphis?color=61dfc6&label=last%20commit">
</p>

  <a href="https://memphis.dev/docs/">Docs</a> - <a href="https://twitter.com/Memphis_Dev">X</a> - <a href="https://www.youtube.com/channel/UCVdMDLCSxXOqtgrBaRUHKKg">YouTube</a>
</p></b>

<div align="center">

  <h4>

**[Memphis.dev](https://memphis.dev)** is a highly scalable, painless, and effortless data streaming platform.<br>
Made to enable developers and data teams to collaborate and build<br>
real-time and streaming apps fast.

  </h4>
  
</div>..

# REST Gateway (HTTP Proxy)

## Introduction

To enable message production via HTTP calls for various use cases and ease of use, Memphis added an HTTP gateway to receive REST-based requests (=messages) and produce those messages to the required station.

Common use cases for the REST Gateway are&#x20;

* Produce events directly from a frontend
* Produce CDC events using the Debezium HTTP server
* ArgoCD webhooks
* Receive data from Fivetran/Rivery/Any ETL platform using HTTP calls

## Architecture

1. An endpoint creates an HTTP request toward the REST Gateway using **port 4444**
2. The REST gateway receives the incoming request and produces it as a message to the station

![REST gateway](https://user-images.githubusercontent.com/70286779/212469259-9f092921-63fa-4121-83cf-90f745d4b952.jpeg)


For scale requirements, the "REST gateway" component is separate from the brokers' pod and can scale out individually.

## Security Mechanisms

### JWT

Memphis REST (HTTP) gateway makes use of JWT-type identification.\
[JSON Web Tokens](https://jwt.io/) are an open, industry-standard RFC 7519 method for representing claims securely between two parties.

### API Token

Soon.

## Sequence diagram

![Sequence diagram](https://user-images.githubusercontent.com/70286779/212469294-ebf2da3f-af30-46bc-bb42-ef860159356e.jpeg)


## Getting started

If you are using Memphis **Open-Source** version, please make sure your 'REST gateway' component is exposed either through localhost or public IP.<br><br>
If you are using Memphis **Cloud**, it is already in.

### 1. Create a JWT token

Please create a JWT token, which will be part of each produce/consume request. For authentication purposes.

* The generated JWT will encapsulate all the needed information for the broker to ensure the requester is authenticated to communicate with Memphis.
* JWT token (by design) has an expiration time. Token refreshment can take place progrematically, but as it is often used to integrate memphis with other systems which are not supporting JWT refreshment, a workaround to overcome it would be to set a very high value in the `token_expiry_in_minutes`.
* The default expiry time is 15 minutes.

**Cloud (Using body params)**<br>
* Please replace the [Cloud], [Region], Username, Password, and Account ID with your parameters.
```bash
curl --location --request POST 'https://[Cloud]-[Region].restgw.cloud.memphis.dev/auth/authenticate' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "CLIENT_TYPE_USERNAME",
    "password": "CLIENT_TYPE_PASSWORD",
    "account_id": 123456789,
    "token_expiry_in_minutes": 6000000,
    "refresh_token_expiry_in_minutes": 100000
}'
```

**Cloud (Using query params)**<br>
* Please replace the [Cloud], [Region], Username, Password, and Account ID with your parameters.
```bash
curl --location --request POST 'https://[Cloud]-[Region].restgw.cloud.memphis.dev/auth/authenticate?accountId=123456789' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "CLIENT_TYPE_USERNAME",
    "password": "CLIENT_TYPE_PASSWORD",
    "token_expiry_in_minutes": 6000000,
    "refresh_token_expiry_in_minutes": 100000
}'
```

**Open-source**
```bash
curl --location --request POST 'https://REST_GW_URL:4444/auth/authenticate' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "CLIENT_TYPE_USERNAME",
    "password": "CLIENT_TYPE_PASSWORD,
    "token_expiry_in_minutes": 6000000,
    "refresh_token_expiry_in_minutes": 100000
}'
```

Expected output:&#x20;

```JSON
{"expires_in":3600000,"jwt":"eyJhbGciO***************nR5cCI6IkpXVCJ9.eyJleHAiOjE2NzQ3MTg0MjV9._A************UFoWZjp21UYVcjXwGWiYtacYPZR8","jwt_refresh_token":"eyJhbGciOiJIUzI1N***************kpXVCJ9.eyJleHAiOjIy*********************7csm-jmJv0J45YrD_slvlEOKu2rs7Q","refresh_token_expires_in":600005520000}
```
<hr>

**Refresh a token**

Before the JWT token expires or after an authentication failure, you must call the refresh procedure and get a new token. The refresh JWT token is valid by default for 5 hours.

**Cloud**<br>
* Please replace the [Cloud], [Region], Username, and Password with your parameters.
```bash
curl --location --request POST 'https://[Cloud]-[Region].restgw.cloud.memphis.dev/auth/refreshToken' \
--header 'Content-Type: application/json' \
--data-raw '{
    "jwt_refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjIyNz*******************VPSjtU4x02z_jbWhu5pIByhCRh6FU8",
    "token_expiry_in_minutes": 60000000,
    "refresh_token_expiry_in_minutes": 10000000
}'
```

**Open-source**
```bash
curl --location --request POST 'https://REST_GW_URL:4444/auth/refreshToken' \
--header 'Content-Type: application/json' \
--data-raw '{
    "jwt_refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjIyNz*******************VPSjtU4x02z_jbWhu5pIByhCRh6FU8",
    "token_expiry_in_minutes": 60000000,
    "refresh_token_expiry_in_minutes": 10000000
}'
```

Expected output:

```json
{"expires_in":3600000,"jwt":"eyJhb**************5cCI6IkpXVCJ9.eyJleHAiOjE2NzQ3MTg3N*******************F1-MmFGXRKn2sM8Yw","jwt_refresh_token":"eyJhbGciOiJIUzI*****************IkpXVCJ9.eyJleHAiOjIyNz***********************grsqYMPApAPS8YDgkT8R-69-Y5E","refresh_token_expires_in":600005520000}
```
<hr>

### 2. Produce a single message

**Supported content types:**

* text
* application/json
* application/x-protobuf


**Cloud (Using body params)**
* Please replace the [Cloud], [Region], JWT token (right after `Bearer`) with your parameters.


```bash
curl --location --request POST 'https://[Cloud]-[Region].restgw.cloud.memphis.dev/stations/STATION_NAME/produce/single' \
--header 'Authorization: Bearer eyJhbGciOiJIU**********.e30.4KOGR**************VrLKsvHo33u3UdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '{"message": "New Message"}'
```

**Cloud (Using query params)**
* Please replace the [Cloud], [Region], JWT token with your parameters.

```bash
curl --location --request POST 'https://[Cloud]-[Region].restgw.cloud.memphis.dev/stations/STATION_NAME/produce/single?authorization=eyJhbGciOiJIU**********.e30.4KOLKsvHo33u3UdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '{"message": "New Message"}'
```

**Open-source**
* Please replace the JWT token (right after `Bearer`) with your parameter.

```bash
curl --location --request POST 'rest_gateway:4444/stations/STATION_NAME/produce/single' \
--header 'Authorization: Bearer eyJhbGciOiJIU**********.e30.4KOGRhUaqvmUdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '{"message": "New Message"}'
```

Expected output:

```json
{"error":null,"success":true}
```

Schema error example:

```json
{"error":"Schema validation has failed: jsonschema: '' does not validate with file:///Users/user/memphisdev/memphis-rest-gateway/123#/required: missing properties: 'field1', 'field2', 'field3'","success":false}
```

<hr>

### 3. Produce a batch of messages
**Supported content types:**

* application/json

**Cloud (Using body params)**
* Please replace the [Cloud], [Region], JWT token (right after `Bearer`) with your parameters.


```bash
curl --location --request POST 'https://[Cloud]-[Region].restgw.cloud.memphis.dev/stations/STATION_NAME/produce/batch' \
--header 'Authorization: Bearer eyJhbGciOiJIU**********.e30.4KOGR**************VrLKsvHo33u3UdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '[
    {"message": "x"},
    {"message": "y"},
    {"message": "z"}
]'
```

**Cloud (Using query params)**
* Please replace the [Cloud], [Region], JWT token with your parameters.

```bash
curl --location --request POST 'https://[Cloud]-[Region].restgw.cloud.memphis.dev/stations/STATION_NAME/produce/batch?authorization=eyJhbGciOiJIU**********.e30.4KOLKsvHo33u3UdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '[
    {"message": "x"},
    {"message": "y"},
    {"message": "z"}
]'
```

**Open-source**
* Please replace the JWT token (right after `Bearer`) with your parameter.

```bash
curl --location --request POST 'rest_gateway:4444/stations/STATION_NAME/produce/batch' \
--header 'Authorization: Bearer eyJhbGciOiJIU**********.e30.4KOGRhUaqvmUdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '[
    {"message": "x"},
    {"message": "y"},
    {"message": "z"}
]'
```

Expected output:

```json
{"error":null,"success":true}
```

Schema error example:

```json
{"errors":["Schema validation has failed: jsonschema: '' does not validate with file:///Users/user/memphisdev/memphis-rest-gateway/123#/required: missing properties: 'field1'","Schema validation has failed: jsonschema: '' does not validate with file:///Users/user/memphisdev/memphis-rest-gateway/123#/required: missing properties: 'field1'"],"fail":2,"sent":1,"success":false}
```

<hr>


### 4. Consume a batch of messages&#x20;

To avoid reading the same message twice and reduce network traffic for an ack per message, which is not scalable, messages are auto-acknowledged by the rest gateway.

**Supported content types:**

* application/json

**Cloud (Using body params)**
* Please replace the [Cloud], [Region], JWT token (right after `Bearer`) with your parameters.


```bash
curl --location --request POST 'https://[Cloud]-[Region].restgw.cloud.memphis.dev/stations/STATION_NAME/consume/batch' \
--header 'Authorization: Bearer eyJ***************XVCJ9.e30.4KOGRhUaqvm-qSHnmMw****************' \
--header 'Content-Type: application/json' \
--data-raw '{
    "consumer_name": <consumer_name>,
    "consumer_group": <consumer_group>,
    "batch_size": <batch_size>,
    "batch_max_wait_time_ms": <batch_max_wait_time>,
}'
```

**Open-source**
* Please replace the JWT token (right after `Bearer`) with your parameter.

```bash
curl --location --request POST 'rest_gateway:4444/stations/STATION_NAME/consume/batch' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.4KOGRhUaqvm-qSHnmMwX5VrLKsvHo33u3UdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '{
    "consumer_name": <consumer_name>,
    "consumer_group": <consumer_group>,
    "batch_size": <batch_size>,
    "batch_max_wait_time_ms": <batch_max_wait_time>
}'
```

Expected output:

```json
[
  {
    "message": "{\n    \"message\": \"message x\"\n}",
    "headers": {
      "Accept": "*/*",
      "Accept-Encoding": "gzip, deflate, br"
    }
  },
  {
    "message": "{\n    \"message\": \"message y\"\n}",
    "headers": {
      "Content-Type": "application/json",
      "Host": "localhost:4444"
    }
  }
]
```

#### Error Examples:

```json
{
    "error": "Consumer name is required",
    "success": false
}
```

## Support 🙋‍♂️🤝

### Ask a question ❓ about Memphis{dev} or something related to us:

We welcome you to our discord server with your questions, doubts and feedback.

<a href="https://memphis.dev/discord"><img src="https://amplication.com/images/discord_banner_purple.svg"/></a>

### Submit a feature 💡 request 

If you have an idea, or you think that we're missing a capability that would make development easier and more robust, please [Submit feature request](https://github.com/memphisdev/memphis/issues/new?assignees=&labels=type%3A%20feature%20request).

If an issue❗with similar feature request already exists, don't forget to leave a "+1".
If you add some more information such as your thoughts and vision about the feature, your comments will be embraced warmly :)

## Contributing

Memphis{dev} is an open-source project.<br>
We are committed to a fully transparent development process and appreciate highly any contributions.<br>
Whether you are helping us fix bugs, proposing new features, improving our documentation or spreading the word - we would love to have you as part of the Memphis{dev} community.

Please refer to our [Contribution Guidelines](https://github.com/memphisdev/memphis/CONTRIBUTING.md) and [Code of Conduct](https://github.com/memphisdev/memphis/code_of_conduct.md).

## Contributors ✨

Thanks goes to these wonderful people ❤:<br><br>
 <a href = "https://github.com/memphisdev/memphis-broker/graphs/contributors">
   <img src = "https://contrib.rocks/image?repo=memphisdev/memphis"/>
 </a>

## License 📃
Memphis is open-sourced and operates under the "Memphis Business Source License 1.0" license
Built out of Apache 2.0, the main difference between the licenses is:
"You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service. A “Service” is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services."
Please check out [License](https://github.com/memphisdev/memphis/LICENSE) to read the full text.
