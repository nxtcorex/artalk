import type { Component } from 'vue'
import { storeToRefs } from 'pinia'
import { RouteLocation } from 'vue-router'
import {
  ArrowLeftRight,
  FileText,
  Globe,
  LayoutDashboard,
  MessageSquare,
  Settings,
  Users,
} from 'lucide-vue-next'
import { useUserStore } from '@/stores/user'
import { useNavStore } from '@/stores/nav'

export type PageItem = {
  label: string
  link: string
  hideOnMobile?: boolean
  icon: Component
}

/**
 * Pages for admin
 */
export const AdminPages: Record<string, PageItem> = {
  overview: {
    label: 'overview',
    link: '/overview',
    icon: LayoutDashboard,
  },
  comments: {
    label: 'comment',
    link: '/comments',
    icon: MessageSquare,
  },
  pages: {
    label: 'page',
    link: '/pages',
    icon: FileText,
  },
  users: {
    label: 'user',
    link: '/users',
    icon: Users,
  },
  sites: {
    label: 'site',
    link: '/sites',
    hideOnMobile: true,
    icon: Globe,
  },
  transfer: {
    label: 'transfer',
    link: '/transfer',
    hideOnMobile: true,
    icon: ArrowLeftRight,
  },
  settings: {
    label: 'settings',
    link: '/settings',
    icon: Settings,
  },
}

/**
 * Pages for user
 */
export const UserPages: Record<string, PageItem> = {
  comments: {
    // Only show comments page for user
    ...AdminPages.comments,
  },
}

export interface NavigationStoreProps {
  onGoPage?: (pageName: string) => void
  onGoTab?: (tabName: string) => void
}

export interface SearchStateApi {
  show: boolean
  value: string
  updateValue: (val: string) => void
  showSearch: () => void
  hideSearch: () => void
}

export type NavGroupDef = {
  key: string
  names: string[]
}

export const NavGroupDefs: NavGroupDef[] = [
  { key: 'navGeneral', names: ['overview', 'comments', 'pages', 'users'] },
  { key: 'navSystem', names: ['sites', 'transfer', 'settings'] },
]

/**
 * Navigation menu hook only use for AppNavigation component
 *
 * (do not expose to other components, others should call `useNavStore`)
 */
export const useNavigationMenu = (props: NavigationStoreProps = {}) => {
  const router = useRouter()
  const { is_admin: isAdmin } = storeToRefs(useUserStore())
  const { tabs, curtPage, curtTab, isMobile, isSearchEnabled } = storeToRefs(useNavStore())

  /**
   * Pages for current user
   */
  const pages = computed((): Record<string, PageItem> => (isAdmin.value ? AdminPages : UserPages))

  /**
   * State for search input
   */
  const searchState = reactive<SearchStateApi>({
    show: false,
    value: '',
    updateValue(val: string) {
      searchState.value = val
    },
    showSearch() {
      searchState.show = true
    },
    hideSearch() {
      searchState.show = false
    },
  })

  /**
   * Navigate to page
   */
  const goPage = (pageName: string) => {
    props.onGoPage?.(pageName)
    router.replace(pages.value[pageName].link)
  }

  /**
   * Navigate to tab
   */
  const goTab = (tabName: string) => {
    props.onGoTab?.(tabName)
    curtTab.value = tabName
  }

  /**
   * Sync current page name from route
   */
  const syncCurtPage = (to?: RouteLocation) => {
    const name = String((to || useRoute()).name ?? '').replace(/^\//, '')
    curtPage.value = name.split('/')[0]
  }

  onMounted(() => {
    syncCurtPage()
  })

  router.afterEach((to, from, failure) => {
    syncCurtPage(to)
    searchState.value = ''
    searchState.hideSearch()
  })

  return reactive({
    pages,
    tabs,
    curtPage,
    curtTab,
    goPage,
    goTab,
    isMobile,
    isSearchEnabled,
    searchState,
    showSearch: searchState.showSearch,
    hideSearch: searchState.hideSearch,
  })
}
