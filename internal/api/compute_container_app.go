// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"strings"
)

type ComputeContainerAppContainerEndpointPortMapping struct {
	ContainerPort int64    `json:"containerPort"`
	ExposedPort   int64    `json:"exposedPort,omitempty"`
	Protocols     []string `json:"protocols,omitempty"`
}

type ComputeContainerAppContainerEndpointStickySessions struct {
	Enabled        bool     `json:"enabled"`
	SessionHeaders []string `json:"sessionHeaders"`
	CookieName     string   `json:"cookieName"`
}

type ComputeContainerAppContainerEndpoint struct {
	DisplayName    string                                              `json:"displayName"`
	PublicHost     string                                              `json:"publicHost,omitempty"`
	Type           string                                              `json:"type"`
	IsSslEnabled   bool                                                `json:"isSslEnabled"`
	PullZoneId     string                                              `json:"pullZoneId,omitempty"`
	PortMappings   []ComputeContainerAppContainerEndpointPortMapping   `json:"portMappings"`
	StickySessions *ComputeContainerAppContainerEndpointStickySessions `json:"stickySessions"`
}

type ComputeContainerAppContainerEnv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ComputeContainerAppContainerVolumeMount struct {
	Name string `json:"name"`
	Path string `json:"mountPath"`
}

type ComputeContainerAppAutoscaling struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

type ComputeContainerAppRegions struct {
	AllowedRegionIds  []string `json:"allowedRegionIds"`
	RequiredRegionIds []string `json:"requiredRegionIds"`
	MaxAllowedRegions *int64   `json:"maxAllowedRegions"`
}

type ComputeContainerAppContainerEntrypoint struct {
	Command          string `json:"command"`
	Arguments        string `json:"arguments"`
	WorkingDirectory string `json:"workingDirectory"`
}

type ComputeContainerAppContainerProbeHttpRequest struct {
	Path       string `json:"path"`
	PortNumber int64  `json:"portNumber"`
}

type ComputeContainerAppContainerProbeHttpResponse struct {
	ExpectedStatusCode *string `json:"expectedStatusCode"`
}

type ComputeContainerAppContainerProbeHttp struct {
	Request  ComputeContainerAppContainerProbeHttpRequest  `json:"request"`
	Response ComputeContainerAppContainerProbeHttpResponse `json:"response"`
}

type ComputeContainerAppContainerProbeTcpRequest struct {
	PortNumber int64 `json:"portNumber"`
}

type ComputeContainerAppContainerProbeTcp struct {
	Request ComputeContainerAppContainerProbeTcpRequest `json:"request"`
}

type ComputeContainerAppContainerProbeGrpcRequest struct {
	PortNumber  int64  `json:"portNumber"`
	ServiceName string `json:"serviceName"`
}

type ComputeContainerAppContainerProbeGrpc struct {
	Request ComputeContainerAppContainerProbeGrpcRequest `json:"request"`
}

type ComputeContainerAppContainerProbes struct {
	Startup   *ComputeContainerAppContainerProbe `json:"startup"`
	Readiness *ComputeContainerAppContainerProbe `json:"readiness"`
	Liveness  *ComputeContainerAppContainerProbe `json:"liveness"`
}

type ComputeContainerAppContainerProbe struct {
	InitialDelaySeconds int64                                  `json:"initialDelaySeconds"`
	PeriodSeconds       int64                                  `json:"periodSeconds"`
	TimeoutSeconds      int64                                  `json:"timeoutSeconds"`
	FailureThreshold    int64                                  `json:"failureThreshold"`
	SuccessThreshold    int64                                  `json:"successThreshold"`
	HttpGet             *ComputeContainerAppContainerProbeHttp `json:"httpGet"`
	TcpSocket           *ComputeContainerAppContainerProbeTcp  `json:"tcpSocket"`
	Grpc                *ComputeContainerAppContainerProbeGrpc `json:"grpc"`
}

