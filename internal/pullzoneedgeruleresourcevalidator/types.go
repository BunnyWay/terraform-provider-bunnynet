package pullzoneedgeruleresourcevalidator

var ActionMap = map[uint8]string{
	0:  "ForceSSL",
	1:  "Redirect",
	2:  "OriginUrl",
	3:  "OverrideCacheTime",
	4:  "BlockRequest",
	5:  "SetResponseHeader",
	6:  "SetRequestHeader",
	7:  "ForceDownload",
	8:  "DisableTokenAuthentication",
	9:  "EnableTokenAuthentication",
	10: "OverrideCacheTimePublic",
	11: "IgnoreQueryString",
	12: "DisableOptimizer",
	13: "ForceCompression",
	14: "SetStatusCode",
	15: "BypassPermaCache",
	16: "OverrideBrowserCacheTime",
	17: "OriginStorage",
	18: "SetNetworkRateLimit",
	19: "SetConnectionLimit",
	20: "SetRequestsPerSecondLimit",
}

var TriggerTypeMap = map[uint8]string{
	0:  "Url",
	1:  "RequestHeader",
	2:  "ResponseHeader",
	3:  "UrlExtension",
	4:  "CountryCode",
	5:  "RemoteIP",
	6:  "UrlQueryString",
	7:  "RandomChance",
	8:  "StatusCode",
	9:  "RequestMethod",
	10: "CookieValue",
	11: "CountryStateCode",
}

var TriggerMatchTypeMap = map[uint8]string{
	0: "MatchAny",
	1: "MatchAll",
	2: "MatchNone",
}
