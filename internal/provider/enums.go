// This file was generated via "go generate". DO NOT EDIT.
package provider

var pullzoneShieldRatelimitRuleLimitTimeframeOptions = []int64{
	1,
	10,
}

var pullzoneShieldRatelimitRuleResponseTimeframeOptions = []int64{
	30,
	60,
	300,
	900,
	1800,
	3600,
}

var pullzoneShieldRuleConditionOperationMap = map[int64]string{
	0:  "BEGINSWITH",
	1:  "ENDSWITH",
	2:  "CONTAINS",
	3:  "CONTAINSWORD",
	4:  "STRMATCH",
	5:  "EQ",
	6:  "GE",
	7:  "GT",
	8:  "LE",
	9:  "LT",
	12: "WITHIN",
	14: "RX",
	15: "STREQ",
	17: "DETECTSQLI",
	18: "DETECTXSS",
}

var pullzoneShieldRuleConditionVariableMap = map[uint8]string{
	0:  "REQUEST_URI",
	1:  "REQUEST_URI_RAW",
	2:  "ARGS",
	3:  "ARGS_COMBINED_SIZE",
	4:  "ARGS_GET",
	5:  "ARGS_GET_NAMES",
	6:  "ARGS_POST",
	7:  "ARGS_POST_NAMES",
	8:  "FILES_NAMES",
	9:  "GEO",
	10: "REMOTE_ADDR",
	11: "QUERY_STRING",
	12: "REQUEST_BASENAME",
	13: "REQUEST_BODY",
	14: "REQUEST_COOKIES_NAMES",
	15: "REQUEST_COOKIES",
	16: "REQUEST_FILENAME",
	17: "REQUEST_HEADERS_NAMES",
	18: "REQUEST_HEADERS",
	19: "REQUEST_LINE",
	20: "REQUEST_METHOD",
	21: "REQUEST_PROTOCOL",
	22: "RESPONSE_BODY",
	23: "RESPONSE_HEADERS",
	24: "RESPONSE_STATUS",
}

var pullzoneShieldRuleTransformationMap = map[int64]string{
	1:  "CMDLINE",
	2:  "COMPRESSWHITESPACE",
	3:  "CSSDECODE",
	4:  "HEXENCODE",
	5:  "HTMLENTITYDECODE",
	6:  "JSDECODE",
	7:  "LENGTH",
	8:  "LOWERCASE",
	9:  "MD5",
	10: "NORMALIZEPATH",
	11: "NORMALISEPATH",
	12: "NORMALIZEPATHWIN",
	13: "NORMALISEPATHWIN",
	14: "REMOVECOMMENTS",
	15: "REMOVENULLS",
	16: "REMOVEWHITESPACE",
	17: "REPLACECOMMENTS",
	18: "SHA1",
	19: "URLDECODE",
	20: "URLDECODEUNI",
	21: "UTF8TOUNICODE",
}

var pullzoneShieldWafBodyLimitMap = map[uint8]string{
	0: "Block",
	1: "Log",
	2: "Ignore",
}

var pullzoneShieldWafRuleResponseActionMap = map[uint8]string{
	1: "Block",
	2: "Log",
	3: "Challenge",
	4: "Allow",
	5: "Bypass",
}

var pullzoneAccessListActionMap = map[uint8]string{
	1: "Allow",
	2: "Block",
	3: "Challenge",
	4: "Log",
	5: "Bypass",
}

var pullzoneAccessListTypeMap = map[uint8]string{
	0: "IP",
	1: "CIDR",
	2: "ASN",
	3: "Country",
}
