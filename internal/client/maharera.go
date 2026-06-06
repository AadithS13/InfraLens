package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/infralens/infralens/internal/model"
)

const (
	baseURL = "https://maharerait.maharashtra.gov.in/api/maha-rera-public-view-project-registration-service/public/projectregistartion"
	authURL = "https://maharerait.maharashtra.gov.in/api/maha-rera-login-service/login/authenticatePublic"

	// Public view credentials (encrypted, hardcoded in the Angular frontend)
	publicViewUser = "U2FsdGVkX18wiIEtItrcwDDt9WRsC8ZkSL/H9nN/0R8OElYaU3TbVvNkHWK1L8rn"
	publicViewPass = "U2FsdGVkX1//DxUv3f29Wg1Saion+lM8Ju2Yz4rocrw="
)

type Client struct {
	http        *http.Client
	mu          sync.RWMutex
	token       string
	tokenExpiry time.Time
}

func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Authenticate(ctx context.Context) error {
	payload := map[string]string{
		"userName": publicViewUser,
		"password": publicViewPass,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://maharerait.maharashtra.gov.in")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("auth read: %w", err)
	}

	// MahaRERA auth response wraps the token in responseObject
	var result model.APIResponse[model.AuthResponse]
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("auth parse: %w, body: %s", err, data)
	}

	if result.ResponseObject.AccessToken == "" {
		return fmt.Errorf("auth: empty token in response: %s", data)
	}

	// token TTL is ~100min; refresh 2min early to avoid mid-crawl expiry
	c.mu.Lock()
	c.token = result.ResponseObject.AccessToken
	c.tokenExpiry = time.Now().Add(98 * time.Minute)
	c.mu.Unlock()

	return nil
}

func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.RLock()
	valid := c.token != "" && time.Now().Before(c.tokenExpiry)
	c.mu.RUnlock()
	if valid {
		return nil
	}
	return c.Authenticate(ctx)
}

func (c *Client) post(ctx context.Context, endpoint string, payload any, out any) error {
	if err := c.ensureToken(ctx); err != nil {
		return fmt.Errorf("token refresh: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/"+endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", "https://maharerait.maharashtra.gov.in")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST %s: status %d", endpoint, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}

var ErrNotFound = fmt.Errorf("project not found")

// API methods

func (c *Client) GetProjectGeneral(ctx context.Context, projectID int) (*model.ProjectGeneral, error) {
	var result model.APIResponse[*model.ProjectGeneral]
	err := c.post(ctx, "getProjectGeneralDetailsByProjectId", map[string]int{"projectId": projectID}, &result)
	if err != nil {
		return nil, err
	}
	if result.ResponseObject == nil {
		return nil, ErrNotFound
	}
	return result.ResponseObject, nil
}

func (c *Client) GetPromoterGeneral(ctx context.Context, userProfileID, projectID int) (*model.PromoterGeneral, error) {
	var result model.APIResponse[*model.PromoterGeneral]
	err := c.post(ctx, "fetchPromoterGeneralDetails", map[string]int{
		"userProfileId": userProfileID,
		"projectId":     projectID,
	}, &result)
	if err != nil {
		return nil, err
	}
	return result.ResponseObject, nil
}

func (c *Client) GetPromoterAssociation(ctx context.Context, projectID int) (*model.PromoterAssociation, error) {
	var result model.APIResponse[*model.PromoterAssociationResponse]
	err := c.post(ctx, "getProjectAndAssociatedPromoterDetails", map[string]string{"projectId": fmt.Sprintf("%d", projectID)}, &result)
	if err != nil {
		return nil, err
	}
	if result.ResponseObject == nil {
		return nil, nil
	}
	return &result.ResponseObject.PromoterDetails, nil
}

func (c *Client) GetProjectAddress(ctx context.Context, projectID int) (*model.ProjectAddress, error) {
	var result model.APIResponse[*model.ProjectAddress]
	err := c.post(ctx, "getProjectLandAddressDetails", map[string]int{"projectId": projectID}, &result)
	if err != nil {
		return nil, err
	}
	return result.ResponseObject, nil
}

func (c *Client) GetPromoterAddress(ctx context.Context, userProfileID, projectID int) (*model.PromoterAddress, error) {
	var result model.APIResponse[*model.PromoterAddress]
	err := c.post(ctx, "getPromoterAddressDetails", map[string]int{
		"userProfileId": userProfileID,
		"projectId":     projectID,
	}, &result)
	if err != nil {
		return nil, err
	}
	return result.ResponseObject, nil
}

func (c *Client) GetProfessionals(ctx context.Context, projectID int, typeName string) ([]model.Professional, error) {
	payload := map[string]any{"projectId": fmt.Sprintf("%d", projectID), "professionalTypeName": typeName}
	if typeName == "" {
		payload["professionalTypeName"] = nil
	}
	var result model.APIResponse[[]model.Professional]
	err := c.post(ctx, "getProjectProfessionalByType", payload, &result)
	if err != nil {
		return nil, err
	}
	return result.ResponseObject, nil
}

func (c *Client) GetAgents(ctx context.Context, projectID int) ([]model.Agent, error) {
	var result model.APIResponse[[]model.Agent]
	err := c.post(ctx, "getAgentByProjectId", map[string]string{"projectId": fmt.Sprintf("%d", projectID)}, &result)
	if err != nil {
		return nil, err
	}
	return result.ResponseObject, nil
}

func (c *Client) GetPromoterContacts(ctx context.Context, userProfileID, projectID int) ([]model.PersonnelContact, error) {
	var result model.APIResponse[[]model.PersonnelContact]
	err := c.post(ctx, "fetchPromoterPersonnelContactAddressDetails", map[string]int{
		"userProfileId": userProfileID,
		"projectId":     projectID,
	}, &result)
	if err != nil {
		return nil, err
	}
	return result.ResponseObject, nil
}
