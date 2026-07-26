import YAML from 'yaml'
import { getFlattenNodes, getTree, type OptionNode } from './settings-option'

type EnvValue = string | string[]

export interface EnvRef {
  key: string
  value: EnvValue
}

export class Settings {
  private tree: OptionNode
  private flatten: { [path: string]: OptionNode }
  private customs = shallowRef<YAML.Document.Parsed<YAML.ParsedNode>>()
  private envs = shallowRef<{ [key: string]: string }>({})

  constructor(yamlObj: YAML.Document.Parsed) {
    this.tree = getTree(yamlObj)
    this.flatten = getFlattenNodes(this.tree)
  }

  getTree() {
    return this.tree
  }

  getNode(path: string) {
    return this.flatten[path]
  }

  getCustoms() {
    return this.customs
  }

  setCustoms(yamlStr: string) {
    this.customs.value = YAML.parseDocument(yamlStr)
  }

  setEnvs(envs: string[]) {
    const envsObj: { [key: string]: string } = {}
    envs.forEach((env) => {
      const separatorIndex = env.indexOf('=')
      const key = separatorIndex === -1 ? env : env.slice(0, separatorIndex)
      const value = separatorIndex === -1 ? '' : env.slice(separatorIndex + 1)
      if (!key) return
      envsObj[key] = value
    })
    this.envs.value = envsObj
  }

  getEnv(key: string) {
    return Object.prototype.hasOwnProperty.call(this.envs.value, key) ? this.envs.value[key] : null
  }

  getEnvRefByPath(path: string): EnvRef | null {
    for (const key of this.getEnvKeyCandidates(path)) {
      const value = this.getEnv(key)
      if (value !== null) return { key, value }
    }

    return this.getIndexedEnvRefByPath(path)
  }

  getEnvByPath(path: string) {
    return this.getEnvRefByPath(path)?.value ?? null
  }

  getEnvKeyByPath(path: string) {
    return this.getEnvRefByPath(path)?.key ?? this.getEnvKeyCandidates(path)[0]
  }

  getCustom(path: string) {
    const env = this.getEnvByPath(path)
    if (env !== null) return env

    const value = this.customs.value?.getIn(path.split('.')) as any
    if (value !== undefined) return value

    return this.flatten[path]?.default
  }

  setCustom(path: string, value: any) {
    const pathArr = path.split('.')

    this.makeSureObject(pathArr)

    this.customs.value?.setIn(pathArr, value)
  }

  // @see https://github.com/eemeli/yaml/issues/174#issuecomment-632281283
  private makeSureObject(pathArr: string[]) {
    for (let i = pathArr.length - 1; i >= 1; i--) {
      const parentPath = pathArr.slice(0, -i)

      const parentNode = this.customs.value?.getIn(parentPath)
      if (!parentNode) {
        this.customs.value?.setIn(parentPath, new YAML.YAMLMap())
      }
    }
  }

  private getEnvKeyCandidates(path: string) {
    const segments = path.split('.')
    return unique([
      `ATK_${segments.join('_').toUpperCase()}`,
      `ATK_${segments.map((segment) => segment.toUpperCase().replace(/_/g, '__')).join('_')}`,
    ])
  }

  private getIndexedEnvRefByPath(path: string): EnvRef | null {
    const matches = this.getEnvKeyCandidates(path).flatMap((key) => this.getIndexedEnvs(key))
    if (matches.length === 0) return null

    matches.sort((a, b) => a.index - b.index)

    return {
      key: matches[0].key,
      value: matches.map((match) => match.value),
    }
  }

  private getIndexedEnvs(baseKey: string) {
    const indexedEnvReg = new RegExp(`^${escapeRegExp(baseKey)}_(\\d+)$`)
    return Object.entries(this.envs.value)
      .map(([key, value]) => {
        const match = indexedEnvReg.exec(key)
        if (!match) return null

        return {
          key,
          value,
          index: Number(match[1]),
        }
      })
      .filter((item): item is { key: string; value: string; index: number } => item !== null)
  }
}

// -------------------------------------------------------

export * from './settings-option'

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function unique<T>(items: T[]) {
  return Array.from(new Set(items))
}

// Singleton instance
let instance: Settings

export default {
  init: (yamlObj: YAML.Document.Parsed) => (instance = new Settings(yamlObj)),
  get: () => instance,
}
