package remote

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/rand"

	tfe "github.com/hashicorp/go-tfe"
)

type MockConfigurationVersions struct {
	configVersions map[string]*tfe.ConfigurationVersion
	uploadURLs     map[string]*tfe.ConfigurationVersion
	workspaces     map[string]*tfe.ConfigurationVersion
}

func NewMockConfigurationVersions() *MockConfigurationVersions {
	return &MockConfigurationVersions{
		configVersions: make(map[string]*tfe.ConfigurationVersion),
		uploadURLs:     make(map[string]*tfe.ConfigurationVersion),
		workspaces:     make(map[string]*tfe.ConfigurationVersion),
	}
}

func (m *MockConfigurationVersions) List(ctx context.Context, workspaceID string, options tfe.ConfigurationVersionListOptions) ([]*tfe.ConfigurationVersion, error) {
	var cvs []*tfe.ConfigurationVersion
	for _, cv := range m.configVersions {
		cvs = append(cvs, cv)
	}
	return cvs, nil
}

func (m *MockConfigurationVersions) Create(ctx context.Context, workspaceID string, options tfe.ConfigurationVersionCreateOptions) (*tfe.ConfigurationVersion, error) {
	id := generateID("cv-")
	url := fmt.Sprintf("https://app.terraform.io/_archivist/%s", id)

	cv := &tfe.ConfigurationVersion{
		ID:        id,
		Status:    tfe.ConfigurationPending,
		UploadURL: url,
	}

	m.configVersions[cv.ID] = cv
	m.uploadURLs[url] = cv
	m.workspaces[workspaceID] = cv

	return cv, nil
}

func (m *MockConfigurationVersions) Read(ctx context.Context, cvID string) (*tfe.ConfigurationVersion, error) {
	cv, ok := m.configVersions[cvID]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	return cv, nil
}

func (m *MockConfigurationVersions) Upload(ctx context.Context, url, path string) error {
	cv, ok := m.uploadURLs[url]
	if !ok {
		return errors.New("404 not found")
	}
	cv.Status = tfe.ConfigurationUploaded
	return nil
}

type MockOrganizations struct {
	organizations map[string]*tfe.Organization
}

func NewMockOrganizations() *MockOrganizations {
	return &MockOrganizations{
		organizations: make(map[string]*tfe.Organization),
	}
}

