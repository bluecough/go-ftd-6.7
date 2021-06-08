package goftd

import (
	"fmt"
)

type DeployObject struct {
	ReferenceObject
	Description     string `json:"description,omitempty"`
	StatusMessage   string `json:"subType"`
	CliErrorMessage string `json:"value"`
	State           string `json:"isSystemDefined,omitempty"`
	Links           *Links `json:"links,omitempty"`
}

// Reference Returns a reference object
func (n *DeployObject) Reference() *ReferenceObject {
	r := ReferenceObject{
		ID:      n.ID,
		Name:    n.Name,
	}

	return &r
}


func (f *FTD) PostDeploy(n *DeployObject) error {
	var err error
	_, err = f.Post(apiDeploy, nil)
	if err != nil {
		fmt.Errorf("error: %s\n", err)
	}
	return nil
}
