//main.go
package main

import (
	"github.com/hashicorp/terraform/builtin/providers/sky"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: sky.Provider,
	})
}

//provider.go
package sky

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"sky_latestimage":     dataSourceHttp(),
		},
		ResourcesMap: map[string]*schema.Resource{
                        "sky_latestimage": schema.DataSourceResourceShim(
                                "sky_latestimage",
                                dataSourceHttp(),
                        ),
		},
	}
}

//datasource_sky_latestimage.go

package sky

import (
        "crypto/sha256"
        "encoding/hex"
        "encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceHttp() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceHttpRead,

		Schema: map[string]*schema.Schema{
			"url": &schema.Schema{
				Type:          schema.TypeString,
				Required:      true,
				Description:   "HTTP URL address to connect",
				ForceNew:      true,
			  ValidateFunc:  validateHttpTemplateAttribute,
			},
			"rendered": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "all payload",
			},
      			"latest": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "latest version json value",
			},
		},
	}
}

type sky_template_resources struct {
	Node struct {
    		Value string `"json:value"`
	} `"json:node"`
}

func json_unmarshall(rendered string) (sky_template_resources, error) {

  sky_template_json := sky_template_resources{}

  err := json.Unmarshal([]byte(rendered), &sky_template_json)
  return sky_template_json, err
}

func dataSourceHttpRead(d *schema.ResourceData, meta interface{}) error {
	rendered, err := renderURL(d)
	if err != nil {
		return err
	}
	d.Set("rendered", rendered)
	d.SetId(hash(rendered))

	sky_json, err := json_unmarshall(rendered)
  	if err != nil {
    		return err
  	}

        d.Set("latest", sky_json.Node.Value)
        d.SetId(hash(sky_json.Node.Value))

	return nil
}

func httpRead(url string) (string, bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", false ,err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	return string(body), true, err

}

func renderURL(d *schema.ResourceData) (string, error) {
	url := d.Get("url").(string)

	rendered, _, err := httpRead(url)
	if err != nil {
		return "", err
	}

	return rendered, nil
}

func hash(s string) string {
        sha := sha256.Sum256([]byte(s))
        return hex.EncodeToString(sha[:])
}

func validateHttpTemplateAttribute(v interface{}, key string) (ws []string, es []error) {
	_, wasPath, err := httpRead(v.(string))
	if err != nil {
		es = append(es, err)
		return
	}

	if wasPath {
		ws = append(ws, fmt.Sprintf("%s: looks like you specified a worng URL.", key))
	}

	return
}
