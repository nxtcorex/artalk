import { execFileSync } from 'node:child_process'
import { readFile, writeFile } from 'node:fs/promises'

const [version, requestedPrevious = '', ...args] = process.argv.slice(2)
if (!/^v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$/.test(version || '')) {
  throw new Error(`Invalid version: ${version || '<empty>'}`)
}

const outputIndex = args.indexOf('--output')
const output = outputIndex >= 0 ? args[outputIndex + 1] : '.ci-release-changelog.md'
const prepend = args.includes('--prepend')

const tags = execFileSync('git', ['tag', '--list', 'v[0-9]*.[0-9]*.[0-9]*', '--sort=-v:refname'], {
  encoding: 'utf8',
})
  .trim()
  .split('\n')
  .filter(Boolean)

const previous = requestedPrevious || tags.find((tag) => tag !== version)
if (!previous) throw new Error('Unable to resolve the previous release tag')

const log = execFileSync('git', ['log', '--format=%H%x1f%s%x1f%b%x1e', `${previous}..HEAD`], {
  encoding: 'utf8',
})

const groups = new Map([
  ['feat', { title: 'Features', commits: [] }],
  ['fix', { title: 'Bug Fixes', commits: [] }],
  ['perf', { title: 'Performance Improvements', commits: [] }],
  ['refactor', { title: 'Code Refactoring', commits: [] }],
  ['test', { title: 'Testing', commits: [] }],
  ['docs', { title: 'Documentation', commits: [] }],
])
const breaking = []

for (const record of log.split('\x1e')) {
  const [, subject = '', body = ''] = record.trim().split('\x1f')
  const match = subject.match(/^(feat|fix|perf|refactor|test|docs)(?:\(([^)]+)\))?!?:\s+(.+)$/)
  if (match) {
    const [, type, scope, summary] = match
    groups.get(type).commits.push(`* ${scope ? `**${scope}:** ` : ''}${summary}`)
  }

  const breakingMatch = body.match(/(?:BREAKING CHANGE|BREAKING-CHANGE):\s*([^\n]+)/)
  if (breakingMatch) breaking.push(`* ${breakingMatch[1]}`)
}

const lines = [`## ${version} (${new Date().toISOString().slice(0, 10)})`, '']
let hasChanges = false
for (const { title, commits } of groups.values()) {
  if (commits.length === 0) continue
  hasChanges = true
  lines.push(`### ${title}`, '', ...commits, '')
}
if (breaking.length > 0) {
  hasChanges = true
  lines.push('### BREAKING CHANGE', '', ...breaking, '')
}
if (!hasChanges) lines.push('No user-facing changes.', '')

const changelog = `${lines.join('\n').trim()}\n`
await writeFile(output, changelog)

if (prepend) {
  const current = await readFile('CHANGELOG.md', 'utf8')
  await writeFile('CHANGELOG.md', `${changelog}\n${current}`)
}
