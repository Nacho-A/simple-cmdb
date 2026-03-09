export interface CMDBAsset {
  id: number
  service_name: string
  private_ip: string
  public_ip: string
  labels: Record<string, any>
  tags: string
  owner: string
  cloud_provider: string
  region: string
  instance_type: string
  status: string
  remark: string
  created_at: string
  updated_at: string
}

