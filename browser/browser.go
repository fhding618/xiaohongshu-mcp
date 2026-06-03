package browser

import (
	"encoding/json"
	"net/url"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

type browserConfig struct {
	binPath string
}

type Option func(*browserConfig)

type Browser struct {
	rodBrowser *rod.Browser
	launcher   *launcher.Launcher
}

func WithBinPath(binPath string) Option {
	return func(c *browserConfig) {
		c.binPath = binPath
	}
}

// maskProxyCredentials masks username and password in proxy URL for safe logging.
func maskProxyCredentials(proxyURL string) string {
	u, err := url.Parse(proxyURL)
	if err != nil || u.User == nil {
		return proxyURL
	}
	if _, hasPassword := u.User.Password(); hasPassword {
		u.User = url.UserPassword("***", "***")
	} else {
		u.User = url.User("***")
	}
	return u.String()
}

func NewBrowser(headless bool, options ...Option) *Browser {
	cfg := &browserConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	l := launcher.New().
		Headless(headless).
		Set("--no-sandbox")

	if cfg.binPath != "" {
		l = l.Bin(cfg.binPath)
	}

	// Read proxy from environment variable
	if proxy := os.Getenv("XHS_PROXY"); proxy != "" {
		l = l.Proxy(proxy)
		logrus.Infof("Using proxy: %s", maskProxyCredentials(proxy))
	}

	url := l.MustLaunch()
	rodBrowser := rod.New().
		ControlURL(url).
		NoDefaultDevice().
		MustConnect()

	cookiePath := cookies.GetCookiesFilePath()
	cookieLoader := cookies.NewLoadCookie(cookiePath)
	if data, err := cookieLoader.LoadCookies(); err == nil {
		var cks []*proto.NetworkCookie
		if err := json.Unmarshal(data, &cks); err != nil {
			logrus.Warnf("failed to unmarshal cookies: %v", err)
		} else {
			rodBrowser.MustSetCookies(cks...)
		}
		logrus.Debugf("loaded cookies from filesuccessfully")
	} else {
		logrus.Warnf("failed to load cookies: %v", err)
	}

	return &Browser{
		rodBrowser: rodBrowser,
		launcher:   l,
	}
}

func (b *Browser) Close() {
	b.rodBrowser.MustClose()
	b.launcher.Cleanup()
}

func (b *Browser) NewPage() *rod.Page {
	return stealth.MustPage(b.rodBrowser)
}