type ComputeContainerAppContainer struct {
	Id                   string                                    `json:"id,omitempty"`
	Name                 string                                    `json:"name"`
	PackageId            string                                    `json:"packageId"`
	ImageNamespace       string                                    `json:"imageNamespace"`
	ImageName            string                                    `json:"imageName"`
	ImageTag             string                                    `json:"imageTag"`
	ImageRegistryId      string                                    `json:"imageRegistryId"`
	ImagePullPolicy      string                                    `json:"imagePullPolicy"`
	EntryPoint           ComputeContainerAppContainerEntrypoint    `json:"entryPoint"`
	Probes               ComputeContainerAppContainerProbes        `json:"probes"`
	EnvironmentVariables []ComputeContainerAppContainerEnv         `json:"environmentVariables"`
	Endpoints            []ComputeContainerAppContainerEndpoint    `json:"endpoints"`
	VolumeMounts         []ComputeContainerAppContainerVolumeMount `json:"volumeMounts"`
}

type ComputeContainerAppVolume struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type computeContainerAppSaveApplicationContainerEndpointPortMappingRequest struct {
	ContainerPort int64    `json:"containerPort"`
	ExposedPort   int64    `json:"exposedPort,omitempty"`
	Protocols     []string `json:"protocols,omitempty"`
}

type computeContainerAppSaveApplicationContainerEndpointCdnRequest struct {
	IsSslEnabled   bool                                                                    `json:"isSslEnabled"`
	StickySessions *ComputeContainerAppContainerEndpointStickySessions                     `json:"stickySessions"`
	PortMappings   []computeContainerAppSaveApplicationContainerEndpointPortMappingRequest `json:"portMappings"`
}

type computeContainerAppSaveApplicationContainerEndpointInternalIpRequest struct {
	PortMappings []computeContainerAppSaveApplicationContainerEndpointPortMappingRequest `json:"portMappings"`
}

type computeContainerAppSaveApplicationContainerEndpointAnycastRequest struct {
	Type         string                                                                  `json:"type"`
	PortMappings []computeContainerAppSaveApplicationContainerEndpointPortMappingRequest `json:"portMappings"`
}

type computeContainerAppSaveApplicationContainerEndpointRequest struct {
	DisplayName string                                                                `json:"displayName"`
	Cdn         *computeContainerAppSaveApplicationContainerEndpointCdnRequest        `json:"cdn,omitempty"`
	InternalIp  *computeContainerAppSaveApplicationContainerEndpointInternalIpRequest `json:"internalIp,omitempty"`
	Anycast     *computeContainerAppSaveApplicationContainerEndpointAnycastRequest    `json:"anycast,omitempty"`
}

type computeContainerAppSaveApplicationContainerRequest struct {
	Id                   string                                                       `json:"id,omitempty"`
	PackageId            string                                                       `json:"packageId"`
	Name                 string                                                       `json:"name"`
	ImageNamespace       string                                                       `json:"imageNamespace"`
	ImageName            string                                                       `json:"imageName"`
	ImageTag             string                                                       `json:"imageTag"`
	ImageRegistryId      string                                                       `json:"imageRegistryId"`
	ImagePullPolicy      string                                                       `json:"imagePullPolicy"`
	EntryPoint           ComputeContainerAppContainerEntrypoint                       `json:"entryPoint"`
	Probes               ComputeContainerAppContainerProbes                           `json:"probes"`
	Endpoints            []computeContainerAppSaveApplicationContainerEndpointRequest `json:"endpoints"`
	EnvironmentVariables []ComputeContainerAppContainerEnv                            `json:"environmentVariables"`
	VolumeMounts         []ComputeContainerAppContainerVolumeMount                    `json:"volumeMounts,omitempty"`
}

type computeContainerAppSaveApplicationRequest struct {
	Id                 string                                               `json:"id,omitempty"`
	Name               string                                               `json:"name"`
	RuntimeType        string                                               `json:"runtimeType"`
	AutoScaling        ComputeContainerAppAutoscaling                       `json:"autoScaling"`
	RegionSettings     ComputeContainerAppRegions                           `json:"regionSettings"`
	ContainerTemplates []computeContainerAppSaveApplicationContainerRequest `json:"containerTemplates"`
	Volumes            []ComputeContainerAppVolume                          `json:"volumes,omitempty"`
}