func (m *MockOrganizations) List(ctx context.Context, options tfe.OrganizationListOptions) ([]*tfe.Organization, error) {
	var orgs []*tfe.Organization
	for _, org := range m.organizations {
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (m *MockOrganizations) Create(ctx context.Context, options tfe.OrganizationCreateOptions) (*tfe.Organization, error) {
	org := &tfe.Organization{Name: *options.Name}
	m.organizations[org.Name] = org
	return org, nil
}

func (m *MockOrganizations) Read(ctx context.Context, name string) (*tfe.Organization, error) {
	org, ok := m.organizations[name]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	return org, nil
}

func (m *MockOrganizations) Update(ctx context.Context, name string, options tfe.OrganizationUpdateOptions) (*tfe.Organization, error) {
	org, ok := m.organizations[name]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	org.Name = *options.Name
	return org, nil

}

func (m *MockOrganizations) Delete(ctx context.Context, name string) error {
	delete(m.organizations, name)
	return nil
}

type MockPlans struct {
	plans map[string]*tfe.Plan
}

func NewMockPlans() *MockPlans {
	return &MockPlans{
		plans: make(map[string]*tfe.Plan),
	}
}

func (m *MockPlans) Read(ctx context.Context, planID string) (*tfe.Plan, error) {
	p, ok := m.plans[planID]
	if !ok {
		url := fmt.Sprintf("https://app.terraform.io/_archivist/%s", planID)

		p = &tfe.Plan{
			ID:         planID,
			LogReadURL: url,
			Status:     tfe.PlanFinished,
		}

		m.plans[p.ID] = p
	}

	return p, nil
}

func (m *MockPlans) Logs(context.Context, string) (io.Reader, error) {
	logs := bytes.NewBufferString(`Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + null_resource.hello
      id: <computed>


Plan: 1 to add, 0 to change, 0 to destroy.`)

	return logs, nil
}

type MockRuns struct {
	runs       map[string]*tfe.Run
	workspaces map[string][]*tfe.Run
}

func NewMockRuns() *MockRuns {
	return &MockRuns{
		runs:       make(map[string]*tfe.Run),
		workspaces: make(map[string][]*tfe.Run),
	}
}

func (m *MockRuns) List(ctx context.Context, workspaceID string, options tfe.RunListOptions) ([]*tfe.Run, error) {
	var rs []*tfe.Run
	for _, r := range m.workspaces[workspaceID] {
		rs = append(rs, r)
	}
	return rs, nil
}

func (m *MockRuns) Create(ctx context.Context, options tfe.RunCreateOptions) (*tfe.Run, error) {
	id := generateID("run-")
	p := &tfe.Plan{
		ID:     generateID("plan-"),
		Status: tfe.PlanPending,
	}

	r := &tfe.Run{
		ID:     id,
		Plan:   p,
		Status: tfe.RunPending,
	}

	m.runs[r.ID] = r
	m.workspaces[options.Workspace.ID] = append(m.workspaces[options.Workspace.ID], r)

	return r, nil
}

func (m *MockRuns) Read(ctx context.Context, runID string) (*tfe.Run, error) {
	r, ok := m.runs[runID]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	return r, nil
}

func (m *MockRuns) Apply(ctx context.Context, runID string, options tfe.RunApplyOptions) error {
	panic("not implemented")
}

func (m *MockRuns) Cancel(ctx context.Context, runID string, options tfe.RunCancelOptions) error {
	panic("not implemented")
}

func (m *MockRuns) Discard(ctx context.Context, runID string, options tfe.RunDiscardOptions) error {
	panic("not implemented")
}

type MockStateVersions struct {
	states        map[string][]byte
	stateVersions map[string]*tfe.StateVersion
	workspaces    map[string][]string
}

func NewMockStateVersions() *MockStateVersions {
	return &MockStateVersions{
		states:        make(map[string][]byte),
		stateVersions: make(map[string]*tfe.StateVersion),
		workspaces:    make(map[string][]string),
	}
}

func (m *MockStateVersions) List(ctx context.Context, options tfe.StateVersionListOptions) ([]*tfe.StateVersion, error) {
	var svs []*tfe.StateVersion
	for _, sv := range m.stateVersions {
		svs = append(svs, sv)
	}
	return svs, nil
}

func (m *MockStateVersions) Create(ctx context.Context, workspaceID string, options tfe.StateVersionCreateOptions) (*tfe.StateVersion, error) {
	id := generateID("sv-")
	url := fmt.Sprintf("https://app.terraform.io/_archivist/%s", id)

	sv := &tfe.StateVersion{
		ID:          id,
		DownloadURL: url,
		Serial:      *options.Serial,
	}

	state, err := base64.StdEncoding.DecodeString(*options.State)
	if err != nil {
		return nil, err
	}

	m.states[sv.DownloadURL] = state
	m.stateVersions[sv.ID] = sv
	m.workspaces[workspaceID] = append(m.workspaces[workspaceID], sv.ID)

	return sv, nil
}

func (m *MockStateVersions) Read(ctx context.Context, svID string) (*tfe.StateVersion, error) {
	sv, ok := m.stateVersions[svID]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	return sv, nil
}

func (m *MockStateVersions) Current(ctx context.Context, workspaceID string) (*tfe.StateVersion, error) {
	svs, ok := m.workspaces[workspaceID]
	if !ok || len(svs) == 0 {
		return nil, tfe.ErrResourceNotFound
	}
	sv, ok := m.stateVersions[svs[len(svs)-1]]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	return sv, nil
}

func (m *MockStateVersions) Download(ctx context.Context, url string) ([]byte, error) {
	state, ok := m.states[url]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	return state, nil
}

type MockWorkspaces struct {
	workspaceIDs   map[string]*tfe.Workspace
	workspaceNames map[string]*tfe.Workspace
}

func NewMockWorkspaces() *MockWorkspaces {
	return &MockWorkspaces{
		workspaceIDs:   make(map[string]*tfe.Workspace),
		workspaceNames: make(map[string]*tfe.Workspace),
	}
}

func (m *MockWorkspaces) List(ctx context.Context, organization string, options tfe.WorkspaceListOptions) ([]*tfe.Workspace, error) {
	var ws []*tfe.Workspace
	for _, w := range m.workspaceIDs {
		ws = append(ws, w)
	}
	return ws, nil
}

func (m *MockWorkspaces) Create(ctx context.Context, organization string, options tfe.WorkspaceCreateOptions) (*tfe.Workspace, error) {
	id := generateID("ws-")
	w := &tfe.Workspace{
		ID:   id,
		Name: *options.Name,
	}
	m.workspaceIDs[w.ID] = w
	m.workspaceNames[w.Name] = w
	return w, nil
}

func (m *MockWorkspaces) Read(ctx context.Context, organization, workspace string) (*tfe.Workspace, error) {
	w, ok := m.workspaceNames[workspace]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	return w, nil
}

func (m *MockWorkspaces) Update(ctx context.Context, organization, workspace string, options tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	w, ok := m.workspaceNames[workspace]
	if !ok {
		return nil, tfe.ErrResourceNotFound
	}
	w.Name = *options.Name
	w.TerraformVersion = *options.TerraformVersion

	delete(m.workspaceNames, workspace)
	m.workspaceNames[w.Name] = w

	return w, nil
}

func (m *MockWorkspaces) Delete(ctx context.Context, organization, workspace string) error {
	if w, ok := m.workspaceNames[workspace]; ok {
		delete(m.workspaceIDs, w.ID)
	}
	delete(m.workspaceNames, workspace)
	return nil
}

func (m *MockWorkspaces) Lock(ctx context.Context, workspaceID string, options tfe.WorkspaceLockOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *MockWorkspaces) Unlock(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *MockWorkspaces) AssignSSHKey(ctx context.Context, workspaceID string, options tfe.WorkspaceAssignSSHKeyOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *MockWorkspaces) UnassignSSHKey(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
	panic("not implemented")
}

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateID(s string) string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return s + string(b)
}
