# Telegram Rich Messages

Use this reference whenever generating, sending, editing, or streaming Telegram
Rich Messages with mtgo. Rich Messages are distinct from legacy message text,
Markdown, MarkdownV2, HTML parse modes, captions, and text entities.

Official reference:
<https://core.telegram.org/bots/api#rich-message-formatting-options>

## mtgo API

Send persistent Rich Markdown:

```go
msg, err := client.SendRichMessage(ctx, chatID, markdown, &params.SendMessage{
	ParseMode: params.ParseModeRichMarkdown,
	Silent:    true,
})
```

Send persistent Rich HTML:

```go
msg, err := client.SendRichMessage(ctx, chatID, html, &params.SendMessage{
	ParseMode: params.ParseModeRichHTML,
})
```

Edit a persistent rich message:

```go
msg, err := client.EditRichMessage(ctx, chatID, messageID, markdown, &params.EditMessage{
	ParseMode: params.ParseModeRichMarkdown,
})
```

Use `ParseModeRichMarkdown` and `ParseModeRichHTML` only with the rich message
methods. The legacy `ParseModeMarkdown` and `ParseModeHTML` modes produce normal
message text plus entities and are intentionally rejected by these methods.

Stream a temporary preview while generating content:

```go
draftID := client.RandomID()

for partial := range generatedContent {
	err := client.SendRichMessageDraft(
		ctx,
		chatID,
		draftID,
		partial,
	)
	if err != nil {
		return err
	}
}

_, err := client.SendRichMessage(
	ctx,
	chatID,
	complete,
	&params.SendMessage{ParseMode: params.ParseModeRichMarkdown},
)
```

For one generation, keep the same non-zero `draftID`. Drafts are private-chat
only, last about 30 seconds, and never become persistent automatically. Send the
complete result with `SendRichMessage`.

## Generation Rules

- Prefer Rich Markdown for model-generated content.
- Generate one complete document, not a JSON representation of RichBlock.
- Do not escape content using MarkdownV2 rules.
- Use HTML inside Rich Markdown only for features without Markdown syntax.
- Keep media as separate blocks with public HTTP or HTTPS URLs.
- Do not use Rich Message syntax in media captions.
- Keep table cells to inline formatting.
- Treat formula content as raw LaTeX.
- Preserve the full accumulated document in each streamed draft update.
- Update drafts in meaningful chunks instead of sending every token.
- Never invent custom emoji IDs, user IDs, URLs, timestamps, or media URLs.

## Limits

- 32768 UTF-8 characters, including formula source and custom emoji alt text.
- 500 total blocks, including nested blocks and list/table/detail children.
- 16 levels of nested formatting and blocks.
- 50 total media attachments.
- 20 table columns.

## Rich Markdown

Rich Markdown follows GitHub Flavored Markdown where possible and supports
Telegram's rich HTML tags inline.

### Inline Formatting

```markdown
**bold**
__bold__
*italic*
_italic_
~~strikethrough~~
`inline code`
==marked text==
||spoiler||
[URL](https://example.com)
[email](mailto:user@example.com)
[phone](tel:+123456789)
[user](tg://user?id=123456789)
![🙂](tg://emoji?id=5368324170671202286)
![22:45 tomorrow](tg://time?unix=1647531900&format=wDT)
$x^2 + y^2$
```

Plain URLs, email addresses, mentions, hashtags, cashtags, bot commands, phone
numbers, and bank card numbers are detected automatically.

Use HTML for underline, subscript, and superscript:

```html
<u>underlined</u>
<ins>underlined</ins>
<sub>subscript</sub>
<sup>superscript</sup>
```

### Block Formatting

Headings:

```markdown
# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6
```

Code, divider, lists, tasks, and quotation:

````markdown
```go
fmt.Println("hello")
```

---

- unordered
* unordered
+ unordered

1. ordered
2. ordered

- [ ] pending task
- [x] completed task

> Block quotation
>
> Continued quotation
````

Table:

```markdown
| Metric | Value |
|:-------|------:|
| Speed  | **42 ms** |
| Status | ==Ready== |
```

Footnotes and references:

```markdown
Text with a reference[^note].

[^note]: Referenced text with *inline formatting*.
```

Block formulas:

````markdown
$$E = mc^2$$

```math
E = mc^2
```
````

### Media Blocks

Media must appear as a separate block. Only HTTP and HTTPS URLs are accepted.
Telegram determines the media type from the MIME type and URL. The optional
title becomes the caption.

```markdown
![](https://example.com/photo.jpg)
![](https://example.com/video.mp4 "Video caption")
![](https://example.com/audio.mp3 "Audio caption")
![](https://example.com/voice.ogg "Voice note caption")
![](https://example.com/animation.gif "Animation caption")
```

