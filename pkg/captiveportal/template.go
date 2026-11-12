package captiveportal

import (
	"context"
	"fmt"
	"strings"

	"github.com/browningluke/opnsense-go/pkg/api"
)

// Captive portal templates are zip files (base64) with a custom login page.
// OPNsense 26.7 serves them from captiveportal/template/*; earlier releases
// from captiveportal/service/*Template. The calls try the 26.7 path first and
// fall back on a 404.

const (
	templateBase       = "/captiveportal/template/"
	templateBaseLegacy = "/captiveportal/service/"
)

type Template struct {
	Uuid    string `json:"uuid"`
	Name    string `json:"name"`
	FileId  string `json:"fileid"`
	Content string `json:"content"`
}

type templateSaveResult struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

type templateSearchResult struct {
	Rows []Template `json:"rows"`
}

type templateDeleteResult struct {
	Result string `json:"result"`
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "404")
}

func templateCall[R any](c *Controller, ctx context.Context, action string, path []string, body map[string]interface{}, result *R) error {
	call := func(base string) error {
		opts := api.RPCOpts{
			BaseEndpoint:    base + action,
			Method:          "POST",
			PathParameters:  path,
			QueryParameters: map[string]string{},
			BodyParameters:  body,
		}
		_, err := api.Call(c.Client(), ctx, opts, result)
		return err
	}
	err := call(templateBase)
	if isNotFound(err) {
		err = call(templateBaseLegacy)
	}
	return err
}

// SaveTemplate creates the template, or replaces the one with uuid when uuid
// is set, and returns its UUID. content is the base64 zip; OPNsense strips
// files identical to its default template before storing it.
func (c *Controller) SaveTemplate(ctx context.Context, uuid, name, content string) (string, error) {
	body := map[string]interface{}{"name": name, "content": content}
	if uuid != "" {
		body["uuid"] = uuid
	}
	res := &templateSaveResult{}
	if err := templateCall(c, ctx, "save_template", nil, body, res); err != nil {
		return "", fmt.Errorf("save template: %w", err)
	}
	if res.Error != "" {
		return "", fmt.Errorf("save template %q: %s", name, res.Error)
	}
	if uuid != "" {
		return uuid, nil
	}
	t, err := c.findTemplate(ctx, func(t Template) bool { return t.Name == name })
	if err != nil {
		return "", err
	}
	if t == nil {
		return "", fmt.Errorf("template %q saved but not found", name)
	}
	return t.Uuid, nil
}

// GetTemplate returns the template, or nil when it doesn't exist.
func (c *Controller) GetTemplate(ctx context.Context, uuid string) (*Template, error) {
	return c.findTemplate(ctx, func(t Template) bool { return t.Uuid == uuid })
}

func (c *Controller) findTemplate(ctx context.Context, match func(Template) bool) (*Template, error) {
	res := &templateSearchResult{}
	body := map[string]interface{}{"current": 1, "rowCount": -1}
	if err := templateCall(c, ctx, "search_templates", nil, body, res); err != nil {
		return nil, fmt.Errorf("search templates: %w", err)
	}
	for _, t := range res.Rows {
		if match(t) {
			found := t
			return &found, nil
		}
	}
	return nil, nil
}

// DeleteTemplate removes the template and reloads the captive portal.
func (c *Controller) DeleteTemplate(ctx context.Context, uuid string) error {
	res := &templateDeleteResult{}
	if err := templateCall(c, ctx, "del_template", []string{uuid}, nil, res); err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	if res.Result != "deleted" {
		return fmt.Errorf("delete template: result %q", res.Result)
	}
	return c.Client().ReconfigureService(ctx, "/captiveportal/service/reconfigure")
}
