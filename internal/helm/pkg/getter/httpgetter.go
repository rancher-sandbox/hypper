package getter

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/rancher-sandbox/hypper/internal/helm/tlsutil"
	"github.com/rancher-sandbox/hypper/internal/helm/urlutil"
	"github.com/rancher-sandbox/hypper/internal/helm/version"
)

// HTTPGetter is the default HTTP(/S) backend handler
type HTTPGetter struct {
	opts options
}

//Get performs a Get from repo.Getter and returns the body.
func (g *HTTPGetter) Get(href string, options ...Option) (*bytes.Buffer, error) {
	for _, opt := range options {
		opt(&g.opts)
	}
	return g.get(href)
}

func (g *HTTPGetter) get(href string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	// Set a helm specific user agent so that a repo server and metrics can
	// separate helm calls from other tools interacting with repos.
	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return buf, err
	}

	req.Header.Set("User-Agent", version.GetUserAgent())
	if g.opts.userAgent != "" {
		req.Header.Set("User-Agent", g.opts.userAgent)
	}

	if g.opts.username != "" && g.opts.password != "" {
		req.SetBasicAuth(g.opts.username, g.opts.password)
	}

	client, err := g.httpClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return buf, err
	}
	if resp.StatusCode != 200 {
		return buf, errors.Errorf("failed to fetch %s : %s", href, resp.Status)
	}

	_, err = io.Copy(buf, resp.Body)
	resp.Body.Close()
	return buf, err
}

// NewHTTPGetter constructs a valid http/https client as a Getter
func NewHTTPGetter(options ...Option) (Getter, error) {
	var client HTTPGetter

	for _, opt := range options {
		opt(&client.opts)
	}

	return &client, nil
}

func (g *HTTPGetter) httpClient() (*http.Client, error) {
	transport := &http.Transport{
		DisableCompression: true,
		Proxy:              http.ProxyFromEnvironment,
	}
	if (g.opts.certFile != "" && g.opts.keyFile != "") || g.opts.caFile != "" {
		tlsConf, err := tlsutil.NewClientTLS(g.opts.certFile, g.opts.keyFile, g.opts.caFile)
		if err != nil {
			return nil, errors.Wrap(err, "can't create TLS config for client")
		}
		tlsConf.BuildNameToCertificate()

		sni, err := urlutil.ExtractHostname(g.opts.url)
		if err != nil {
			return nil, err
		}
		tlsConf.ServerName = sni

		transport.TLSClientConfig = tlsConf
	}

	if g.opts.insecureSkipVerifyTLS {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		} else {
			transport.TLSClientConfig.InsecureSkipVerify = true
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   g.opts.timeout,
	}

	return client, nil
}
