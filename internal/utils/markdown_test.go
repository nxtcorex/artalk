package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkedRendersCommonMarkdown(t *testing.T) {
	html, err := Marked(`
# Title

[Artalk](https://example.com)

![logo](https://example.com/logo.png)

- first
- second

| Name | Value |
| --- | --- |
| alpha | 1 |
`)

	require.NoError(t, err)
	assert.Contains(t, html, `<h1 id="title">Title</h1>`)
	assert.Contains(t, html, `<a href="https://example.com"`)
	assert.Contains(t, html, `rel="nofollow noreferrer"`)
	assert.Contains(t, html, `<img src="https://example.com/logo.png" alt="logo"/>`)
	assert.Contains(t, html, `<ul>`)
	assert.Contains(t, html, `<li>first</li>`)
	assert.Contains(t, html, `<table>`)
	assert.Contains(t, html, `<td>alpha</td>`)
}

func TestMarkedKeepsAllowedCommentHTMLAttributes(t *testing.T) {
	html, err := Marked(`
<p align="center" class="lead" style="color:red">paragraph</p>
<span align="right" class="hljs-keyword" style="color:#abc">span</span>
<img src="https://example.com/emote.png" width="32" height="32" align="left" atk-emoticon="smile" alt="smile" />
`)

	require.NoError(t, err)
	assert.Contains(t, html, `align="center"`)
	assert.Contains(t, html, `class="lead"`)
	assert.Contains(t, html, `style="color:red"`)
	assert.Contains(t, html, `align="right"`)
	assert.Contains(t, html, `class="hljs-keyword"`)
	assert.Contains(t, html, `style="color:#abc"`)
	assert.Contains(t, html, `width="32"`)
	assert.Contains(t, html, `height="32"`)
	assert.Contains(t, html, `atk-emoticon="smile"`)
}

func TestMarkedRemovesExecutableHTMLAndDangerousURLs(t *testing.T) {
	html, err := Marked(`
<script>alert(1)</script>
<a href="javascript:alert(1)" onclick="alert(1)">bad link</a>
<img src="javascript:alert(1)" onerror="alert(1)" alt="bad image" />
`)

	require.NoError(t, err)
	assert.NotContains(t, html, `<script`)
	assert.NotContains(t, html, `javascript:`)
	assert.NotContains(t, html, `onclick`)
	assert.NotContains(t, html, `onerror`)
	assert.Contains(t, html, `bad link`)
	assert.Contains(t, html, `alt="bad image"`)
}

func TestMarkedRejectsUnsafeDataURIs(t *testing.T) {
	html, err := Marked(`
<p><a href="data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==">click</a></p>
<p><a href="data:text/javascript,alert(1)">js</a></p>
<p><img src="data:image/svg+xml;base64,PHN2ZyBvbmxvYWQ9YWxlcnQoMSk+" alt="svg" /></p>
<p><img src="data:application/javascript,alert(1)" alt="app js" /></p>
`)

	require.NoError(t, err)
	assert.NotContains(t, html, `href="data:`)
	assert.NotContains(t, html, `src="data:`)
	assert.NotContains(t, html, `data:text/html`)
	assert.NotContains(t, html, `data:text/javascript`)
	assert.NotContains(t, html, `data:image/svg+xml`)
	assert.NotContains(t, html, `data:application/javascript`)
	assert.Contains(t, html, `click`)
	assert.Contains(t, html, `js`)
	assert.Contains(t, html, `alt="svg"`)
	assert.Contains(t, html, `alt="app js"`)
}
