// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/memorycache"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"slices"
)

const LoadbalancerOriginTypeOriginUrl = 0
const LoadbalancerOriginTypeStorageZone = 1
const LoadbalancerOriginTypeComputeScript = 2
const LoadbalancerOriginTypeComputeContainer = 3

type Loadbalancer struct {
	Id            int64  `json:"LoadBalancerId,omitempty"`
	Name          string `json:"Name"`
	Method        uint8  `json:"OriginGroupLoadBalancingMethod"`
	StickySession uint8  `json:"StickySessionType"`

	Groups []LoadbalancerGroup `json:"OriginGroups,omitempty"`
}

type LoadbalancerGroupHealthcheck struct {
	Enabled bool   `json:"Enabled"`
	Path    string `json:"Path"`
}

type LoadbalancerGroup struct {
	Id             int64 `json:"Id,omitempty"`
	LoadbalancerId int64 `json:"-"`
	Priority       uint8 `json:"Priority"`

	Enabled   bool    `json:"Enabled"`
	Name      string  `json:"Name"`
	Weight    uint8   `json:"Weight,omitempty"`
	Method    uint8   `json:"OriginLoadBalancingMethod"`
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`

	HealthCheck LoadbalancerGroupHealthcheck `json:"HealthCheckSettings"`
	Origins     []LoadbalancerOrigin         `json:"Origins,omitempty"`
}

type LoadbalancerOriginOriginUrl struct {
	OriginUrl  string `json:"OriginUrl"`
	HostHeader string `json:"HostHeader"`
	VerifySsl  bool   `json:"VerifySsl"`
}

type LoadbalancerOriginOriginMC struct {
	AppId      string `json:"AppId"`
	EndpointId string `json:"EndPointId"`
}

type LoadbalancerOrigin struct {
	Id             int64 `json:"Id,omitempty"`
	LoadbalancerId int64 `json:"-"`
	GroupId        int64 `json:"-"`
	Weight         uint8 `json:"Weight,omitempty"`

	Name       string `json:"Name"`
	Enabled    bool   `json:"Enabled"`
	OriginType uint8  `json:"OriginType"`

	OriginUrl             *LoadbalancerOriginOriginUrl `json:"StandardOrigin,omitempty"`
	EdgeScriptId          *int64                       `json:"EdgeScriptId,omitempty"`
	StorageZoneId         *int64                       `json:"StorageZoneId,omitempty"`
	MagicContainersOrigin *LoadbalancerOriginOriginMC  `json:"MagicContainersOrigin,omitempty"`
}

var loadbalancerCache = memorycache.New[int64, Loadbalancer]()
var loadbalancerMutex = utils.NewIdMutex()

func (c *Client) GetLoadbalancer(ctx context.Context, id int64) (Loadbalancer, error) {
	loadbalancerMutex.Lock(id)
	defer loadbalancerMutex.Unlock(id)

	if v, ok := loadbalancerCache.Get(id); ok {
		tflog.Warn(ctx, "loadbalancerCache.Get()")
		return *v, nil
	}

	var data Loadbalancer
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/loadbalancer/%d", c.apiUrl, id), nil)
	if err != nil {
		return data, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return data, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return data, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	tflog.Debug(ctx, fmt.Sprintf("GET /loadbalancer/%d: %s", id, string(bodyResp)))

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return data, err
	}

	// fill IDs in sub-objects
	for i, group := range data.Groups {
		group.LoadbalancerId = data.Id

		for j, origin := range group.Origins {
			origin.LoadbalancerId = data.Id
			origin.GroupId = group.Id
			group.Origins[j] = origin
		}

		data.Groups[i] = group
	}

	loadbalancerSort(&data)

	loadbalancerCache.Set(data.Id, data)

	return data, nil
}

func (c *Client) CreateLoadbalancer(ctx context.Context, data Loadbalancer) (Loadbalancer, error) {
	data.Groups = nil

	body, err := json.Marshal(data)
	if err != nil {
		return Loadbalancer{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer: %s", string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/loadbalancer", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return Loadbalancer{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := extractErrorMessage(resp)
		if err != nil {
			return Loadbalancer{}, err
		}

		return Loadbalancer{}, errors.New("create loadbalancer failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return Loadbalancer{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := Loadbalancer{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) UpdateLoadbalancer(ctx context.Context, data Loadbalancer) (Loadbalancer, error) {
	data.Groups = nil

	body, err := json.Marshal(data)
	if err != nil {
		return Loadbalancer{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d: %s", data.Id, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/loadbalancer/%d", c.apiUrl, data.Id), bytes.NewReader(body))
	if err != nil {
		return Loadbalancer{}, err
	}

	loadbalancerCache.Delete(data.Id)

	if resp.StatusCode != http.StatusOK {
		err := extractErrorMessage(resp)
		if err != nil {
			return Loadbalancer{}, err
		}

		return Loadbalancer{}, errors.New("update loadbalancer failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	_ = resp.Body.Close()

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d: %s", data.Id, string(bodyResp)))

	dataApiResult := Loadbalancer{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeleteLoadbalancer(ctx context.Context, id int64) error {
	tflog.Debug(ctx, fmt.Sprintf("DELETE /loadbalancer/%d", id))

	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/loadbalancer/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	loadbalancerCache.Delete(id)

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) GetLoadbalancerGroup(ctx context.Context, loadbalancerId int64, groupId int64) (LoadbalancerGroup, error) {
	var data LoadbalancerGroup
	loadbalancer, err := c.GetLoadbalancer(ctx, loadbalancerId)
	if err != nil {
		return data, err
	}

	for _, group := range loadbalancer.Groups {
		if group.Id == groupId {
			group.LoadbalancerId = loadbalancer.Id
			return group, nil
		}
	}

	return data, ErrNotFound
}

func (c *Client) CreateLoadbalancerGroup(ctx context.Context, data LoadbalancerGroup) (LoadbalancerGroup, error) {
	loadbalancerMutex.Lock(data.LoadbalancerId)
	defer loadbalancerMutex.Unlock(data.LoadbalancerId)

	data.Origins = nil

	body, err := json.Marshal(data)
	if err != nil {
		return LoadbalancerGroup{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d/origingroup: %s", data.LoadbalancerId, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/loadbalancer/%d/origingroup", c.apiUrl, data.LoadbalancerId), bytes.NewReader(body))
	if err != nil {
		return data, err
	}

	loadbalancerCache.Delete(data.LoadbalancerId)

	if resp.StatusCode != http.StatusOK {
		err := extractErrorMessage(resp)
		if err != nil {
			return data, err
		}

		return data, errors.New("create loadbalancer group failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	_ = resp.Body.Close()

	loadbalancerCache.Delete(data.Id)
	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d/origingroup: %s", data.LoadbalancerId, string(bodyResp)))

	dataApiResult := LoadbalancerGroup{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	dataApiResult.LoadbalancerId = data.LoadbalancerId
	return dataApiResult, nil
}

func (c *Client) UpdateLoadbalancerGroup(ctx context.Context, data LoadbalancerGroup) (LoadbalancerGroup, error) {
	loadbalancerMutex.Lock(data.LoadbalancerId)
	defer loadbalancerMutex.Unlock(data.LoadbalancerId)

	data.Origins = nil

	body, err := json.Marshal(data)
	if err != nil {
		return data, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d/origingroup/%d: %s", data.LoadbalancerId, data.Id, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/loadbalancer/%d/origingroup/%d", c.apiUrl, data.LoadbalancerId, data.Id), bytes.NewReader(body))
	if err != nil {
		return data, err
	}

	loadbalancerCache.Delete(data.LoadbalancerId)

	if resp.StatusCode != http.StatusOK {
		err := extractErrorMessage(resp)
		if err != nil {
			return data, err
		}

		return data, errors.New("update loadbalancer group failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	_ = resp.Body.Close()

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d/origingroup/%d: %s", data.LoadbalancerId, data.Id, string(bodyResp)))

	dataApiResult := LoadbalancerGroup{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	dataApiResult.LoadbalancerId = data.LoadbalancerId
	return dataApiResult, nil
}

func (c *Client) DeleteLoadbalancerGroup(ctx context.Context, loadbalancerId int64, id int64) error {
	loadbalancerMutex.Lock(loadbalancerId)
	defer loadbalancerMutex.Unlock(loadbalancerId)

	tflog.Debug(ctx, fmt.Sprintf("DELETE /loadbalancer/%d/origingroup/%d", loadbalancerId, id))

	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/loadbalancer/%d/origingroup/%d", c.apiUrl, loadbalancerId, id), nil)
	if err != nil {
		return err
	}

	loadbalancerCache.Delete(loadbalancerId)

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) GetLoadbalancerOrigin(ctx context.Context, loadbalancerId int64, groupId int64, id int64) (LoadbalancerOrigin, error) {
	var data LoadbalancerOrigin
	group, err := c.GetLoadbalancerGroup(ctx, loadbalancerId, groupId)
	if err != nil {
		return data, err
	}

	for _, origin := range group.Origins {
		if origin.Id == id {
			return origin, nil
		}
	}

	return data, ErrNotFound
}

func (c *Client) CreateLoadbalancerOrigin(ctx context.Context, data LoadbalancerOrigin) (LoadbalancerOrigin, error) {
	loadbalancerMutex.Lock(data.LoadbalancerId)
	defer loadbalancerMutex.Unlock(data.LoadbalancerId)

	body, err := json.Marshal(loadbalancerOriginPayload(data))
	if err != nil {
		return LoadbalancerOrigin{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d/origingroup/%d/origin: %s", data.LoadbalancerId, data.GroupId, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/loadbalancer/%d/origingroup/%d/origin", c.apiUrl, data.LoadbalancerId, data.GroupId), bytes.NewReader(body))
	if err != nil {
		return LoadbalancerOrigin{}, err
	}

	loadbalancerCache.Delete(data.LoadbalancerId)

	if resp.StatusCode != http.StatusOK {
		err := extractErrorMessage(resp)
		if err != nil {
			return LoadbalancerOrigin{}, err
		}

		return LoadbalancerOrigin{}, errors.New("create loadbalancer origin failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return LoadbalancerOrigin{}, err
	}
	_ = resp.Body.Close()

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d/origingroup/%d/origin: %s", data.LoadbalancerId, data.Id, string(bodyResp)))

	dataApiResult := LoadbalancerOrigin{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	dataApiResult.LoadbalancerId = data.LoadbalancerId
	dataApiResult.GroupId = data.GroupId

	return dataApiResult, nil
}

func (c *Client) UpdateLoadbalancerOrigin(ctx context.Context, data LoadbalancerOrigin) (LoadbalancerOrigin, error) {
	loadbalancerMutex.Lock(data.LoadbalancerId)
	defer loadbalancerMutex.Unlock(data.LoadbalancerId)

	body, err := json.Marshal(loadbalancerOriginPayload(data))
	if err != nil {
		return LoadbalancerOrigin{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d/origingroup/%d/origin/%d: %s", data.LoadbalancerId, data.GroupId, data.Id, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/loadbalancer/%d/origingroup/%d/origin/%d", c.apiUrl, data.LoadbalancerId, data.GroupId, data.Id), bytes.NewReader(body))
	if err != nil {
		return LoadbalancerOrigin{}, err
	}

	loadbalancerCache.Delete(data.LoadbalancerId)

	if resp.StatusCode != http.StatusOK {
		err := extractErrorMessage(resp)
		if err != nil {
			return LoadbalancerOrigin{}, err
		}

		return LoadbalancerOrigin{}, errors.New("update loadbalancer origin failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return LoadbalancerOrigin{}, err
	}
	_ = resp.Body.Close()

	tflog.Debug(ctx, fmt.Sprintf("POST /loadbalancer/%d/origingroup/%d/origin/%d: %s", data.LoadbalancerId, data.GroupId, data.Id, string(bodyResp)))

	dataApiResult := LoadbalancerOrigin{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	dataApiResult.LoadbalancerId = data.LoadbalancerId
	dataApiResult.GroupId = data.GroupId

	return dataApiResult, nil
}

func (c *Client) DeleteLoadbalancerOrigin(ctx context.Context, loadbalancerId int64, groupId int64, id int64) error {
	loadbalancerMutex.Lock(loadbalancerId)
	defer loadbalancerMutex.Unlock(loadbalancerId)

	tflog.Debug(ctx, fmt.Sprintf("DELETE /loadbalancer/%d/origingroup/%d/origin/%d", loadbalancerId, groupId, id))

	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/loadbalancer/%d/origingroup/%d/origin/%d", c.apiUrl, loadbalancerId, groupId, id), nil)
	if err != nil {
		return err
	}

	loadbalancerCache.Delete(loadbalancerId)

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

func loadbalancerSort(loadbalancer *Loadbalancer) {
	slices.SortFunc(loadbalancer.Groups, func(a, b LoadbalancerGroup) int {
		if a.Name == b.Name {
			return 0
		}

		if a.Name < b.Name {
			return -1
		}

		return 1
	})

	for _, group := range loadbalancer.Groups {
		slices.SortFunc(group.Origins, func(a, b LoadbalancerOrigin) int {
			if a.Name == b.Name {
				return 0
			}

			if a.Name < b.Name {
				return -1
			}

			return 1
		})
	}
}

func loadbalancerOriginPayload(data LoadbalancerOrigin) map[string]interface{} {
	values := map[string]interface{}{
		"Name":       data.Name,
		"Enabled":    data.Enabled,
		"OriginType": data.OriginType,
	}

	if data.Weight > 0 {
		values["Weight"] = data.Weight
	}

	switch data.OriginType {
	case LoadbalancerOriginTypeOriginUrl:
		values["StandardOrigin"] = data.OriginUrl

	case LoadbalancerOriginTypeStorageZone:
		values["StorageZoneId"] = data.StorageZoneId

	case LoadbalancerOriginTypeComputeScript:
		values["EdgeScriptId"] = data.EdgeScriptId

	case LoadbalancerOriginTypeComputeContainer:
		values["MagicContainersOrigin"] = data.MagicContainersOrigin
	}

	return values
}
