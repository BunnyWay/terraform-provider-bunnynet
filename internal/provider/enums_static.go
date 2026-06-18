package provider

import "github.com/bunnyway/terraform-provider-bunnynet/internal/api"

var accountSubuserPermissionsMap = map[string]string{
	"zones":   "SubuserManage",
	"billing": "SubuserBilling",
	"support": "SubuserSupport",
	"abuse":   "SubuserAbuse",
	"users":   "SubuserUsers",
}

var computeContainerAppImagePullPolicyOptions = []string{"Always", "IfNotPresent"}

// format: key: terraform, value: api
var computeContainerAppContainerEndpointTypeMap = map[string]string{
	"CDN":        "CDN",
	"Anycast":    "Anycast",
	"InternalIP": "PublicIp", // @TODO replace with InternalIp
}

// format: key: terraform, value: api
var computeContainerAppContainerEndpointProtocolMap = map[string]string{
	"TCP":  "Tcp",
	"UDP":  "Udp",
	"SCTP": "Sctp",
}

var containerAppContainerProbeTypeOptions = []string{
	"http",
	"tcp",
	"grpc",
}

var computeContainerAppContainerProbeHttpStatusMap = map[int64]string{
	100: "Continue",
	101: "SwitchingProtocols",
	102: "Processing",
	103: "EarlyHints",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "NonAuthoritativeInformation",
	204: "NoContent",
	205: "ResetContent",
	206: "PartialContent",
	207: "MultiStatus",
	208: "AlreadyReported",
	226: "IMUsed",
	300: "MultipleChoices",
	301: "MovedPermanently",
	302: "Found",
	303: "SeeOther",
	304: "NotModified",
	305: "UseProxy",
	306: "Unused",
	307: "TemporaryRedirect",
	308: "PermanentRedirect",
	400: "BadRequest",
	401: "Unauthorized",
	402: "PaymentRequired",
	403: "Forbidden",
	404: "NotFound",
	405: "MethodNotAllowed",
	406: "NotAcceptable",
	407: "ProxyAuthenticationRequired",
	408: "RequestTimeout",
	409: "Conflict",
	410: "Gone",
	411: "LengthRequired",
	412: "PreconditionFailed",
	413: "RequestEntityTooLarge",
	414: "RequestUriTooLong",
	415: "UnsupportedMediaType",
	416: "RequestedRangeNotSatisfiable",
	417: "ExpectationFailed",
	421: "MisdirectedRequest",
	422: "UnprocessableEntity",
	423: "Locked",
	424: "FailedDependency",
	426: "UpgradeRequired",
	428: "PreconditionRequired",
	429: "TooManyRequests",
	431: "RequestHeaderFieldsTooLarge",
	451: "UnavailableForLegalReasons",
	500: "InternalServerError",
	501: "NotImplemented",
	502: "BadGateway",
	503: "ServiceUnavailable",
	504: "GatewayTimeout",
	505: "HttpVersionNotSupported",
	506: "VariantAlsoNegotiates",
	507: "InsufficientStorage",
	508: "LoopDetected",
	510: "NotExtended",
	511: "NetworkAuthenticationRequired",
}

var computeContainerImageRegistryRegistryOptions = []string{"GitHub", "DockerHub"}

var computeScriptTypeMap = map[uint8]string{
	api.ScriptTypeStandalone: "standalone",
	api.ScriptTypeMiddleware: "middleware",
}

var pullzoneOriginTypeMap = map[uint8]string{
	0: "OriginUrl",
	1: "DnsAccelerate",
	2: "StorageZone",
	4: "ComputeScript",
	5: "ComputeContainer",
}

var pullzoneRoutingTierMap = map[uint8]string{
	0: "Standard",
	1: "Volume",
}

var pullzoneCacheVaryOptions = []string{"querystring", "webp", "country", "state", "hostname", "mobile", "avif", "cookie"}
var pullzoneCacheStaleOptions = []string{"offline", "updating"}
var pullzoneTlsSupportOptions = []string{"TLSv1.0", "TLSv1.1"}
var pullzoneSafehopRetryReasonsOptions = []string{"connectionTimeout", "5xxResponse", "responseTimeout"}
var pullzoneRoutingZonesOptions = []string{"AF", "ASIA", "EU", "SA", "US"}
var pullzoneRoutingFiltersOptions = []string{"all", "eu", "scripting"}
var pullzoneOriginShieldZoneOptions = []string{"IL", "FR"}

var pullzoneShieldDdosLevelMap = map[uint8]string{
	0: "Asleep",
	1: "Low",
	2: "Medium",
	3: "High",
	4: "Extreme",
}

var pullzoneShieldDdosModeMap = map[uint8]string{
	0: "Log",
	1: "Block",
}

var pullzoneShieldBotDetectionModeMap = map[uint8]string{
	0: "Log",
	1: "Challenge",
}

var pullzoneShieldWafModeMap = map[uint8]string{
	0: "Log",
	1: "Block",
}

var pullzoneShieldWafAllowedHttpVersionsOptions = []string{
	"HTTP/1.0",
	"HTTP/1.1",
	"HTTP/2",
	"HTTP/2.0",
}

var pullzoneShieldWafAllowedHttpMethodsOptions = []string{
	"GET",
	"HEAD",
	"POST",
	"PUT",
	"PATCH",
	"OPTIONS",
	"DELETE",
	"CONNECT",
	"TRACE",
}

var storageZoneTypeMap = map[uint8]string{
	0: "Standard",
	1: "S3",
}

var streamLibraryFontFamilyOptions = []string{"arial", "inter", "lato", "oswald", "raleway", "roboto", "rubik", "ubuntu"}
var streamLibraryPlayerControlsOptions = []string{"airplay", "captions", "chromecast", "current-time", "duration", "fast-forward", "fullscreen", "mute", "pip", "play", "play-large", "progress", "rewind", "settings", "volume"}
var streamLibraryOutputCodecsOptions = []string{"x264", "vp9", "hevc", "av1"}
