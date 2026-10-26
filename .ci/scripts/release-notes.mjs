import { execFileSync } from 'node:child_process'
import { mkdir, writeFile } from 'node:fs/promises'

const [version, ref = version] = process.argv.slice(2)
if (!/^v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$/.test(version || '')) {
  throw new Error(`Invalid release version: ${version || '<empty>'}`)
}

// Read the release commit, not today's HEAD or the newest tag. Retrying an old
// release must keep its original changelog, even after newer releases exist.
const changelog = execFileSync('git', ['show', `${ref}:CHANGELOG.md`], { encoding: 'utf8' })
const sections = changelog.split(/(?=^## v\d+\.\d+\.\d+\b)/m)
const notes = sections.find((section) => section.startsWith(`## ${version} (`))
const description =
  notes?.trim() || `## ${version}\n\nSee CHANGELOG.md for changes in this release.`

await mkdir('local', { recursive: true })
await writeFile('local/release-notes.md', `${description}\n`)
