import DOMPurify from 'dompurify'

const allowedTags = [
  'a',
  'abbr',
  'article',
  'b',
  'blockquote',
  'br',
  'caption',
  'code',
  'del',
  'details',
  'div',
  'em',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'hr',
  'i',
  'img',
  'ins',
  'kbd',
  'li',
  'main',
  'mark',
  'ol',
  'p',
  'pre',
  'section',
  'span',
  'strike',
  'strong',
  'sub',
  'summary',
  'sup',
  'table',
  'tbody',
  'td',
  'th',
  'thead',
  'tr',
  'u',
  'ul',
]

const globalAttributes = new Set(['title', 'accesskey'])
const tagAttributes: Record<string, Set<string>> = {
  a: new Set(['href', 'name', 'target', 'aria-label', 'rel']),
  img: new Set([
    'src',
    'alt',
    'title',
    'atk-emoticon',
    'aria-label',
    'data-src',
    'class',
    'loading',
  ]),
  code: new Set(['class']),
  span: new Set(['class', 'style']),
}
const allowedAttributes = [
  ...new Set([...globalAttributes, ...Object.values(tagAttributes).flatMap((v) => [...v])]),
]
const urlAttributes = new Set(['href', 'src', 'data-src'])
const allowedSchemes = new Set(['http', 'https', 'mailto'])
const allowedDataImageMimeTypes = new Set(['image/png', 'image/jpeg', 'image/gif', 'image/webp'])

function normalizeUrlValue(value: string): string {
  // Browsers ignore ASCII control characters inside a URL scheme. Remove them
  // before checking so values such as "java\nscript:" cannot bypass the filter.
  return [...value]
    .filter((character) => {
      const codePoint = character.codePointAt(0) || 0
      return codePoint > 0x20 && (codePoint < 0x7f || codePoint > 0x9f)
    })
    .join('')
}

function isAllowedDataImage(value: string): boolean {
  const match = /^data:([^;,]+);base64,/i.exec(value)
  return !!match && allowedDataImageMimeTypes.has(match[1].toLowerCase())
}

function hasAllowedUrl(tag: string, attribute: string, value: string): boolean {
  const normalized = normalizeUrlValue(value)
  const scheme = /^([a-z][a-z\d+.-]*):/i.exec(normalized)?.[1].toLowerCase()

  if (!scheme) return true
  if (scheme !== 'data') return allowedSchemes.has(scheme)

  return tag === 'img' && attribute === 'src' && isAllowedDataImage(normalized)
}

function filterAttributes(root: HTMLTemplateElement): void {
  root.content.querySelectorAll('*').forEach((element) => {
    const tag = element.tagName.toLowerCase()
    const tagAllowedAttributes = tagAttributes[tag]

    for (const attribute of [...element.attributes]) {
      if (!globalAttributes.has(attribute.name) && !tagAllowedAttributes?.has(attribute.name)) {
        element.removeAttribute(attribute.name)
        continue
      }

      if (urlAttributes.has(attribute.name) && !hasAllowedUrl(tag, attribute.name, attribute.value)) {
        element.removeAttribute(attribute.name)
      }
    }

    const className = element.getAttribute('class')
    const allowedClass =
      (tag === 'code' && /^hljs\W+language-(.*)$/.test(className || '')) ||
      (tag === 'span' && /^(hljs-.*)$/.test(className || '')) ||
      (tag === 'img' && /^lazyload$/.test(className || ''))
    if (className && !allowedClass) element.removeAttribute('class')

    const style = element.getAttribute('style')
    if (tag === 'span' && style && !/^color:(\W+)?#[0-9a-f]{3,6};?$/i.test(style)) {
      element.removeAttribute('style')
    }
  })
}

export function sanitize(content: string): string {
  // Keep Artalk's existing HTML allowlist while using a maintained, non-regex
  // parser. The second pass enforces the original per-tag attribute rules.
  const sanitized = DOMPurify.sanitize(content, {
    ALLOWED_TAGS: allowedTags,
    ALLOWED_ATTR: allowedAttributes,
    ALLOW_ARIA_ATTR: false,
    ALLOW_DATA_ATTR: false,
    ALLOW_UNKNOWN_PROTOCOLS: true,
  })
  const template = document.createElement('template')
  template.innerHTML = sanitized
  filterAttributes(template)
  return template.innerHTML
}
