package decoder

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"testing"
)

func manifestWith(kind, providerPath string) []byte {
	return []byte("" +
		"version: \"0.0.1\"\n" +
		"type: plugin\n" +
		"author: test\n" +
		"name: demo\n" +
		"label:\n  en_US: demo\n" +
		"description:\n  en_US: demo\n" +
		"icon: icon.svg\n" +
		"resource:\n  memory: 134217728\n" +
		"plugins:\n  " + kind + ":\n    - \"" + providerPath + "\"\n" +
		"meta:\n  version: \"0.0.1\"\n  arch: [amd64]\n  runner:\n    language: python\n    version: \"3.10\"\n    entrypoint: run.sh\n")
}

// ----- Tool provider tests -----
func toolProviderYAML(win bool) []byte {
	p := "tools/tool1.yaml"
	if win {
		p = "tools\\tool1.yaml"
	}
	return []byte("" +
		"identity:\n  author: test\n  name: tp\n  label:\n    en_US: tp\n  icon: icon.svg\n" +
		"tools:\n  - '" + p + "'\n")
}

func toolYAML() []byte {
	return []byte("" +
		"identity:\n  author: test\n  name: t1\n  label:\n    en_US: t1\n" +
		"description:\n  human:\n    en_US: x\n  llm: x\nparameters: []\n")
}

func TestToolProvider_WindowsPaths_FSAndZip(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "icon.svg", []byte("x"))
	// FS unix provider
	writeFile(t, root, "manifest.yaml", manifestWith("tools", "tool_provider.yaml"))
	writeFile(t, root, "tool_provider.yaml", toolProviderYAML(false))
	writeFile(t, root, filepath.Join("tools", "tool1.yaml"), toolYAML())

	dec, err := NewFSPluginDecoder(root)
	if err != nil {
		t.Fatalf("fs unix init: %v", err)
	}
	m, err := dec.Manifest()
	if err != nil {
		t.Fatalf("fs unix manifest: %v", err)
	}
	if m.Tool == nil || len(m.Tool.Tools) != 1 {
		t.Fatalf("want 1 tool, got %d", len(m.Tool.Tools))
	}

	// FS windows provider
	writeFile(t, root, "manifest.yaml", manifestWith("tools", "tool_provider_win.yaml"))
	writeFile(t, root, "tool_provider_win.yaml", toolProviderYAML(true))
	dec2, err := NewFSPluginDecoder(root)
	if err != nil {
		t.Fatalf("fs win init: %v", err)
	}
	m2, err := dec2.Manifest()
	if err != nil {
		t.Fatalf("fs win manifest: %v", err)
	}
	if m2.Tool == nil || len(m2.Tool.Tools) != 1 {
		t.Fatalf("want 1 tool (win), got %d", len(m2.Tool.Tools))
	}

	// ZIP windows provider
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	add := func(name string, data []byte) {
		f, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := f.Write(data); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	add("icon.svg", []byte("x"))
	add("manifest.yaml", manifestWith("tools", "tool_provider.yaml"))
	add("tool_provider.yaml", toolProviderYAML(true))
	add("tools/tool1.yaml", toolYAML())
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}

	zdec, err := NewZipPluginDecoder(buf.Bytes())
	if err != nil {
		t.Fatalf("zip init: %v", err)
	}
	zm, err := zdec.Manifest()
	if err != nil {
		t.Fatalf("zip manifest: %v", err)
	}
	if zm.Tool == nil || len(zm.Tool.Tools) != 1 {
		t.Fatalf("want 1 tool (zip), got %d", len(zm.Tool.Tools))
	}
}

// ----- Endpoint provider tests -----
func endpointProviderYAML(win bool) []byte {
	p := "endpoints/get.yaml"
	if win {
		p = "endpoints\\get.yaml"
	}
	return []byte("endpoints:\n  - '" + p + "'\n")
}

func endpointYAML() []byte { return []byte("path: /hello\nmethod: GET\n") }