type ComputeContainerApp struct {
	Id                 string                         `json:"id"`
	Name               string                         `json:"name"`
	RuntimeType        string                         `json:"runtimeType"`
	RegionSettings     ComputeContainerAppRegions     `json:"regionSettings"`
	ContainerTemplates []ComputeContainerAppContainer `json:"containerTemplates"`
	Volumes            []ComputeContainerAppVolume    `json:"volumes"`
	AutoScaling        ComputeContainerAppAutoscaling `json:"autoScaling"`
}

func (c *Client) GetComputeContainerApp(ctx context.Context, id string) (ComputeContainerApp, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/mc/apps/%s", c.apiUrl, id), nil)
	if err != nil {
		return ComputeContainerApp{}, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ComputeContainerApp{}, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return ComputeContainerApp{}, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return ComputeContainerApp{}, err
	}
	var result ComputeContainerApp

	tflog.Debug(ctx, fmt.Sprintf("GET /mc/apps/%s: %s", id, string(bodyResp)))

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return ComputeContainerApp{}, err
	}

	computeContainerAppSort(&result)

	return result, nil
}

func (c *Client) CreateComputeContainerApp(ctx context.Context, data ComputeContainerApp) (ComputeContainerApp, error) {
	return c.UpdateComputeContainerApp(ctx, data)
}

func (c *Client) UpdateComputeContainerApp(ctx context.Context, data ComputeContainerApp) (ComputeContainerApp, error) {
	dataRequest := computeContainerAppSaveApplicationRequest{}
	dataRequest.Id = data.Id
	dataRequest.Name = data.Name
	dataRequest.RuntimeType = "Shared"
	dataRequest.AutoScaling = data.AutoScaling
	dataRequest.RegionSettings = data.RegionSettings

	for _, b := range data.ContainerTemplates {
		endpoints := make([]computeContainerAppSaveApplicationContainerEndpointRequest, len(b.Endpoints))

		for i, e := range b.Endpoints {
			endpoint, err := convertEndpointToSaveRequest(e)
			if err != nil {
				return ComputeContainerApp{}, err
			}
			endpoints[i] = endpoint
		}

		dataRequest.ContainerTemplates = append(dataRequest.ContainerTemplates, computeContainerAppSaveApplicationContainerRequest{
			Id:                   b.Id,
			Name:                 b.Name,
			PackageId:            b.PackageId,
			ImageRegistryId:      b.ImageRegistryId,
			ImageNamespace:       b.ImageNamespace,
			ImageName:            b.ImageName,
			ImageTag:             b.ImageTag,
			ImagePullPolicy:      b.ImagePullPolicy,
			EntryPoint:           b.EntryPoint,
			Probes:               b.Probes,
			Endpoints:            endpoints,
			EnvironmentVariables: b.EnvironmentVariables,
			VolumeMounts:         b.VolumeMounts,
		})
	}

	dataRequest.Volumes = append(dataRequest.Volumes, data.Volumes...)

	body, err := json.Marshal(dataRequest)
	if err != nil {
		return ComputeContainerApp{}, err
	}

	method := http.MethodPost
	url := fmt.Sprintf("%s/mc/apps", c.apiUrl)

	if data.Id != "" {
		method = http.MethodPut
		url = fmt.Sprintf("%s/mc/apps/%s", c.apiUrl, data.Id)
	}

	tflog.Debug(ctx, fmt.Sprintf("%s %s: %s", method, url, string(body)))

	resp, err := c.doRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return ComputeContainerApp{}, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		err := utils.ExtractMCErrorMessage(resp)
		if err != nil {
			return ComputeContainerApp{}, err
		} else {
			return ComputeContainerApp{}, errors.New(resp.Status)
		}
	}

	bodyStr, _ := io.ReadAll(resp.Body)

	defer func() {
		_ = resp.Body.Close()
	}()

	tflog.Debug(ctx, fmt.Sprintf("%s %s response: %s", method, url, string(bodyStr)))

	var result struct {
		Id string `json:"id"`
	}

	err = json.Unmarshal(bodyStr, &result)
	if err != nil {
		return ComputeContainerApp{}, err
	}

	return c.GetComputeContainerApp(ctx, result.Id)
}

