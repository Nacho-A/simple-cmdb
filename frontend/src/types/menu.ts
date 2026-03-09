export interface MenuNode {
  id: number
  name: string
  path: string
  component: string
  icon: string
  parent_id?: number | null
  order: number
  hidden: boolean
  children?: MenuNode[]
}

