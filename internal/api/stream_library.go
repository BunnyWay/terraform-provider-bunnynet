// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
)

type StreamLibrary struct {
	Id                                      int64    `json:"Id,omitempty"`
	Name                                    string   `json:"Name"`
	UILanguage                              string   `json:"UILanguage"`
	FontFamily                              string   `json:"FontFamily"`
	PlayerKeyColor                          string   `json:"PlayerKeyColor"`
	CaptionsFontColor                       string   `json:"CaptionsFontColor"`
	CaptionsFontSize                        uint16   `json:"CaptionsFontSize"`
	CaptionsBackground                      string   `json:"CaptionsBackground"`
	ShowHeatmap                             bool     `json:"ShowHeatmap"`
	Controls                                string   `json:"Controls"`
	CustomHTML                              *string  `json:"CustomHTML"`
	VastTagUrl                              string   `json:"VastTagUrl"`
	KeepOriginalFiles                       bool     `json:"KeepOriginalFiles"`
	AllowEarlyPlay                          bool     `json:"AllowEarlyPlay"`
	EnableContentTagging                    bool     `json:"EnableContentTagging"`
	EnableMP4Fallback                       bool     `json:"EnableMP4Fallback"`
	EnabledResolutions                      string   `json:"EnabledResolutions"`
	Bitrate240P                             uint32   `json:"Bitrate240p"`
	Bitrate360P                             uint32   `json:"Bitrate360p"`
	Bitrate480P                             uint32   `json:"Bitrate480p"`
	Bitrate720P                             uint32   `json:"Bitrate720p"`
	Bitrate1080P                            uint32   `json:"Bitrate1080p"`
	Bitrate1440P                            uint32   `json:"Bitrate1440p"`
	Bitrate2160P                            uint32   `json:"Bitrate2160p"`
	WatermarkPositionLeft                   uint8    `json:"WatermarkPositionLeft"`
	WatermarkPositionTop                    uint8    `json:"WatermarkPositionTop"`
	WatermarkWidth                          uint16   `json:"WatermarkWidth"`
	WatermarkHeight                         uint16   `json:"WatermarkHeight"`
	EnableTranscribing                      bool     `json:"EnableTranscribing"`
	EnableTranscribingTitleGeneration       bool     `json:"EnableTranscribingTitleGeneration"`
	EnableTranscribingDescriptionGeneration bool     `json:"EnableTranscribingDescriptionGeneration"`
	TranscribingCaptionLanguages            []string `json:"TranscribingCaptionLanguages"`
	AllowDirectPlay                         bool     `json:"AllowDirectPlay"`
	AllowedReferrers                        []string `json:"AllowedReferrers"`
	BlockedReferrers                        []string `json:"BlockedReferrers"`
	BlockNoneReferrer                       bool     `json:"BlockNoneReferrer"`
	PlayerTokenAuthenticationEnabled        bool     `json:"PlayerTokenAuthenticationEnabled"`
	EnableTokenAuthentication               bool     `json:"EnableTokenAuthentication"`
	EnableDRM                               bool     `json:"EnableDRM"`
	WebhookUrl                              *string  `json:"WebhookUrl"`
	PullZoneId                              int64    `json:"PullZoneId"`
	StorageZoneId                           int64    `json:"StorageZoneId"`
	ApiKey                                  string   `json:"ApiKey"`
}

func (c *Client) GetStreamLibrary(id int64) (StreamLibrary, error) {
	var data StreamLibrary
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/videoLibrary/%d", c.apiUrl, id), nil)
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

	pullzone, err := c.GetPullzone(data.PullZoneId)
	if err != nil {
		return data, err
	}

	data.EnableTokenAuthentication = pullzone.ZoneSecurityEnabled

	return data, nil
}

