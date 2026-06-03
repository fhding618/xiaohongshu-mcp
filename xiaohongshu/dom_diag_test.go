package xiaohongshu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/browser"
)

func TestDiagnosePublishButtonDOM(t *testing.T) {
	binPath := os.Getenv("ROD_BROWSER_BIN")
	if binPath == "" {
		binPath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}
	b := browser.NewBrowser(false, browser.WithBinPath(binPath))
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	pp := page.Timeout(30 * time.Second)
	if err := pp.Navigate(urlOfPublic); err != nil {
		t.Fatalf("navigate publish page: %v", err)
	}
	if err := pp.WaitLoad(); err != nil {
		t.Logf("wait load: %v", err)
	}
	time.Sleep(8 * time.Second)

	if err := mustClickPublishTab(pp, "上传图文"); err != nil {
		t.Fatalf("click image-text tab: %v", err)
	}
	time.Sleep(2 * time.Second)

	imagePath, err := filepath.Abs("../assets/search_result.png")
	if err != nil {
		t.Fatalf("abs image path: %v", err)
	}
	if err := uploadImages(pp, []string{imagePath}); err != nil {
		t.Fatalf("upload image: %v", err)
	}
	time.Sleep(5 * time.Second)

	titleElem, err := pp.Element("div.d-input input")
	if err != nil {
		t.Fatalf("find title input: %v", err)
	}
	if err := titleElem.Input("DOM 诊断测试标题"); err != nil {
		t.Fatalf("input title: %v", err)
	}
	contentElem, ok := getContentElement(pp)
	if !ok {
		t.Fatalf("find content input")
	}
	if err := contentElem.Input("DOM 诊断测试正文"); err != nil {
		t.Fatalf("input content: %v", err)
	}
	time.Sleep(3 * time.Second)
	_, _ = pp.Eval(`() => window.scrollTo(0, document.body.scrollHeight)`)
	time.Sleep(2 * time.Second)

	result, err := pp.Eval(`() => {
		const currentSelector = '.publish-page-publish-btn button.bg-red';
		const legacy = document.querySelector(currentSelector);
		const normalize = (s) => (s || '').replace(/\s+/g, ' ').trim();
		const describe = (el) => {
			const rect = el.getBoundingClientRect();
			return {
				tag: el.tagName,
				text: normalize(el.textContent),
				className: String(el.className || ''),
				id: el.id || '',
				role: el.getAttribute('role') || '',
				type: el.getAttribute('type') || '',
				disabledAttr: el.hasAttribute('disabled'),
				ariaDisabled: el.getAttribute('aria-disabled') || '',
				visible: !!(rect.width && rect.height),
				rect: {
					x: Math.round(rect.x),
					y: Math.round(rect.y),
					width: Math.round(rect.width),
					height: Math.round(rect.height),
				},
				outerHTML: el.outerHTML.slice(0, 500),
			};
		};

		const candidates = Array.from(document.querySelectorAll(
			'button, [role="button"], .d-button, [class*="btn"], [class*="button"]'
		)).filter((el) => {
			const text = normalize(el.textContent);
			const cls = String(el.className || '');
			return text.includes('发布') || cls.includes('publish') || cls.includes('Publish');
		}).map(describe);

		return {
			url: location.href,
			title: document.title,
			legacySelector: currentSelector,
			legacyFound: !!legacy,
			legacy: legacy ? describe(legacy) : null,
			candidates,
			publishContainers: Array.from(document.querySelectorAll('[class*="publish"], [class*="Publish"]')).slice(0, 80).map(describe),
		};
	}`)
	if err != nil {
		t.Fatalf("eval dom: %v", err)
	}

	data, _ := json.MarshalIndent(result.Value, "", "  ")
	fmt.Println(string(data))
}
