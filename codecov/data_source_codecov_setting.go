package codecov

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	maxRetry      = 6
	retryWaitBase = time.Second
)

// Config represents json format returned from https://codecov.io/api/v2/gh/owner/repos/repo/config
type Config struct {
	UploadToken string `json:"upload_token"`
}

func dataSourceCodecovConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataCodecovConfigRead,
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
			"upload_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataCodecovConfigRead(d *schema.ResourceData, meta interface{}) error {
	service := d.Get("service").(string)
	owner := d.Get("owner").(string)
	repo := d.Get("repo").(string)
	token := meta.(string)
	if token == "" {
		return errors.New("codecov: CODECOV_API_V2_TOKEN is not given")
	}

	var (
		c   *Config
		err error
	)
	wait := retryWaitBase
	for i := 0; ; i++ {
		c, err = readRepoConfig(service, owner, repo, token)
		if err == nil {
			d.SetId(fmt.Sprintf("%s/%s/%s", service, owner, repo))
			d.Set("upload_token", c.UploadToken)
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

func readRepoConfig(service, owner, repo, token string) (*Config, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://codecov.io/api/v2/%s/%s/repos/%s/config/",
			service, owner, repo,
		),
		http.NoBody,
	)
	req.Header.Add("Authorization", fmt.Sprintf("bearer %s", token))

	cli := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusBadGateway:
		return nil, &timeoutError{errors.New(resp.Status)}
	case http.StatusFound, http.StatusTemporaryRedirect:
		// There was a bug that the request was redirected to the html setting page.
		// Wait extra 1 second and retry to workaround the problem.
		time.Sleep(time.Second)
		return nil, &timeoutError{errors.New(resp.Status)}
	default:
		return nil, errors.New(resp.Status)
	}

	dec := json.NewDecoder(resp.Body)

	var c Config
	if err := dec.Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

type timeoutError struct {
	error
}

func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
