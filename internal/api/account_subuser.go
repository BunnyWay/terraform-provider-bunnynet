package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"slices"
)

type AccountSubuser struct {
	Id        string   `json:"Id,omitempty"`
	Firstname string   `json:"FirstName"`
	Lastname  string   `json:"LastName"`
	Email     string   `json:"Email"`
	Password  string   `json:"Password,omitempty"`
	Roles     []string `json:"Roles"`
}

func (c *Client) GetAccountSubuser(ctx context.Context, id string) (AccountSubuser, error) {
	var data AccountSubuser

	resp, err := c.doJWTRequest(http.MethodGet, fmt.Sprintf("%s/team/member/%s", c.apiUrl, id), nil)
	if err != nil {
		return data, err
	}

	tflog.Info(ctx, fmt.Sprintf("GET /team/member/%s: %s", id, resp.Status))

	if resp.StatusCode == http.StatusNotFound {
		return data, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return data, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	tflog.Info(ctx, fmt.Sprintf("GET /team/member/%s: %s", id, string(bodyResp)))

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return data, err
	}

	accountSubuserCleanupPermissions(&data)

	return data, nil
}

func (c *Client) GetAccountSubuserByEmail(ctx context.Context, email string) (AccountSubuser, error) {
	var result struct {
		Items        []AccountSubuser
		CurrentPage  uint64
		TotalItems   uint64
		HasMoreItems bool
	}

	resp, err := c.doJWTRequest(http.MethodGet, fmt.Sprintf("%s/team/member", c.apiUrl), nil)
	if err != nil {
		return AccountSubuser{}, err
	}

	tflog.Info(ctx, fmt.Sprintf("GET /team/member: %s", resp.Status))

	if resp.StatusCode == http.StatusNotFound {
		return AccountSubuser{}, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return AccountSubuser{}, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountSubuser{}, err
	}

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return AccountSubuser{}, err
	}

	for _, item := range result.Items {
		if item.Email == email {
			accountSubuserCleanupPermissions(&item)
			return item, nil
		}
	}

	return AccountSubuser{}, nil
}

func (c *Client) CreateAccountSubuser(ctx context.Context, data AccountSubuser) (AccountSubuser, error) {
	body, err := json.Marshal(data)

	if err != nil {
		return AccountSubuser{}, err
	}

	resp, err := c.doJWTRequest(http.MethodPost, fmt.Sprintf("%s/team/member", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return AccountSubuser{}, err
	}

	tflog.Info(ctx, fmt.Sprintf("POST /team/member: %s", string(body)))

	if resp.StatusCode != http.StatusCreated {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return AccountSubuser{}, err
		}

		return AccountSubuser{}, errors.New("create account sub-user failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountSubuser{}, err
	}
	_ = resp.Body.Close()

	var dataApiResult AccountSubuser
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return AccountSubuser{}, err
	}

	return c.GetAccountSubuser(ctx, dataApiResult.Id)
}

func (c *Client) UpdateAccountSubuser(ctx context.Context, data AccountSubuser) (AccountSubuser, error) {
	accountSubuserCleanupPermissions(&data)
	id := data.Id
	data.Id = ""
	body, err := json.Marshal(data)
	if err != nil {
		return AccountSubuser{}, err
	}

	tflog.Info(ctx, fmt.Sprintf("POST /team/member/%s: %s", id, string(body)))

	resp, err := c.doJWTRequest(http.MethodPost, fmt.Sprintf("%s/team/member/%s", c.apiUrl, id), bytes.NewReader(body))
	if err != nil {
		return AccountSubuser{}, err
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return AccountSubuser{}, err
		}

		return AccountSubuser{}, errors.New("update account sub-user failed with " + resp.Status)
	}

	return c.GetAccountSubuser(ctx, id)
}

func (c *Client) DeleteAccountSubuser(ctx context.Context, id string) error {
	resp, err := c.doJWTRequest(http.MethodDelete, fmt.Sprintf("%s/team/member/%s", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

func accountSubuserCleanupPermissions(data *AccountSubuser) {
	for i, p := range data.Roles {
		if p == "Subuser" {
			data.Roles = slices.Delete(data.Roles, i, i+1)
		}
	}
}