func TestEndpointProvider_WindowsPaths_FSAndZip(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "manifest.yaml", manifestWith("endpoints", "endpoint_provider.yaml"))
	writeFile(t, root, "endpoint_provider.yaml", endpointProviderYAML(false))
	writeFile(t, root, filepath.Join("endpoints", "get.yaml"), endpointYAML())

	dec, err := NewFSPluginDecoder(root)
	if err != nil {
		t.Fatalf("fs unix init: %v", err)
	}
	m, err := dec.Manifest()
	if err != nil {
		t.Fatalf("fs unix manifest: %v", err)
	}
	if m.Endpoint == nil || len(m.Endpoint.Endpoints) != 1 {
		t.Fatalf("want 1 endpoint, got %d", len(m.Endpoint.Endpoints))
	}

	writeFile(t, root, "manifest.yaml", manifestWith("endpoints", "endpoint_provider_win.yaml"))
	writeFile(t, root, "endpoint_provider_win.yaml", endpointProviderYAML(true))
	dec2, err := NewFSPluginDecoder(root)
	if err != nil {
		t.Fatalf("fs win init: %v", err)
	}
	m2, err := dec2.Manifest()
	if err != nil {
		t.Fatalf("fs win manifest: %v", err)
	}
	if m2.Endpoint == nil || len(m2.Endpoint.Endpoints) != 1 {
		t.Fatalf("want 1 endpoint (win), got %d", len(m2.Endpoint.Endpoints))
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	add := func(name string, data []byte) { f, _ := zw.Create(name); f.Write(data) }
	add("manifest.yaml", manifestWith("endpoints", "endpoint_provider.yaml"))
	add("endpoint_provider.yaml", endpointProviderYAML(true))
	add("endpoints/get.yaml", endpointYAML())
	_ = zw.Close()
	zdec, err := NewZipPluginDecoder(buf.Bytes())
	if err != nil {
		t.Fatalf("zip init: %v", err)
	}
	zm, err := zdec.Manifest()
	if err != nil {
		t.Fatalf("zip manifest: %v", err)
	}
	if zm.Endpoint == nil || len(zm.Endpoint.Endpoints) != 1 {
		t.Fatalf("want 1 endpoint (zip), got %d", len(zm.Endpoint.Endpoints))
	}
}

// ----- Trigger/Datasource/AgentStrategy smoke test with Windows paths -----
func triggerProviderYAML(win bool) []byte {
	p := "triggers/event.yaml"
	if win {
		p = "triggers\\event.yaml"
	}
	return []byte("identity:\n  author: a\n  name: tp\n  label:\n    en_US: t\n  icon: icon.svg\nsubscription_schema: []\n" +
		"events:\n  - '" + p + "'\n")
}

func triggerEventYAML() []byte {
	return []byte("identity:\n  author: a\n  name: e\n  label:\n    en_US: e\n" +
		"description:\n  en_US: d\n")
}

func datasourceProviderYAML(win bool) []byte {
	p := "datasources/d.yaml"
	if win {
		p = "datasources\\d.yaml"
	}
	return []byte("identity:\n  author: a\n  name: dp\n  label:\n    en_US: d\n  icon: icon.svg\nprovider_type: website_crawl\ncredentials_schema: []\n" +
		"datasources:\n  - '" + p + "'\n")
}

func datasourceYAML() []byte {
	return []byte("identity:\n  author: a\n  name: d\n  label:\n    en_US: d\n" +
		"description:\n  en_US: d\nparameters: []\n")
}

func agentProviderYAML(win bool) []byte {
	p := "strategies/s.yaml"
	if win {
		p = "strategies\\s.yaml"
	}
	return []byte("identity:\n  author: a\n  name: ap\n  label:\n    en_US: a\n  icon: icon.svg\n" +
		"strategies:\n  - '" + p + "'\n")
}

func agentStrategyYAML() []byte {
	return []byte("identity:\n  author: a\n  name: s\n  label:\n    en_US: s\n" +
		"description:\n  en_US: d\nparameters: []\nfeatures: []\n")
}

func TestOtherProviders_WindowsPaths_Zip(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	add := func(name string, data []byte) { f, _ := zw.Create(name); f.Write(data) }

	add("icon.svg", []byte("x"))
	add("manifest.yaml", []byte(""+
		"version: \"0.0.1\"\ntype: plugin\nauthor: test\nname: demo\nlabel:\n  en_US: demo\ndescription:\n  en_US: demo\nicon: icon.svg\nresource:\n  memory: 134217728\nplugins:\n  triggers:\n    - trigger_provider.yaml\n  datasources:\n    - datasource_provider.yaml\n  agent_strategies:\n    - agent_provider.yaml\nmeta:\n  version: \"0.0.1\"\n  arch: [amd64]\n  runner:\n    language: python\n    version: \"3.10\"\n    entrypoint: run.sh\n"))
	add("trigger_provider.yaml", triggerProviderYAML(true))
	add("datasource_provider.yaml", datasourceProviderYAML(true))
	add("agent_provider.yaml", agentProviderYAML(true))
	add("triggers/event.yaml", triggerEventYAML())
	add("datasources/d.yaml", datasourceYAML())
	add("strategies/s.yaml", agentStrategyYAML())
	_ = zw.Close()

	dec, err := NewZipPluginDecoder(buf.Bytes())
	if err != nil {
		t.Fatalf("zip init: %v", err)
	}
	m, err := dec.Manifest()
	if err != nil {
		t.Fatalf("zip manifest: %v", err)
	}
	if m.Trigger == nil || len(m.Trigger.Events) != 1 {
		t.Fatalf("want 1 event, got %d", len(m.Trigger.Events))
	}
	if m.Datasource == nil || len(m.Datasource.Datasources) != 1 {
		t.Fatalf("want 1 datasource, got %d", len(m.Datasource.Datasources))
	}
	if m.AgentStrategy == nil || len(m.AgentStrategy.Strategies) != 1 {
		t.Fatalf("want 1 strategy, got %d", len(m.AgentStrategy.Strategies))
	}
}
