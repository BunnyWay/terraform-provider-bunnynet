// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/pullzoneshieldresourcevalidator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
)

type PullzoneShieldBotCategory struct {
	Id     uint8
	Action uint8
	Bots   []PullzoneShieldBotCategoryBot
}

type PullzoneShieldBotCategoryBot struct {
	Id     uint64
	Name   string
	Action uint8
}

func (c *Client) getPullzoneShieldBotCategorization(ctx context.Context, shieldZoneId int64) ([]PullzoneShieldBotCategory, error) {
	var result []PullzoneShieldBotCategory

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/shield-zone/%d/bot-categorization", c.apiUrl, shieldZoneId), nil)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		err := extractShieldErrorMessage(resp)
		if err != nil {
			return result, err
		}

		return result, errors.New("get bot categorization failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	tflog.Debug(ctx, fmt.Sprintf("GET /shield/shield-zone/%d/bot-categorization: %s", shieldZoneId, string(bodyResp)), nil)

	var httpResult struct {
		Categories []struct {
			Category       uint8 `json:"category"`
			CategoryAction uint8 `json:"categoryAction"`
			Bots           []struct {
				BotId          uint64 `json:"botId"`
				UserAgentMatch string `json:"userAgentMatch"`
				Action         uint8  `json:"action"`
			} `json:"bots"`
		} `json:"categories"`
	}

	err = json.Unmarshal(bodyResp, &httpResult)
	if err != nil {
		return result, err
	}

	for _, category := range httpResult.Categories {
		bots := make([]PullzoneShieldBotCategoryBot, 0, len(category.Bots))
		for _, bot := range category.Bots {
			botAction := bot.Action

			if botAction == 0 { // None
				switch category.CategoryAction {
				case pullzoneshieldresourcevalidator.BotCategorizationCategoryActionOptIgnore:
					botAction = pullzoneshieldresourcevalidator.BotCategorizationBotActionOptIgnore
				case pullzoneshieldresourcevalidator.BotCategorizationCategoryActionOptBlock:
					botAction = pullzoneshieldresourcevalidator.BotCategorizationBotActionOptBlock
				case pullzoneshieldresourcevalidator.BotCategorizationCategoryActionOptAllow:
					botAction = pullzoneshieldresourcevalidator.BotCategorizationBotActionOptAllow
				default:
					panic("unexpected pullzoneShieldBotCategorizationCategoryAction")
				}
			}

			bots = append(bots, PullzoneShieldBotCategoryBot{
				Id:     bot.BotId,
				Name:   bot.UserAgentMatch,
				Action: botAction,
			})
		}

		result = append(result, PullzoneShieldBotCategory{
			Id:     category.Category,
			Action: category.CategoryAction,
			Bots:   bots,
		})
	}

	return result, nil
}

func (c *Client) pullzoneShieldBotCategorizationUpdateCategoryAction(ctx context.Context, shieldZoneId int64, categoryId uint8, action uint8) error {
	body, err := json.Marshal(map[string]uint8{
		"action": action,
	})

	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("PUT /shield/shield-zone/%d/bot-categorization/categories/%d: %s", shieldZoneId, categoryId, string(body)))

	resp, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/shield/shield-zone/%d/bot-categorization/categories/%d", c.apiUrl, shieldZoneId, categoryId), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("PUT /shield/shield-zone/%d/bot-categorization/categories/%d: %s", shieldZoneId, categoryId, resp.Status))

	if resp.StatusCode != http.StatusOK {
		err := extractShieldErrorMessage(resp)
		if err == nil {
			return errors.New("update botcategorization failed with " + resp.Status)
		}

		return err
	}

	return nil
}

func (c *Client) pullzoneShieldBotCategorizationUpdateBotAction(ctx context.Context, shieldZoneId int64, botId uint64, action uint8) error {
	body, err := json.Marshal(map[string]uint8{
		"action": action,
	})

	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("PUT /shield/shield-zone/%d/bot-categorization/bots/%d: %s", shieldZoneId, botId, string(body)))

	resp, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/shield/shield-zone/%d/bot-categorization/bots/%d", c.apiUrl, shieldZoneId, botId), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("PUT /shield/shield-zone/%d/bot-categorization/bots/%d: %s", shieldZoneId, botId, resp.Status))

	if resp.StatusCode != http.StatusOK {
		err := extractShieldErrorMessage(resp)
		if err == nil {
			return errors.New("update botcategorization override failed with " + resp.Status)
		}

		return err
	}

	return nil
}
