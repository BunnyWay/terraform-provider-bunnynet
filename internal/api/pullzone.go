package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Pullzone struct {
	Id                  int64  `json:"Id,omitempty"`
	Name                string `json:"Name,omitempty"`
	CnameDomain         string `json:"CnameDomain,omitempty"`
	DisableLetsEncrypt  bool   `json:"DisableLetsEncrypt"`
	UseBackgroundUpdate bool   `json:"UseBackgroundUpdate"`

	// headers
	EnableAccessControlOriginHeader     bool     `json:"EnableAccessControlOriginHeader"`
	AccessControlOriginHeaderExtensions []string `json:"AccessControlOriginHeaderExtensions"`
	AddCanonicalHeader                  bool     `json:"AddCanonicalHeader"`

	// caching
	EnableSmartCache                   bool     `json:"EnableSmartCache"`
	CacheControlMaxAgeOverride         int64    `json:"CacheControlMaxAgeOverride"`
	CacheControlPublicMaxAgeOverride   int64    `json:"CacheControlPublicMaxAgeOverride"`
	EnableQueryStringOrdering          bool     `json:"EnableQueryStringOrdering"`
	CacheErrorResponses                bool     `json:"CacheErrorResponses"`
	IgnoreQueryStrings                 bool     `json:"IgnoreQueryStrings"`
	EnableWebPVary                     bool     `json:"EnableWebPVary"`
	EnableCountryCodeVary              bool     `json:"EnableCountryCodeVary"`
	EnableHostnameVary                 bool     `json:"EnableHostnameVary"`
	EnableMobileVary                   bool     `json:"EnableMobileVary"`
	EnableAvifVary                     bool     `json:"EnableAvifVary"`
	EnableCookieVary                   bool     `json:"EnableCookieVary"`
	QueryStringVaryParameters          []string `json:"QueryStringVaryParameters"`
	CookieVaryParameters               []string `json:"CookieVaryParameters"`
	DisableCookies                     bool     `json:"DisableCookies"`
	EnableCacheSlice                   bool     `json:"EnableCacheSlice"`
	UseStaleWhileUpdating              bool     `json:"UseStaleWhileUpdating"`
	UseStaleWhileOffline               bool     `json:"UseStaleWhileOffline"`
	PermaCacheStorageZoneId            uint64   `json:"PermaCacheStorageZoneId"`
	EnableOriginShield                 bool     `json:"EnableOriginShield"`
	OriginShieldEnableConcurrencyLimit bool     `json:"OriginShieldEnableConcurrencyLimit"`
	OriginShieldMaxConcurrentRequests  uint64   `json:"OriginShieldMaxConcurrentRequests"`
	OriginShieldMaxQueuedRequests      uint64   `json:"OriginShieldMaxQueuedRequests"`
	OriginShieldQueueMaxWaitTime       uint64   `json:"OriginShieldQueueMaxWaitTime"`
	EnableRequestCoalescing            bool     `json:"EnableRequestCoalescing"`
	RequestCoalescingTimeout           uint64   `json:"RequestCoalescingTimeout"`

	// security
	BlockRootPathAccess             bool     `json:"BlockRootPathAccess"`
	BlockPostRequests               bool     `json:"BlockPostRequests"`
	AllowedReferrers                []string `json:"AllowedReferrers"`
	BlockedReferrers                []string `json:"BlockedReferrers"`
	BlockNoneReferrer               bool     `json:"BlockNoneReferrer"`
	BlockedIps                      []string `json:"BlockedIps"`
	EnableLogging                   bool     `json:"EnableLogging"`
	LoggingIPAnonymizationEnabled   bool     `json:"LoggingIPAnonymizationEnabled"`
	LogAnonymizationType            uint8    `json:"LogAnonymizationType"`
	LogForwardingEnabled            bool     `json:"LogForwardingEnabled"`
	LogForwardingHostname           string   `json:"LogForwardingHostname"`
	LogForwardingPort               uint16   `json:"LogForwardingPort"`
	LogForwardingToken              string   `json:"LogForwardingToken"`
	LogForwardingProtocol           uint8    `json:"LogForwardingProtocol"`
	LogForwardingFormat             uint8    `json:"LogForwardingFormat"`
	LoggingSaveToStorage            bool     `json:"LoggingSaveToStorage"`
	LoggingStorageZoneId            uint64   `json:"LoggingStorageZoneId"`
	EnableTLS1                      bool     `json:"EnableTLS1"`
	EnableTLS1_1                    bool     `json:"EnableTLS1_1"`
	ErrorPageWhitelabel             bool     `json:"ErrorPageWhitelabel"`
	ErrorPageEnableStatuspageWidget bool     `json:"ErrorPageEnableStatuspageWidget"`
	ErrorPageStatuspageCode         string   `json:"ErrorPageStatuspageCode"`
	ErrorPageEnableCustomCode       bool     `json:"ErrorPageEnableCustomCode"`
	ErrorPageCustomCode             string   `json:"ErrorPageCustomCode"`
	AWSSigningEnabled               bool     `json:"AWSSigningEnabled"`
	AWSSigningKey                   string   `json:"AWSSigningKey"`
	AWSSigningSecret                string   `json:"AWSSigningSecret"`
	AWSSigningRegionName            string   `json:"AWSSigningRegionName"`
	ZoneSecurityEnabled             bool     `json:"ZoneSecurityEnabled"`
	ZoneSecurityIncludeHashRemoteIP bool     `json:"ZoneSecurityIncludeHashRemoteIp"`
	ZoneSecurityKey                 string   `json:"ZoneSecurityKey"`

	// safe hop
	EnableSafeHop                bool   `json:"EnableSafeHop"`
	OriginRetries                uint8  `json:"OriginRetries"`
	OriginRetryDelay             uint64 `json:"OriginRetryDelay"`
	OriginRetryConnectionTimeout bool   `json:"OriginRetryConnectionTimeout"`
	OriginRetry5XXResponses      bool   `json:"OriginRetry5XXResponses"`
	OriginRetryResponseTimeout   bool   `json:"OriginRetryResponseTimeout"`
	OriginConnectTimeout         uint64 `json:"OriginConnectTimeout"`
	OriginResponseTimeout        uint64 `json:"OriginResponseTimeout"`

	// sub-resources
	Edgerules        []PullzoneEdgerule       `json:"Edgerules"`
	Hostnames        []PullzoneHostname       `json:"Hostnames"`
	OptimizerClasses []PullzoneOptimizerClass `json:"OptimizerClasses"`

	// origin
	OriginType       uint8  `json:"OriginType"`
	OriginUrl        string `json:"OriginUrl,omitempty"`
	StorageZoneId    int64  `json:"StorageZoneId,omitempty"`
	OriginHostHeader string `json:"OriginHostHeader,omitempty"`
	AddHostHeader    bool   `json:"AddHostHeader"`
	VerifyOriginSSL  bool   `json:"VerifyOriginSSL"`
	FollowRedirects  bool   `json:"FollowRedirects"`

	// routing
	Type                      uint8    `json:"Type"`
	EnableGeoZoneAF           bool     `json:"EnableGeoZoneAF"`
	EnableGeoZoneASIA         bool     `json:"EnableGeoZoneASIA"`
	EnableGeoZoneEU           bool     `json:"EnableGeoZoneEU"`
	EnableGeoZoneSA           bool     `json:"EnableGeoZoneSA"`
	EnableGeoZoneUS           bool     `json:"EnableGeoZoneUS"`
	RoutingFilters            []string `json:"RoutingFilters"`
	BudgetRedirectedCountries []string `json:"BudgetRedirectedCountries"`
	BlockedCountries          []string `json:"BlockedCountries"`

	// optimizer
	OptimizerEnabled                      bool    `json:"OptimizerEnabled"`
	OptimizerMinifyCss                    bool    `json:"OptimizerMinifyCSS"`
	OptimizerMinifyJs                     bool    `json:"OptimizerMinifyJavaScript"`
	OptimizerWebp                         bool    `json:"OptimizerEnableWebP"`
	OptimizerForceClasses                 bool    `json:"OptimizerForceClasses"`
	OptimizerImageOptimization            bool    `json:"OptimizerEnableManipulationEngine"`
	OptimizerAutomaticOptimizationEnabled bool    `json:"OptimizerAutomaticOptimizationEnabled"`
	OptimizerDesktopMaxWidth              uint64  `json:"OptimizerDesktopMaxWidth"`
	OptimizerMobileMaxWidth               uint64  `json:"OptimizerMobileMaxWidth"`
	OptimizerImageQuality                 uint8   `json:"OptimizerImageQuality"`
	OptimizerMobileImageQuality           uint8   `json:"OptimizerMobileImageQuality"`
	OptimizerWatermarkEnabled             bool    `json:"OptimizerWatermarkEnabled"`
	OptimizerWatermarkUrl                 string  `json:"OptimizerWatermarkUrl"`
	OptimizerWatermarkPosition            uint8   `json:"OptimizerWatermarkPosition"`
	OptimizerWatermarkOffset              float64 `json:"OptimizerWatermarkOffset"`
	OptimizerWatermarkMinImageSize        uint64  `json:"OptimizerWatermarkMinImageSize"`

	// limits
	LimitRatePerSecond        float64 `json:"LimitRatePerSecond"`
	RequestLimit              uint64  `json:"RequestLimit"`
	LimitRateAfter            float64 `json:"LimitRateAfter"`
	BurstSize                 uint64  `json:"BurstSize"`
	ConnectionLimitPerIPCount uint64  `json:"ConnectionLimitPerIPCount"`
	MonthlyBandwidthLimit     uint64  `json:"MonthlyBandwidthLimit"`
}

func (c *Client) GetPullzone(id int64) (Pullzone, error) {
	var data Pullzone
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/pullzone/%d", c.apiUrl, id), nil)
	if err != nil {
		return data, err
	}

	if resp.StatusCode != http.StatusOK {
		return data, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func (c *Client) CreatePullzone(data Pullzone) (Pullzone, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return Pullzone{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return Pullzone{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return Pullzone{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return Pullzone{}, err
		}

		return Pullzone{}, errors.New(obj.Message)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return Pullzone{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := Pullzone{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) UpdatePullzone(dataApi Pullzone) (Pullzone, error) {
	id := dataApi.Id

	body, err := json.Marshal(dataApi)
	if err != nil {
		return Pullzone{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d", c.apiUrl, id), bytes.NewReader(body))
	if err != nil {
		return Pullzone{}, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return Pullzone{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return Pullzone{}, err
		}

		return Pullzone{}, errors.New(obj.Message)
	}

	dataApiResult, err := c.GetPullzone(id)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeletePullzone(id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/pullzone/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
