import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const scripts = fileURLToPath(new URL('../scripts/', import.meta.url))

function run(cwd, command, args, env = {}, expectedStatus = 0) {
  const result = spawnSync(command, args, {
    cwd,
    encoding: 'utf8',
    env: { ...process.env, ...env },
  })
  assert.equal(result.status, expectedStatus, result.stdout + result.stderr)
  return result.stdout.trim()
}

async function fixture(t) {
  const root = await mkdtemp(join(tmpdir(), 'artalk-release-test-'))
  t.after(() => rm(root, { recursive: true, force: true }))
  const repo = join(root, 'repo')
  const remote = join(root, 'remote.git')
  await mkdir(repo)
  run(root, 'git', ['init', '--bare', remote])
  run(repo, 'git', ['init', '--initial-branch=master'])
  run(repo, 'git', ['config', 'user.name', 'Release test'])
  run(repo, 'git', ['config', 'user.email', 'release@example.com'])
  run(repo, 'git', ['config', 'commit.gpgSign', 'false'])
  run(repo, 'git', ['remote', 'add', 'origin', remote])
  for (const directory of ['internal/config', 'ui/artalk', 'docs/docs/code', '.ci/scripts']) {
    await mkdir(join(repo, directory), { recursive: true })
  }
  for (const script of ['ensure-release-tag.sh', 'tag-merged-release.sh']) {
    await writeFile(join(repo, '.ci/scripts', script), await readFile(join(scripts, script)))
  }
  await writeFile(
    join(repo, 'internal/config/version.go'),
    'package config\nconst Version = "v1.0.0"\n',
  )
  await writeFile(join(repo, 'ui/artalk/package.json'), '{\n  "version": "1.0.0"\n}\n')
  await writeFile(join(repo, 'docs/docs/code/ArtalkVersion.json'), '{\n  "latest": "1.0.0"\n}\n')
  await writeFile(join(repo, 'CHANGELOG.md'), '## v1.0.0 (2026-01-01)\n\nOriginal release notes.\n')
  run(repo, 'git', ['add', '.'])
  run(repo, 'git', ['commit', '-m', 'chore: release v1.0.0'])
  run(repo, 'git', ['push', 'origin', 'master'])
  const commit = run(repo, 'git', ['rev-parse', 'HEAD'])
  const env = {
    RELEASE_VERSION: 'v1.0.0',
    RELEASE_COMMIT: commit,
    ARTALK_RELEASE_WRITES: 'true',
    DRY_RUN: 'false',
  }
  const ensure = (overrides = {}, status = 0) =>
    run(repo, 'sh', [join(scripts, 'ensure-release-tag.sh')], { ...env, ...overrides }, status)
  return { repo, remote, commit, ensure, env }
}

test('tag creation can be retried with an existing or missing local tag', async (t) => {
  const { repo, remote, commit, ensure } = await fixture(t)
  ensure()
  const tag = run(remote, 'git', ['rev-parse', 'refs/tags/v1.0.0'])
  assert.equal(run(remote, 'git', ['rev-parse', 'refs/tags/v1.0.0^{}']), commit)
  ensure()
  run(repo, 'git', ['tag', '-d', 'v1.0.0'])
  ensure({ RELEASE_COMMIT: '' })
  assert.equal(run(remote, 'git', ['rev-parse', 'refs/tags/v1.0.0']), tag)
})

test('retries refuse to move an existing remote tag to another commit', async (t) => {
  const { repo, remote, commit, ensure } = await fixture(t)
  ensure()
  run(repo, 'git', ['commit', '--allow-empty', '-m', 'fix: later changes'])
  const later = run(repo, 'git', ['rev-parse', 'HEAD'])
  ensure({ RELEASE_COMMIT: later }, 1)
  assert.equal(run(remote, 'git', ['rev-parse', 'refs/tags/v1.0.0^{}']), commit)
})

test('missing tags require an explicit release commit and matching version files', async (t) => {
  const { repo, remote, ensure } = await fixture(t)
  ensure({ RELEASE_COMMIT: '' }, 1)
  ensure({ RELEASE_VERSION: 'v01.0.0' }, 1)
  ensure({ RELEASE_VERSION: 'v2.0.0' }, 1)
  await writeFile(join(repo, 'docs/docs/code/ArtalkVersion.json'), '{\n  "latest": "2.0.0"\n}\n')
  run(repo, 'git', ['add', '.'])
  run(repo, 'git', ['commit', '-m', 'chore: mismatched docs version'])
  ensure({ RELEASE_COMMIT: 'HEAD' }, 1)
  assert.equal(run(remote, 'git', ['tag', '--list']), '')
})

test('dry runs do not create tags', async (t) => {
  const { repo, remote, ensure } = await fixture(t)
  ensure({ DRY_RUN: 'true' })
  ensure({ ARTALK_RELEASE_WRITES: 'false' })
  assert.equal(run(repo, 'git', ['tag', '--list']), '')
  assert.equal(run(remote, 'git', ['tag', '--list']), '')
})

test('merged release retries use the merge commit even when HEAD has advanced', async (t) => {
  const { repo, remote, commit, env } = await fixture(t)
  run(repo, 'git', ['commit', '--allow-empty', '-m', 'fix: another commit'])
  const mergedEnv = {
    ...env,
    CNB_PULL_REQUEST_BRANCH: 'release/v1.0.0',
    CNB_PULL_REQUEST_MERGE_SHA: commit,
  }
  for (let attempt = 0; attempt < 2; attempt++) {
    run(repo, 'sh', [join(scripts, 'tag-merged-release.sh')], mergedEnv)
  }
  assert.equal(run(remote, 'git', ['rev-parse', 'refs/tags/v1.0.0^{}']), commit)
  run(
    repo,
    'sh',
    [join(scripts, 'tag-merged-release.sh')],
    { ...env, CNB_PULL_REQUEST_BRANCH: 'fix/example' },
    78,
  )
})

test('release notes come from the original tag when newer releases exist', async (t) => {
  const { repo, ensure } = await fixture(t)
  ensure()
  await writeFile(join(repo, 'CHANGELOG.md'), '## v2.0.0 (2026-02-01)\n\nLater release notes.\n')
  run(repo, 'git', ['add', '.'])
  run(repo, 'git', ['commit', '-m', 'chore: release v2.0.0'])
  run(repo, process.execPath, [join(scripts, 'release-notes.mjs'), 'v1.0.0'])
  assert.equal(
    await readFile(join(repo, 'local/release-notes.md'), 'utf8'),
    '## v1.0.0 (2026-01-01)\n\nOriginal release notes.\n',
  )
})
