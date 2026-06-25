package browser

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Browser struct {
	rodBrowser *rod.Browser
	page       *rod.Page
}

func NewBrowser(proxyAddress string) (*Browser, error) {
	l := launcher.New()
	l.Headless(false)
	l.Leakless(false)

	if proxyAddress != "" {
    	l.Proxy(proxyAddress)
	}

	u, err := l.Launch()
	if err != nil {
		return nil, err
	}

	rodBrowser := rod.New().ControlURL(u)

	err = rodBrowser.Connect()
	if err != nil {
		return nil, err
	}

    return &Browser{
		rodBrowser: rodBrowser,
	}, nil
}

func (b *Browser) GetPage() *rod.Page {
	return b.page
}

func (b *Browser) Close() error {
	if err := b.rodBrowser.Close(); err != nil {
		return err
	}

	return nil
}

func (b *Browser) Navigate(url string) error {
	if b.page == nil {
		p, err := b.rodBrowser.Page(proto.TargetCreateTarget{}) 
		if err != nil {
			return err
		}
		b.page = p
	}

	err := b.page.Navigate(url)
	if err != nil {
		return err
	}

	return nil
}