func convertEndpointToSaveRequest(e ComputeContainerAppContainerEndpoint) (computeContainerAppSaveApplicationContainerEndpointRequest, error) {
	endpoint := computeContainerAppSaveApplicationContainerEndpointRequest{
		DisplayName: e.DisplayName,
	}

	if e.Type == "CDN" {
		portMappings := make([]computeContainerAppSaveApplicationContainerEndpointPortMappingRequest, len(e.PortMappings))
		for j, pm := range e.PortMappings {
			portMappings[j] = computeContainerAppSaveApplicationContainerEndpointPortMappingRequest{
				ContainerPort: pm.ContainerPort,
			}
		}

		stickySessions := e.StickySessions
		if stickySessions != nil && !stickySessions.Enabled {
			stickySessions = nil
		}

		endpoint.Cdn = &computeContainerAppSaveApplicationContainerEndpointCdnRequest{
			IsSslEnabled:   e.IsSslEnabled,
			PortMappings:   portMappings,
			StickySessions: stickySessions,
		}

		return endpoint, nil
	}

	if e.Type == "Anycast" {
		portMappings := make([]computeContainerAppSaveApplicationContainerEndpointPortMappingRequest, len(e.PortMappings))
		for j, pm := range e.PortMappings {
			portMappings[j] = computeContainerAppSaveApplicationContainerEndpointPortMappingRequest(pm)
		}

		endpoint.Anycast = &computeContainerAppSaveApplicationContainerEndpointAnycastRequest{
			Type:         "IPv4",
			PortMappings: portMappings,
		}

		return endpoint, nil
	}

	// @TODO replace with InternalIp
	if e.Type == "PublicIp" {
		portMappings := make([]computeContainerAppSaveApplicationContainerEndpointPortMappingRequest, len(e.PortMappings))
		for j, pm := range e.PortMappings {
			portMappings[j] = computeContainerAppSaveApplicationContainerEndpointPortMappingRequest{
				ContainerPort: pm.ContainerPort,
				Protocols:     pm.Protocols,
			}
		}

		endpoint.InternalIp = &computeContainerAppSaveApplicationContainerEndpointInternalIpRequest{
			PortMappings: portMappings,
		}

		return endpoint, nil
	}

	return endpoint, errors.New("Invalid endpoint type: " + e.Type)
}

func (c *Client) DeleteComputeContainerApp(id string) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/mc/apps/%s", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func computeContainerAppSort(app *ComputeContainerApp) {
	slices.SortFunc(app.ContainerTemplates, func(a, b ComputeContainerAppContainer) int {
		if a.Name == b.Name {
			return 0
		}

		if a.Name < b.Name {
			return -1
		}

		return 1
	})

	for i, container := range app.ContainerTemplates {
		switch strings.ToLower(container.ImagePullPolicy) {
		case "always":
			app.ContainerTemplates[i].ImagePullPolicy = "Always"
		case "ifnotpresent":
			app.ContainerTemplates[i].ImagePullPolicy = "IfNotPresent"
		}

		for j, endpoint := range container.Endpoints {
			switch strings.ToLower(endpoint.Type) {
			case "anycast":
				app.ContainerTemplates[i].Endpoints[j].Type = "Anycast"
			case "cdn":
				app.ContainerTemplates[i].Endpoints[j].Type = "CDN"
			case "publicip":
				app.ContainerTemplates[i].Endpoints[j].Type = "PublicIp"
			}
		}

		slices.SortFunc(container.Endpoints, func(a, b ComputeContainerAppContainerEndpoint) int {
			if a.DisplayName == b.DisplayName {
				return 0
			}

			if a.DisplayName < b.DisplayName {
				return -1
			}

			return 1
		})

		slices.SortFunc(container.EnvironmentVariables, func(a, b ComputeContainerAppContainerEnv) int {
			if a.Name == b.Name {
				return 0
			}

			if a.Name < b.Name {
				return -1
			}

			return 1
		})
	}

	slices.SortFunc(app.Volumes, func(a, b ComputeContainerAppVolume) int {
		if a.Name == b.Name {
			return 0
		}

		if a.Name < b.Name {
			return -1
		}

		return 1
	})
}
