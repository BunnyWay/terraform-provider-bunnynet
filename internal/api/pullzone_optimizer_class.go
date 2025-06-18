// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"encoding/json"
	"errors"
	"golang.org/x/exp/slices"
)

type PullzoneOptimizerClass struct {
	Name       string            `json:"Name"`
	Properties map[string]string `json:"Properties"`
	PullzoneId int64             `json:"-"`
}

func (c *Client) CreatePullzoneOptimizerClass(data PullzoneOptimizerClass) (PullzoneOptimizerClass, error) {
	if data.PullzoneId == 0 {
		return PullzoneOptimizerClass{}, errors.New("pullzone is required")
	}

	pullzone, err := c.GetPullzone(data.PullzoneId)
	if err != nil {
		return PullzoneOptimizerClass{}, err
	}

	pullzone.OptimizerClasses = append(pullzone.OptimizerClasses, data)

	body, err := json.Marshal(map[string]interface{}{
		"OptimizerClasses": pullzone.OptimizerClasses,
	})

	if err != nil {
		return PullzoneOptimizerClass{}, err
	}

	pullzoneResult, err := c.UpdatePullzoneWithBody(pullzone.Id, body)
	if err != nil {
		return PullzoneOptimizerClass{}, err
	}

	class := extractPullzoneOptimizerClass(pullzoneResult, data.Name)
	if class != nil {
		return *class, nil
	}

	return PullzoneOptimizerClass{}, errors.New("Optimizer Image Class not found")
}

func (c *Client) GetPullzoneOptimizerClass(pullzoneId int64, name string) (PullzoneOptimizerClass, error) {
	pullzone, err := c.GetPullzone(pullzoneId)
	if err != nil {
		return PullzoneOptimizerClass{}, err
	}

	class := extractPullzoneOptimizerClass(pullzone, name)
	if class != nil {
		return *class, nil
	}

	return PullzoneOptimizerClass{}, errors.New("Optimizer Image Class not found")
}

func (c *Client) UpdatePullzoneOptimizerClass(data PullzoneOptimizerClass) (PullzoneOptimizerClass, error) {
	pullzone, err := c.GetPullzone(data.PullzoneId)
	if err != nil {
		return PullzoneOptimizerClass{}, err
	}

	for i, class := range pullzone.OptimizerClasses {
		if class.Name == data.Name {
			pullzone.OptimizerClasses[i] = data

			body, err := json.Marshal(map[string]interface{}{
				"OptimizerClasses": pullzone.OptimizerClasses,
			})

			if err != nil {
				return PullzoneOptimizerClass{}, err
			}

			pullzoneResult, err := c.UpdatePullzoneWithBody(pullzone.Id, body)
			if err != nil {
				return PullzoneOptimizerClass{}, err
			}

			class := extractPullzoneOptimizerClass(pullzoneResult, data.Name)
			if class != nil {
				return *class, nil
			}

			break
		}
	}

	return PullzoneOptimizerClass{}, errors.New("Optimizer Image Class not found")
}

func (c *Client) DeletePullzoneOptimizerClass(pullzoneId int64, name string) error {
	pullzone, err := c.GetPullzone(pullzoneId)
	if err != nil {
		return err
	}

	indexToRemove := slices.IndexFunc(pullzone.OptimizerClasses, func(x PullzoneOptimizerClass) bool {
		return x.Name == name
	})

	if indexToRemove == -1 {
		return errors.New("Optimizer Image Class not found")
	}

	pullzone.OptimizerClasses = slices.Delete(pullzone.OptimizerClasses, indexToRemove, indexToRemove+1)

	body, err := json.Marshal(map[string]interface{}{
		"OptimizerClasses": pullzone.OptimizerClasses,
	})

	if err != nil {
		return err
	}

	_, err = c.UpdatePullzoneWithBody(pullzone.Id, body)

	return err
}

func extractPullzoneOptimizerClass(pullzone Pullzone, name string) *PullzoneOptimizerClass {
	for _, class := range pullzone.OptimizerClasses {
		if class.Name == name {
			class.PullzoneId = pullzone.Id
			return &class
		}
	}

	return nil
}