func (c *Client) CreateStreamLibrary(data StreamLibrary) (StreamLibrary, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return StreamLibrary{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/videoLibrary", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return StreamLibrary{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return StreamLibrary{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return StreamLibrary{}, err
		}

		return StreamLibrary{}, errors.New(obj.Message)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return StreamLibrary{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := StreamLibrary{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) UpdateStreamLibrary(dataApi StreamLibrary) (StreamLibrary, error) {
	id := dataApi.Id

	body, err := json.Marshal(dataApi)
	if err != nil {
		return StreamLibrary{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/videoLibrary/%d", c.apiUrl, id), bytes.NewReader(body))
	if err != nil {
		return StreamLibrary{}, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return StreamLibrary{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return StreamLibrary{}, err
		}

		return StreamLibrary{}, errors.New(obj.Message)
	}

	// update EnableTokenAuthentication
	{
		body, err := json.Marshal(map[string]bool{
			"EnableTokenAuthentication": dataApi.EnableTokenAuthentication,
		})
		if err != nil {
			return StreamLibrary{}, err
		}

		resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/videoLibrary/%d", c.apiUrl, id), bytes.NewReader(body))
		if err != nil {
			return StreamLibrary{}, err
		}

		if resp.StatusCode != http.StatusOK {
			bodyResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return StreamLibrary{}, err
			}

			_ = resp.Body.Close()
			var obj struct {
				Message string `json:"Message"`
			}

			err = json.Unmarshal(bodyResp, &obj)
			if err != nil {
				return StreamLibrary{}, err
			}

			return StreamLibrary{}, errors.New(obj.Message)
		}
	}

	dataApiResult, err := c.GetStreamLibrary(id)
	if err != nil {
		return dataApiResult, err
	}

	reloadResult := false
	if !slices.Equal(dataApi.AllowedReferrers, dataApiResult.AllowedReferrers) {
		diff := utils.SliceDiff(dataApi.AllowedReferrers, dataApiResult.AllowedReferrers)
		if len(diff) > 0 {
			for _, hostname := range diff {
				err = c.streamLibraryRefererAddRemove(id, hostname, "Allowed", "add")
				if err != nil {
					return dataApiResult, err
				}
			}

			reloadResult = true
		}

		diff = utils.SliceDiff(dataApiResult.AllowedReferrers, dataApi.AllowedReferrers)
		if len(diff) > 0 {
			for _, hostname := range diff {
				err = c.streamLibraryRefererAddRemove(id, hostname, "Allowed", "remove")
				if err != nil {
					return dataApiResult, err
				}
			}

			reloadResult = true
		}
	}

	if !slices.Equal(dataApi.BlockedReferrers, dataApiResult.BlockedReferrers) {
		diff := utils.SliceDiff(dataApi.BlockedReferrers, dataApiResult.BlockedReferrers)
		if len(diff) > 0 {
			for _, hostname := range diff {
				err = c.streamLibraryRefererAddRemove(id, hostname, "Blocked", "add")
				if err != nil {
					return dataApiResult, err
				}
			}

			reloadResult = true
		}

		diff = utils.SliceDiff(dataApiResult.BlockedReferrers, dataApi.BlockedReferrers)
		if len(diff) > 0 {
			for _, hostname := range diff {
				err = c.streamLibraryRefererAddRemove(id, hostname, "Blocked", "remove")
				if err != nil {
					return dataApiResult, err
				}
			}

			reloadResult = true
		}
	}

	if reloadResult {
		dataApiResult, err = c.GetStreamLibrary(id)
		if err != nil {
			return dataApiResult, err
		}
	}

	return dataApiResult, nil
}

func (c *Client) DeleteStreamLibrary(id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/videoLibrary/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) streamLibraryRefererAddRemove(id int64, hostname string, refType string, method string) error {
	body, err := json.Marshal(map[string]string{
		"Hostname": hostname,
	})

	if err != nil {
		return err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/videoLibrary/%d/%s%sReferrer", c.apiUrl, id, method, refType), bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("StreamLibrary %d %s%sReferrer failed: %s", id, method, refType, resp.Status)
	}

	return nil
}
