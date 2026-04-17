package web

import (
	"bytes"
	htemplate "html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/users"
)

type testWebPlugin struct {
	html string
	data map[string]any
	ok   bool
}

func (p testWebPlugin) NavLinks() map[string]string {
	return map[string]string{}
}

func (p testWebPlugin) WebRequest(r *http.Request) (string, map[string]any, bool) {
	return p.html, p.data, p.ok
}

func TestIsAllowedWebSocketOrigin(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		origin  string
		allowed bool
	}{
		{
			name:    "request host is allowed",
			host:    "play.example.com",
			origin:  "https://play.example.com",
			allowed: true,
		},
		{
			name:    "same host different port is rejected",
			host:    "localhost:80",
			origin:  "http://localhost:3000",
			allowed: false,
		},
		{
			name:    "exact host and port is allowed",
			host:    "localhost:3000",
			origin:  "http://localhost:3000",
			allowed: true,
		},
		{
			name:    "foreign origin is rejected",
			host:    "localhost:80",
			origin:  "https://evil.example",
			allowed: false,
		},
		{
			name:    "missing origin is allowed",
			host:    "localhost:80",
			origin:  "",
			allowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://"+tt.host+"/ws", nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			if got := isAllowedWebSocketOrigin(req); got != tt.allowed {
				t.Fatalf("isAllowedWebSocketOrigin() = %v, want %v", got, tt.allowed)
			}
		})
	}
}

func TestOnlineTemplateEscapesCharacterName(t *testing.T) {
	publicHTML := filepath.Clean(filepath.Join("..", "..", "_datafiles", "html", "public"))
	tmpl, err := htemplate.New("online.html").Funcs(funcMap).ParseFiles(
		publicHTML+"/_header.html",
		publicHTML+"/online.html",
		publicHTML+"/_footer.html",
	)
	if err != nil {
		t.Fatalf("ParseFiles() error = %v", err)
	}

	payload := `<script>alert("xss")</script>`
	templateData := map[string]any{
		"REQUEST": httptest.NewRequest("GET", "http://localhost/online", nil),
		"PATH":    "/online",
		"CONFIG":  configs.GetConfig(),
		"STATS": Stats{
			OnlineUsers: []users.OnlineInfo{
				{CharacterName: payload},
			},
		},
		"NAV": []WebNav{
			{Name: "Who's Online", Target: "/online"},
		},
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateData); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := rendered.String()
	if strings.Contains(output, payload) {
		t.Fatalf("rendered output still contains raw payload: %q", payload)
	}

	if !strings.Contains(output, `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`) {
		t.Fatalf("rendered output did not contain escaped payload: %s", output)
	}
}

