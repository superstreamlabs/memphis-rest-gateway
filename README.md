<div align="center">
  
  ![Banner- Memphis dev streaming  (1)](https://github.com/memphisdev/memphis-rest-gateway/assets/107035359/e28bba50-d6f5-48fc-955e-a242c9bb3028)

  
</div>

<div align="center">

  <h4>

**[Memphis](https://memphis.dev)** is an intelligent, frictionless message broker.<br>
Made to enable developers to build real-time and streaming apps fast.

  </h4>
  
  <a href="https://landscape.cncf.io/?selected=memphis"><img width="200" alt="CNCF Silver Member" src="https://github.com/cncf/artwork/raw/master/other/cncf-member/silver/white/cncf-member-silver-white.svg#gh-dark-mode-only"></a>
  
</div>

<div align="center">
  
  <img width="200" alt="CNCF Silver Member" src="https://github.com/cncf/artwork/raw/master/other/cncf-member/silver/color/cncf-member-silver-color.svg#gh-light-mode-only">
  
</div>
 
 <p align="center">
  <a href="https://memphis.dev/docs/">Docs</a> - <a href="https://twitter.com/Memphis_Dev">Twitter</a> - <a href="https://www.youtube.com/channel/UCVdMDLCSxXOqtgrBaRUHKKg">YouTube</a>
</p>

<p align="center">
<a href="https://discord.gg/WZpysvAeTf"><img src="https://img.shields.io/discord/963333392844328961?color=6557ff&label=discord" alt="Discord"></a>
<a href="https://github.com/memphisdev/memphis/issues?q=is%3Aissue+is%3Aclosed"><img src="https://img.shields.io/github/issues-closed/memphisdev/memphis?color=6557ff"></a> 
  <img src="https://img.shields.io/npm/dw/memphis-dev?color=ffc633&label=installations">
<a href="https://github.com/memphisdev/memphis/blob/master/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Code%20of%20Conduct-v1.0-ff69b4.svg?color=ffc633" alt="Code Of Conduct"></a> 
<a href="https://docs.memphis.dev/memphis/release-notes/releases/v1.2.0-latest"><img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/memphisdev/memphis?color=61dfc6"></a>
<img src="https://img.shields.io/github/last-commit/memphisdev/memphis?color=61dfc6&label=last%20commit">
</p>

Memphis.dev is more than a broker. It's a new streaming stack.<br>

It significantly accelerates the development of real-time applications that<br>require a streaming platform with high throughput, low latency, easy troubleshooting, fast time-to-value, <br>minimal platform operations, and all the observability you can think of.

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

### Authenticate

First, you have to authenticate to get a JWT token.\
The default expiration time is 15 minutes.

#### Example:

* Cloud: Your REST GW URL can be found within a station->code examples
* OS: The REST GW will be deploy by default, and as with any other service, should be exposed based on your deployment type.

```bash
curl --location --request POST 'https://REST_GW_URL:4444/auth/authenticate' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "root",
    // "connection_token": "memphis", // OS Only: In case the chosen auth method is connection_token
    "password": "memphis, // OS + Cloud: client-type user password
    "account_id": 123456789, // Cloud only, in case you don't have the ability to set this field as a body param you can add it as a query string param, for example: https://<REST-GW-ADDRESS>/auth/authenticate?accountId=123456789
    "token_expiry_in_minutes": 60,
    "refresh_token_expiry_in_minutes": 10000092
}'
```

Expected output:&#x20;

```JSON
{"expires_in":3600000,"jwt":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzQ3MTg0MjV9._A-fRI78fPPHL6eUFoWZjp21UYVcjXwGWiYtacYPZR8","jwt_refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjIyNzQ3MjAzNDV9.d89acaIr4CaBp7csm-jmJv0J45YrD_slvlEOKu2rs7Q","refresh_token_expires_in":600005520000}
```

#### Parameters

`username`: Memphis application-type username\
`connection_token`: Memphis application-type connection token\
`token_expiry_in_minutes`: Initial token expiration time.\
`refresh_token_expiry_in_minutes`: When should

### Refresh Token

Before the JWT token expires, you must call the refresh token to get a new one, or after authentication failure.\
The refresh JWT is valid by default for 5 hours.

#### Example:

```bash
curl --location --request POST 'rest_gateway:4444/auth/refreshToken' \
--header 'Content-Type: application/json' \
--data-raw '{
    "jwt_refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjIyNzQ3MjA2NjB9.Furfr5EZlBlglVPSjtU4x02z_jbWhu5pIByhCRh6FU8",
    "token_expiry_in_minutes": 60,
    "refresh_token_expiry_in_minutes": 10000092
}'
```

Expected output:

```json
{"expires_in":3600000,"jwt":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzQ3MTg3NTF9.EO5ersr0kQxQNRI0XlbqzOryt-F1-MmFGXRKn2sM8Yw","jwt_refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjIyNzQ3MjA2NzF9.E621wF_ieC-9rq4IgrsqYMPApAPS8YDgkT8R-69-Y5E","refresh_token_expires_in":600005520000}
```

### Produce a single message

Attach the JWT token to every request.\
JWT token as '`Bearer`' as a header.

#### Supported content types:

* text
* application/json
* application/x-protobuf

#### Example:

```bash
curl --location --request POST 'rest_gateway:4444/stations/<station_name>/produce/single' \
--header 'Authorization: Bearer eyJhbGciOiJIU**********.e30.4KOGRhUaqvm-qSHnmMwX5VrLKsvHo33u3UdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '{"message": "New Message"}'
```

#### If you don't have the option to add the authorization header, you can send the JWT via query parameters:

```bash
curl --location --request POST 'rest_gateway:4444/stations/<station_name>/produce/single?authorization=eyJhbGciOiJIU**********.e30.4KOGRhUaqvm-qSHnmMwX5VrLKsvHo33u3UdJ0qYP0kI' \
--header 'Content-Type: application/json' \
--data-raw '{"message": "New Message"}'
```

Expected output:

```json
{"error":null,"success":true}
```

#### Error Example:

```json
{"error":"Schema validation has failed: jsonschema: '' does not validate with file:///Users/user/memphisdev/memphis-rest-gateway/123#/required: missing properties: 'field1', 'field2', 'field3'","success":false}
```

### Produce a batch of messages&#x20;

Attach the JWT token to every request.\
JWT token as '`Bearer`' as a header.

#### Supported content types:

* application/json

#### Example:

```bash
curl --location --request POST 'rest_gateway:4444/stations/<station_name>/produce/batch' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.4KOGRhUaqvm-qSHnmMwX5VrLKsvHo33u3UdJ0qYP0kI' \
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

#### Error Examples:

```json
{"errors":["Schema validation has failed: jsonschema: '' does not validate with file:///Users/user/memphisdev/memphis-rest-gateway/123#/required: missing properties: 'field1'","Schema validation has failed: jsonschema: '' does not validate with file:///Users/user/memphisdev/memphis-rest-gateway/123#/required: missing properties: 'field1'"],"fail":2,"sent":1,"success":false}
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
