# Zrok.ShareRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**envZId** | **String** |  | [optional] 
**shareMode** | **String** |  | [optional] 
**frontendSelection** | **[String]** |  | [optional] 
**backendMode** | **String** |  | [optional] 
**backendProxyEndpoint** | **String** |  | [optional] 
**authScheme** | **String** |  | [optional] 
**authUsers** | [**[AuthUser]**](AuthUser.md) |  | [optional] 
**oauthProvider** | **String** |  | [optional] 
**oauthEmailDomains** | **[String]** |  | [optional] 
**oauthAuthorizationCheckInterval** | **String** |  | [optional] 
**reserved** | **Boolean** |  | [optional] 
**uniqueName** | **String** |  | [optional] 



## Enum: ShareModeEnum


* `public` (value: `"public"`)

* `private` (value: `"private"`)





## Enum: BackendModeEnum


* `proxy` (value: `"proxy"`)

* `web` (value: `"web"`)

* `tcpTunnel` (value: `"tcpTunnel"`)

* `udpTunnel` (value: `"udpTunnel"`)

* `caddy` (value: `"caddy"`)

* `drive` (value: `"drive"`)

* `socks` (value: `"socks"`)





## Enum: OauthProviderEnum


* `github` (value: `"github"`)

* `google` (value: `"google"`)




