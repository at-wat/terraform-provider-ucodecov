package codecov

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	maxRetry      = 6
	retryWaitBase = time.Second
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
		return errors.New("codecov: CODECOV_API_TOKEN is not given")
	}

	var (
		s   *Settings
		err error
	)
	wait := retryWaitBase
	for i := 0; ; i++ {
		s, err = readRepoSetting(service, owner, repo, token)
		if err == nil {
			d.SetId(fmt.Sprintf("%s/%s/%s", service, owner, repo))
			d.Set("updatestamp", s.Repo.Updatestamp)
			d.Set("upload_token", s.Repo.UploadToken)
			return nil
		}
		var netErr net.Error
		if errors.As(err, &netErr) {
			if !netErr.Timeout() && !netErr.Temporary() {
				// Immediately exit if non-timeout error is returned.
				return fmt.Errorf("codecov: %w", err)
			}
		} else {
			// Immediately exit if non-network error is returned.
			return fmt.Errorf("codecov: %w", err)
		}
		if i >= maxRetry {
			return fmt.Errorf("codecov: %w", err)
		}
		time.Sleep(wait)
		wait *= 2
	}
}

func readRepoSetting(service, owner, repo, token string) (*Settings, error) {
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
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return nil, &timeoutError{errors.New(resp.Status)}
	default:
		return nil, errors.New(resp.Status)
	}

	dec := json.NewDecoder(resp.Body)

	var s Settings
	if err := dec.Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

type timeoutError struct {
	error
}

func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
