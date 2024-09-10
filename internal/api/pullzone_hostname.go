// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"golang.org/x/exp/slices"
	"net/http"
	"strings"
)

type PullzoneHostname struct {
	Id               int64  `json:"Id,omitempty"`
	PullzoneId       int64  `json:"-"`
	Name             string `json:"Value"`
	IsSystemHostname bool   `json:"IsSystemHostname"`
	HasCertificate   bool   `json:"HasCertificate"`
	ForceSSL         bool   `json:"ForceSSL"`
	Certificate      string `json:"Certificate"`
	CertificateKey   string `json:"CertificateKey"`
}

func (c *Client) CreatePullzoneHostname(data PullzoneHostname) (PullzoneHostname, error) {
	pullzoneId := data.PullzoneId
	if pullzoneId == 0 {
		return PullzoneHostname{}, errors.New("pullzone is required")
	}

	pullzone, err := c.GetPullzone(data.PullzoneId)
	if err != nil {
		return PullzoneHostname{}, err
	}

	// if creating the default hostname, return existing
	if strings.HasSuffix(data.Name, "."+pullzone.CnameDomain) {
		hostnameIdx := slices.IndexFunc(pullzone.Hostnames, func(hostname PullzoneHostname) bool {
			return hostname.IsSystemHostname && hostname.Name == data.Name
		})

		if hostnameIdx > -1 {
			data.Id = pullzone.Hostnames[hostnameIdx].Id
			data.PullzoneId = pullzone.Id
			return c.UpdatePullzoneHostname(data, pullzone.Hostnames[hostnameIdx])
		}

		// if system hostname not found, try to create it, the API should return an error
	}

	body, err := json.Marshal(map[string]string{
		"Hostname": data.Name,
	})

	if err != nil {
		return PullzoneHostname{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d/addHostname", c.apiUrl, pullzoneId), bytes.NewReader(body))
	if err != nil {
		return PullzoneHostname{}, err
	}

	if resp.StatusCode != http.StatusNoContent {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneHostname{}, errors.New("addHostname failed: " + err.Error())
		} else {
			return PullzoneHostname{}, errors.New("addHostname failed with " + resp.Status)
		}
	}

	pullzone, err = c.GetPullzone(pullzoneId)
	if err != nil {
		return PullzoneHostname{}, err
	}

	for _, hostname := range pullzone.Hostnames {
		if hostname.Name == data.Name {
			previousData := PullzoneHostname{
				HasCertificate: hostname.HasCertificate,
				ForceSSL:       hostname.ForceSSL,
			}
			hostname.PullzoneId = pullzoneId
			hostname.HasCertificate = data.HasCertificate
			hostname.ForceSSL = data.ForceSSL
			hostname.Certificate = data.Certificate
			hostname.CertificateKey = data.CertificateKey

			return c.UpdatePullzoneHostname(hostname, previousData)
		}
	}

	return PullzoneHostname{}, errors.New("Hostname not found")
}

func (c *Client) UpdatePullzoneHostname(data PullzoneHostname, previousData PullzoneHostname) (PullzoneHostname, error) {
	pullzoneId := data.PullzoneId
	if pullzoneId == 0 {
		return PullzoneHostname{}, errors.New("pullzone is required")
	}

	if data.IsSystemHostname && previousData.HasCertificate && !data.HasCertificate {
		return PullzoneHostname{}, errors.New("removing a certificate from an internal hostname is not supported")
	}

	shouldRemoveCustomCertificate := len(previousData.Certificate) > 0 && len(data.Certificate) == 0
	didNotHaveAManagedCertificate := (!previousData.HasCertificate && data.HasCertificate) || (shouldRemoveCustomCertificate)
	hasCustomCertificate := len(data.Certificate) > 0 && len(data.CertificateKey) > 0

	shouldRemoveCertificate := !data.IsSystemHostname && (shouldRemoveCustomCertificate || previousData.HasCertificate && !data.HasCertificate)
	shouldAddCustomCertificate := !data.IsSystemHostname && data.HasCertificate && hasCustomCertificate && data.Certificate != previousData.Certificate
	shouldAddManagedCertificate := !data.IsSystemHostname && data.HasCertificate && !hasCustomCertificate && didNotHaveAManagedCertificate
	shouldSetForceSsl := previousData.ForceSSL != data.ForceSSL

	if shouldRemoveCertificate {
		body, err := json.Marshal(map[string]interface{}{
			"Hostname": data.Name,
		})

		if err != nil {
			return PullzoneHostname{}, err
		}

		resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/pullzone/%d/removeCertificate", c.apiUrl, pullzoneId), bytes.NewReader(body))
		if err != nil {
			return PullzoneHostname{}, err
		}

		if resp.StatusCode != http.StatusNoContent {
			err := utils.ExtractErrorMessage(resp)
			if err != nil {
				return PullzoneHostname{}, errors.New("removeCertificate failed: " + err.Error())
			} else {
				return PullzoneHostname{}, errors.New("removeCertificate failed with " + resp.Status)
			}
		}
	}

	if shouldAddCustomCertificate {
		certificateB64 := base64.StdEncoding.EncodeToString([]byte(data.Certificate))
		certificateKeyB64 := base64.StdEncoding.EncodeToString([]byte(data.CertificateKey))

		body, err := json.Marshal(map[string]interface{}{
			"Hostname":       data.Name,
			"Certificate":    certificateB64,
			"CertificateKey": certificateKeyB64,
		})

		if err != nil {
			return PullzoneHostname{}, err
		}

		resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d/addCertificate", c.apiUrl, pullzoneId), bytes.NewReader(body))
		if err != nil {
			return PullzoneHostname{}, err
		}

		if resp.StatusCode != http.StatusNoContent {
			err := utils.ExtractErrorMessage(resp)
			if err != nil {
				return PullzoneHostname{}, errors.New("addCertificate failed: " + err.Error())
			} else {
				return PullzoneHostname{}, errors.New("addCertificate failed with " + resp.Status)
			}
		}
	}

	if shouldAddManagedCertificate {
		resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/pullzone/loadFreeCertificate?hostname=%s", c.apiUrl, data.Name), nil)
		if err != nil {
			return PullzoneHostname{}, err
		}

		if resp.StatusCode != http.StatusOK {
			err := utils.ExtractErrorMessage(resp)
			if err != nil {
				return PullzoneHostname{}, errors.New("loadFreeCertificate failed: " + err.Error())
			} else {
				return PullzoneHostname{}, errors.New("loadFreeCertificate failed with " + resp.Status)
			}
		}
	}

	if shouldSetForceSsl {
		body, err := json.Marshal(map[string]interface{}{
			"ForceSSL": data.ForceSSL,
			"Hostname": data.Name,
		})

		if err != nil {
			return PullzoneHostname{}, err
		}

		resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d/setForceSSL", c.apiUrl, pullzoneId), bytes.NewReader(body))
		if err != nil {
			return PullzoneHostname{}, err
		}

		if resp.StatusCode != http.StatusNoContent {
			err := utils.ExtractErrorMessage(resp)
			if err != nil {
				return PullzoneHostname{}, errors.New("forceSSL failed: " + err.Error())
			} else {
				return PullzoneHostname{}, errors.New("forceSSL failed with " + resp.Status)
			}
		}
	}

	return c.GetPullzoneHostname(pullzoneId, data.Id)
}

func (c *Client) GetPullzoneHostname(pullzoneId int64, id int64) (PullzoneHostname, error) {
	pullzone, err := c.GetPullzone(pullzoneId)
	if err != nil {
		return PullzoneHostname{}, err
	}

	for _, hostname := range pullzone.Hostnames {
		if hostname.Id == id {
			hostname.PullzoneId = pullzoneId
			return hostname, nil
		}
	}

	return PullzoneHostname{}, errors.New("Hostname not found")
}

func (c *Client) GetPullzoneHostnameByName(pullzoneId int64, hostname string) (PullzoneHostname, error) {
	pullzone, err := c.GetPullzone(pullzoneId)
	if err != nil {
		return PullzoneHostname{}, err
	}

	for _, v := range pullzone.Hostnames {
		if v.Name == hostname {
			v.PullzoneId = pullzoneId
			return v, nil
		}
	}

	return PullzoneHostname{}, errors.New("Hostname not found")
}

func (c *Client) DeletePullzoneHostname(pullzoneId int64, hostname string) error {
	body, err := json.Marshal(map[string]interface{}{
		"Hostname": hostname,
	})

	if err != nil {
		return err
	}

	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/pullzone/%d/removeHostname", c.apiUrl, pullzoneId), bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return errors.New("delete failed: " + err.Error())
		} else {
			return errors.New("delete failed with " + resp.Status)
		}
	}

	return nil
}
