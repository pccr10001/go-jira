package jira

import (
	"fmt"
	"strings"

	"github.com/trivago/tgo/tcontainer"
)

// CreateMeta contains information about fields and their attributed to create a ticket.
type CreateMetaInfo struct {
	Expand   string         `json:"expand,omitempty"`
	Projects []*MetaProject `json:"projects,omitempty"`
}

// MetaProject is the meta information about a project returned from createmeta api
type MetaProject struct {
	Expand string `json:"expand,omitempty"`
	Self   string `json:"self, omitempty"`
	Id     string `json:"id,omitempty"`
	Key    string `json:"key,omitempty"`
	Name   string `json:"name,omitempty"`
	// omitted avatarUrls
	IssueTypes []*MetaIssueType `json:"issuetypes,omitempty"`
}

// MetaIssueType represents the different issue types a project has.
//
// Note: Fields is interface because this is an object which can
// have arbitraty keys related to customfields. It is not possible to
// expect these for a general way. This will be returning a map.
// Further processing must be done depending on what is required.
type MetaIssueType struct {
	Self        string                `json:"expand,omitempty"`
	Id          string                `json:"id,omitempty"`
	Description string                `json:"description,omitempty"`
	IconUrl     string                `json:"iconurl,omitempty"`
	Name        string                `json:"name,omitempty"`
	Subtasks    bool                  `json:"subtask,omitempty"`
	Expand      string                `json:"expand,omitempty"`
	Fields      tcontainer.MarshalMap `json:"fields,omitempty"`
}

// GetCreateMeta makes the api call to get the meta information required to create a ticket
func (s *IssueService) GetCreateMeta(projectkey string) (*CreateMetaInfo, *Response, error) {

	apiEndpoint := fmt.Sprintf("/rest/api/2/issue/createmeta?projectKeys=%s&expand=projects.issuetypes.fields", projectkey)

	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	meta := new(CreateMetaInfo)
	resp, err := s.client.Do(req, meta)

	if err != nil {
		return nil, resp, err
	}

	return meta, resp, nil
}

// GetProjectWithName returns a project with "name" from the meta information recieved. If not found, this returns nil.
// The comparision of the name is case insensitive.
func (m *CreateMetaInfo) GetProjectWithName(name string) *MetaProject {
	for _, m := range m.Projects {
		if strings.ToLower(m.Name) == strings.ToLower(name) {
			return m
		}
	}
	return nil
}

// GetProjectWithName returns a project with "name" from the meta information recieved. If not found, this returns nil.
// The comparision of the name is case insensitive.
func (m *CreateMetaInfo) GetProjectWithKey(key string) *MetaProject {
	for _, m := range m.Projects {
		if strings.ToLower(m.Key) == strings.ToLower(key) {
			return m
		}
	}
	return nil
}

// GetIssueWithName returns an IssueType with name from a given MetaProject. If not found, this returns nil.
// The comparision of the name is case insensitive
func (p *MetaProject) GetIssueTypeWithName(name string) *MetaIssueType {
	for _, m := range p.IssueTypes {
		if strings.ToLower(m.Name) == strings.ToLower(name) {
			return m
		}
	}
	return nil
}

// GetMandatoryFields returns a map of all the required fields from the MetaIssueTypes.
// if a frield returned by the api was:
// "customfield_10806": {
//					"required": true,
//					"schema": {
//						"type": "any",
//						"custom": "com.pyxis.greenhopper.jira:gh-epic-link",
//						"customId": 10806
//					},
//					"name": "Epic Link",
//					"hasDefaultValue": false,
//					"operations": [
//						"set"
//					]
//				}
// the returned map would have "Epic Link" as the key and "customfield_10806" as value.
// This choice has been made so that the it is easier to generate the create api request later.
func (t *MetaIssueType) GetMandatoryFields() (map[string]string, error) {
	ret := make(map[string]string)
	for key, _ := range t.Fields {
		required, err := t.Fields.Bool(key + "/required")
		if err != nil {
			return nil, err
		}
		if required {
			name, err := t.Fields.String(key + "/name")
			if err != nil {
				return nil, err
			}
			ret[name] = key
		}
	}
	return ret, nil
}

// GetAllFields returns a map of all the fields for an IssueType. This includes all required and not required.
// The key of the returned map is what you see in the form and the value is how it is representated in the jira schema.
func (t *MetaIssueType) GetAllFields() (map[string]string, error) {
	ret := make(map[string]string)
	for key, _ := range t.Fields {

		name, err := t.Fields.String(key + "/name")
		if err != nil {
			return nil, err
		}
		ret[name] = key
	}
	return ret, nil
}
