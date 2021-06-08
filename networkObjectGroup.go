package goftd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/golang/glog"
)

// NetworkObjectGroup Network Object Group
type NetworkObjectGroup struct {
	ReferenceObject
	Description     string             `json:"description,omitempty"`
	IsSystemDefined bool               `json:"isSystemDefined,omitempty"`
	Objects         []*ReferenceObject `json:"objects,omitempty"`
	Links           *Links             `json:"links,omitempty"`
}

// Reference Returns a reference object
func (g *NetworkObjectGroup) Reference() *ReferenceObject {
	r := ReferenceObject{
		ID:      g.ID,
		Name:    g.Name,
		Version: g.Version,
		Type:    g.Type,
	}

	return &r
}

// GetNetworkObjectGroups Get a list of network objects
func (f *FTD) GetNetworkObjectGroups(limit int) ([]*NetworkObjectGroup, error) {
	var err error

	filter := make(map[string]string)
	filter["limit"] = strconv.Itoa(limit)

	data, err := f.Get("object/networkgroups", filter)
	if err != nil {
		return nil, err
	}

	var v struct {
		Items []*NetworkObjectGroup `json:"items"`
	}

	err = json.Unmarshal(data, &v)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return nil, err
	}

	return v.Items, nil
}

func (f *FTD) GetNetworkObjectGroupBy(filterString string) ([]*NetworkObjectGroup, error) {
	var err error

	filter := make(map[string]string)
	filter["filter"] = filterString

	endpoint := "object/networkgroups"
	data, err := f.Get(endpoint, filter)
	if err != nil {
		return nil, err
	}

	var v struct {
		Items []*NetworkObjectGroup `json:"items"`
	}

	err = json.Unmarshal(data, &v)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return nil, err
	}

	return v.Items, nil
}

// CreateNetworkObjectGroup Create a new network object
func (f *FTD) CreateNetworkObjectGroup(n *NetworkObjectGroup, duplicateAction int) error {
	var err error

	n.Type = "networkobjectgroup"
	_, err = f.Post("object/networkgroups", n)
	if err != nil {
		ftdErr := err.(*FTDError)
		//spew.Dump(ftdErr)
		if len(ftdErr.Message) > 0 && (ftdErr.Message[0].Code == "duplicateName" || ftdErr.Message[0].Code == "newInstanceWithDuplicateId") {
			if f.debug {
				glog.Warningf("This is a duplicate\n")
			}
			if duplicateAction == DuplicateActionError {
				return err
			}
		} else {
			if f.debug {
				glog.Errorf("Error: %s\n", err)
			}
			return err
		}
	}

	query := fmt.Sprintf("name:%s", n.Name)
	obj, err := f.GetNetworkObjectGroupBy(query)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return err
	}

	var o *NetworkObjectGroup
	if len(obj) == 1 {
		o = obj[0]
	} else {
		if f.debug {
			glog.Errorf("Error: length of object is not 1\n")
		}
		return err
	}

	switch duplicateAction {
	case DuplicateActionReplace:
		o.Objects = n.Objects

		err = f.UpdateNetworkObjectGroup(o)
		if err != nil {
			if f.debug {
				glog.Errorf("Error: %s\n", err)
			}
			return err
		}
	}

	*n = *o
	return nil
}

// DeleteNetworkObjectGroup Delete a network object
func (f *FTD) DeleteNetworkObjectGroup(n *NetworkObjectGroup) error {
	var err error

	endpoint := fmt.Sprintf("object/networkgroups/%s", n.ID)
	err = f.Delete(endpoint)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return err
	}

	return nil
}

// UpdateNetworkObjectGroup Updates a network object group
func (f *FTD) UpdateNetworkObjectGroup(n *NetworkObjectGroup) error {
	var err error

	endpoint := fmt.Sprintf("object/networkgroups/%s", n.ID)
	data, err := f.Put(endpoint, n)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return err
	}

	err = json.Unmarshal(data, &n)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return err
	}

	return nil
}

// AddToNetworkObjectGroup Add a Network to an Object Group
func (f *FTD) AddToNetworkObjectGroup(g *NetworkObjectGroup, n *NetworkObject) error {
	var err error
	for k := range g.Objects {
		if g.Objects[k].ID == n.ID {
			if f.debug {
				glog.Errorf("object already in object group\n")
				return fmt.Errorf("object already in object group")
			}
		}
	}

	g.Objects = append(g.Objects, n.Reference())

	err = f.UpdateNetworkObjectGroup(g)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return err
	}

	return nil
}

// DeleteFromNetworkObjectGroup Deletes a Network to an Object Group
func (f *FTD) DeleteFromNetworkObjectGroup(g *NetworkObjectGroup, n *NetworkObject) error {
	var err error
	for k := range g.Objects {
		if g.Objects[k].ID == n.ID {
			g.Objects = append(g.Objects[:k], g.Objects[k+1:]...)
			break
		}
	}

	err = f.UpdateNetworkObjectGroup(g)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return err
	}

	return nil
}

// CreateNetworkObjectGroupFromIPs Create an object group from an array of ip address. Network objects = ip.
func (f *FTD) CreateNetworkObjectGroupFromIPs(name string, ips []string, duplicateAction int) (*NetworkObjectGroup, error) {
	var err error

	ns, err := f.CreateNetworkObjectsFromIPs(ips)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return nil, err
	}

	g := new(NetworkObjectGroup)
	g.Name = name

	for i := range ns {
		g.Objects = append(g.Objects, ns[i].Reference())
	}

	err = f.CreateNetworkObjectGroup(g, duplicateAction)
	if err != nil {
		if f.debug {
			glog.Errorf("Error: %s\n", err)
		}
		return nil, err
	}

	return g, nil
}
