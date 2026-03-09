import { defineStore } from 'pinia'

export const useThemeStore = defineStore('theme', {
  state: () => ({
    dark: true,
  }),
  persist: true,
  actions: {
    applyTheme() {
      const html = document.documentElement
      if (this.dark) html.classList.add('dark')
      else html.classList.remove('dark')
    },
    toggle() {
      this.dark = !this.dark
      this.applyTheme()
    },
    setDark(v: boolean) {
      this.dark = v
      this.applyTheme()
    },
  },
})