func TestEscapeHTMLDoesNotDoubleEscapeTemplateOutput(t *testing.T) {
	tmpl, err := htemplate.New("test").Funcs(funcMap).Parse(`{{ escapehtml .Value }}`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	templateData := map[string]any{
		"Value": `Tom & Jerry "quoted"`,
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateData); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := rendered.String()
	if strings.Contains(output, "&amp;quot;") {
		t.Fatalf("rendered output double-escaped quotes: %s", output)
	}

	if got, want := output, `Tom &amp; Jerry &#34;quoted&#34;`; got != want {
		t.Fatalf("rendered output = %q, want %q", got, want)
	}
}

func TestRPadHTMLPreservesNonBreakingSpaces(t *testing.T) {
	tmpl, err := htemplate.New("test").Funcs(funcMap).Parse(`{{ rpadhtml 4 "x" "&nbsp;" }}`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, nil); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got, want := rendered.String(), `x&nbsp;&nbsp;&nbsp;`; got != want {
		t.Fatalf("rendered output = %q, want %q", got, want)
	}
}

func TestItemScriptRendersWithoutDoubleEscaping(t *testing.T) {
	adminHTML := filepath.Clean(filepath.Join("..", "..", "_datafiles", "html", "admin"))
	tmpl, err := htemplate.New("item.data.html").Funcs(funcMap).ParseFiles(adminHTML + "/items/item.data.html")
	if err != nil {
		t.Fatalf("ParseFiles() error = %v", err)
	}

	script := `if (a < b && c == "d") { console.log("<ok>") }`
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, map[string]any{
		"itemSpec":     items.ItemSpec{},
		"buffSpecs":    nil,
		"itemTypes":    nil,
		"itemSubtypes": nil,
		"script":       script,
	}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := rendered.String()
	if strings.Contains(output, "&amp;lt;") || strings.Contains(output, "&amp;quot;") {
		t.Fatalf("rendered output double-escaped script: %s", output)
	}
	if !strings.Contains(output, `if (a &lt; b &amp;&amp; c == &#34;d&#34;) { console.log(&#34;&lt;ok&gt;&#34;) }`) {
		t.Fatalf("rendered output did not contain correctly escaped script body: %s", output)
	}
}

func TestAdminPickerDataContentPreservesMarkupAndEscapesUserText(t *testing.T) {
	itemAttr := adminItemDataContent(items.ItemSpec{
		ItemId:     12,
		Name:       `<b>Blade & "Dagger"</b>`,
		QuestToken: "quest",
		Cursed:     true,
	})
	itemOutput := string(itemAttr)
	if !strings.Contains(itemOutput, `<span class='badge badge-secondary'>12</span>`) {
		t.Fatalf("item data-content lost rich HTML markup: %s", itemOutput)
	}
	if strings.Contains(itemOutput, `<b>Blade`) {
		t.Fatalf("item data-content included raw user text: %s", itemOutput)
	}
	if !strings.Contains(itemOutput, `&lt;b&gt;Blade &amp; &#34;Dagger&#34;&lt;/b&gt;`) {
		t.Fatalf("item data-content did not escape user text: %s", itemOutput)
	}

	mobAttr := adminMobDataContent(mobs.Mob{
		MobId:      34,
		QuestFlags: []string{"quest"},
		Character:  characters.Character{Name: `Sneak <script>alert(1)</script>`},
	})
	mobOutput := string(mobAttr)
	if !strings.Contains(mobOutput, `<span class='text-warning'>&#x2605;</span>`) {
		t.Fatalf("mob data-content lost quest badge markup: %s", mobOutput)
	}
	if strings.Contains(mobOutput, `<script>alert(1)</script>`) {
		t.Fatalf("mob data-content included raw script text: %s", mobOutput)
	}
	if !strings.Contains(mobOutput, `Sneak &lt;script&gt;alert(1)&lt;/script&gt;`) {
		t.Fatalf("mob data-content did not escape name: %s", mobOutput)
	}
}

func TestAdminPickerDataContentRendersAsTrustedHTMLAttr(t *testing.T) {
	tmpl, err := htemplate.New("picker").Parse(`<option {{ .Attr }} value="1">Item</option>`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, map[string]any{
		"Attr": adminItemDataContent(items.ItemSpec{
			ItemId: 7,
			Name:   `Widget`,
		}),
	}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := rendered.String()
	if !strings.Contains(output, `data-content="<span class='badge badge-secondary'>7</span> <span class='font-weight-bold'>Widget</span>"`) {
		t.Fatalf("rendered output lost trusted data-content markup: %s", output)
	}
	if strings.Contains(output, `ZgotmplZ`) {
		t.Fatalf("rendered output rejected trusted attribute: %s", output)
	}
}

func TestServeTemplatePreservesTrustedPluginHTML(t *testing.T) {
	mudlog.SetupLogger(nil, "", "", false)

	prevRoot := httpRoot
	prevPlugin := webPlugins
	httpRoot = t.TempDir()
	if err := os.WriteFile(filepath.Join(httpRoot, "_test_include.html"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	webPlugins = testWebPlugin{
		html: `<pre>{{ .contents }}</pre>`,
		data: map[string]any{
			"contents": htemplate.HTML(`<span class="ansi">formatted</span>`),
		},
		ok: true,
	}
	t.Cleanup(func() {
		httpRoot = prevRoot
		webPlugins = prevPlugin
	})

	req := httptest.NewRequest("GET", "http://localhost/help-details", nil)
	rec := httptest.NewRecorder()

	serveTemplate(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `<pre><span class="ansi">formatted</span></pre>`) {
		t.Fatalf("trusted plugin html was escaped: %s", body)
	}
}