### Markdown With Rich HTML

Use supported HTML for blocks without native Markdown syntax:

```html
<a name="chapter-1"></a>
<aside>Pull quote<cite>The Author</cite></aside>
<details open>
  <summary>Summary with **Markdown**</summary>

  ## Nested heading
  - Nested list item
</details>
<tg-map lat="41.9" long="12.5" zoom="14"/>
```

Collage and slideshow containers can contain Markdown media blocks:

```html
<tg-collage>

![](https://example.com/photo.jpg)
![](https://example.com/video.mp4)

</tg-collage>

<tg-slideshow>

![](https://example.com/photo.jpg)
![](https://example.com/video.mp4)

</tg-slideshow>
```

## Rich HTML

Use only supported tags. Prefer Rich Markdown unless HTML provides a clearer or
more reliable representation.

Inline tags:

```html
<b>bold</b><strong>bold</strong>
<i>italic</i><em>italic</em>
<u>underline</u><ins>underline</ins>
<s>strike</s><strike>strike</strike><del>strike</del>
<code>inline code</code>
<mark>marked</mark>
<sub>subscript</sub><sup>superscript</sup>
<tg-spoiler>spoiler</tg-spoiler>
<a href="https://example.com">URL</a>
<a href="mailto:user@example.com">email</a>
<a href="tel:+123456789">phone</a>
<a href="tg://user?id=123456789">user</a>
<a href="#chapter-1">anchor link</a>
<a name="chapter-1"></a>
<tg-reference name="note-1">Referenced text</tg-reference>
<tg-emoji emoji-id="5368324170671202286">🙂</tg-emoji>
<img src="tg://emoji?id=5368324170671202286" alt="🙂"/>
<tg-time unix="1647531900" format="wDT">22:45 tomorrow</tg-time>
<tg-math>x^2 + y^2</tg-math>
```

Structural tags:

```html
<h1>Heading</h1>
<p>Paragraph</p>
<pre><code class="language-go">fmt.Println("hello")</code></pre>
<footer>Footer</footer>
<hr/>
<ul><li>Item</li></ul>
<ol start="3" type="a" reversed><li value="7">Item</li></ol>
<blockquote>Quotation<cite>The Author</cite></blockquote>
<aside>Pull quotation<cite>The Author</cite></aside>
<details open><summary>Title</summary>Rich content</details>
<tg-math-block>E = mc^2</tg-math-block>
<tg-map lat="41.9" long="12.5" zoom="14"/>
```

Media and composition:

```html
<img src="https://example.com/photo.jpg"/>
<video src="https://example.com/video.mp4"></video>
<audio src="https://example.com/audio.mp3"></audio>

<figure>
  <img src="https://example.com/photo.jpg" tg-spoiler/>
  <figcaption>Caption<cite>Credit</cite></figcaption>
</figure>

<tg-collage>
  <img src="https://example.com/photo.jpg"/>
  <video src="https://example.com/video.mp4"/>
  <figcaption>Collage caption</figcaption>
</tg-collage>

<tg-slideshow>
  <img src="https://example.com/photo.jpg"/>
  <video src="https://example.com/video.mp4"/>
  <figcaption>Slideshow caption</figcaption>
</tg-slideshow>
```

Tables:

```html
<table bordered striped>
  <caption>Table caption</caption>
  <tr>
    <th>Metric</th>
    <th>Value</th>
  </tr>
  <tr>
    <td align="left">Status</td>
    <td align="right">Ready</td>
  </tr>
</table>
```

Table cells support `colspan`, `rowspan`, `align` (`left`, `center`, `right`),
and `valign` (`top`, `middle`, `bottom`).

Supported named HTML entities are `&lt;`, `&gt;`, `&amp;`, `&quot;`, `&apos;`,
`&nbsp;`, `&hellip;`, `&mdash;`, `&ndash;`, `&lsquo;`, `&rsquo;`, `&ldquo;`,
and `&rdquo;`. Numerical HTML entities are also supported.

## Streaming Content

`SendRichMessageDraft` is a convenience wrapper around `messages.setTyping`
with `tg.InputSendMessageRichMessageDraftAction`. Telegram treats repeated
actions with the same draft ID as revisions of one temporary preview.

When streaming model output:

1. Generate one non-zero draft ID.
2. Accumulate generated content locally.
3. Send valid partial Rich Markdown or HTML at sensible intervals.
4. Reuse the same draft ID for every update.
5. Send the complete document with `SendRichMessage`.

A draft update replaces the preview; it is not a token delta. Do not send only
the newly generated suffix unless the intended preview is only that suffix.

`RichBlockThinking` and its `<tg-thinking>` representation are draft-only. Do
not include thinking blocks in the final persistent message.
