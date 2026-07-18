import { readFile, writeFile } from 'node:fs/promises'

const version = process.argv[2]
if (!/^v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$/.test(version || '')) {
  throw new Error(`Invalid version: ${version || '<empty>'}`)
}

const packageVersion = version.slice(1)

async function updateJson(path, mutate) {
  const data = JSON.parse(await readFile(path, 'utf8'))
  mutate(data)
  await writeFile(path, `${JSON.stringify(data, null, 2)}\n`)
}

const goVersionPath = 'internal/config/version.go'
const goVersion = await readFile(goVersionPath, 'utf8')
const updatedGoVersion = goVersion.replace(
  /^const Version = "v\d+\.\d+\.\d+"$/m,
  `const Version = "${version}"`,
)
if (updatedGoVersion === goVersion) {
  throw new Error(`Version constant was not found in ${goVersionPath}`)
}
await writeFile(goVersionPath, updatedGoVersion)

await updateJson('ui/artalk/package.json', (pkg) => {
  pkg.version = packageVersion
})
await updateJson('docs/docs/code/ArtalkVersion.json', (data) => {
  data.latest = packageVersion
})
