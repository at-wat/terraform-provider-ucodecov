package codecov

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type Owner struct {
	Username    string `json:"username"`
	Name        string `json:"name"`
	Service     string `json:"service"`
	Updatestamp string `json:"updatestamp"`
	AvatarURL   string `json:"avatar_url"`
	ServiceID   string `json:"service_id"`
}

type Repo struct {
	UsingIntegration bool   `json:"using_integration"`
	Name             string `json:"name"`
	Language         string `json:"language"`
	Deleted          bool   `json:"deleted"`
	BotUsername      string `json:"bot_username"`
	Activated        bool   `json:"activated"`
	Private          bool   `json:"private"`
	Updatestamp      string `json:"updatestamp"`
	Branch           string `json:"branch"`
	UploadToken      string `json:"upload_token"`
	Active           bool   `json:"active"`
	ServiceID        string `json:"service_id"`
	ImageToken       string `json:"image_token"`
}

// Settings represents json format returned from https://codecov.io/api/pub/gh/owner/repo/settings.
type Settings struct {
	Owner Owner `json:"owner"`
	Repo  Repo  `json:"repo"`
}

func dataSourceCodecovSettings() *schema.Resource {
	return &schema.Resource{
		Read: dataCodecovSettingsRead,
		Schema: map[string]*schema.Schema{
			"service": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repo": {
				Type:     schema.TypeString,
				Required: true,
			},
			"updatestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"upload_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataCodecovSettingsRead(d *schema.ResourceData, meta interface{}) error {
	service := d.Get("service").(string)
	owner := d.Get("owner").(string)
	repo := d.Get("repo").(string)

	token, ok := os.LookupEnv("CODECOV_API_TOKEN")
	if !ok {
		return errors.New("CODECOV_API_TOKEN is not given")
	}

	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://codecov.io/api/pub/%s/%s/%s/settings",
			service, owner, repo,
		),
		http.NoBody,
	)
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	dec := json.NewDecoder(resp.Body)

	var s Settings
	if err := dec.Decode(&s); err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s/%s/%s", service, owner, repo))
	d.Set("updatestamp", s.Repo.Updatestamp)
	d.Set("upload_token", s.Repo.UploadToken)
	return nil
}
