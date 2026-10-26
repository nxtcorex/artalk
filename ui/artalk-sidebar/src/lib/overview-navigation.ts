const pageLinks: Record<string, string> = {
  overview: '/overview',
  comments: '/comments',
  pages: '/pages',
  users: '/users',
  sites: '/sites',
  transfer: '/transfer',
  settings: '/settings',
}

export function initialPage(isAdmin: boolean, fromSidebar: boolean, view: string) {
  if (!isAdmin) return '/comments'
  return pageLinks[view] || (fromSidebar ? '/comments' : '/overview')
}

export function commentsTab(value: unknown, isAdmin: boolean) {
  const allowed = isAdmin
    ? ['all', 'pending', 'personal_all']
    : ['all', 'mentions', 'mine', 'pending']
  return typeof value === 'string' && allowed.includes(value) ? value : 'all'
}

export function overviewURL(value: string, commentID?: number) {
  try {
    const url = new URL(value, window.location.href)
    if (!['http:', 'https:'].includes(url.protocol) || !value) return undefined
    if (commentID) url.hash = `atk-comment-${commentID}`
    return url.href
  } catch {
    return undefined
  }
}
