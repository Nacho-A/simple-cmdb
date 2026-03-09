package model

import "gorm.io/datatypes"

type CMDBAsset struct {
	BaseModel
	ServiceName   string            `gorm:"column:service_name;type:varchar(100);not null" json:"service_name"`
	PrivateIP     string            `gorm:"column:private_ip;type:varchar(50)" json:"private_ip"`
	PublicIP      string            `gorm:"column:public_ip;type:varchar(50)" json:"public_ip"`
	Labels        datatypes.JSONMap  `gorm:"column:labels;type:json" json:"labels"`
	Tags          string            `gorm:"column:tags;type:varchar(200)" json:"tags"`
	Owner         string            `gorm:"column:owner;type:varchar(50)" json:"owner"`
	CloudProvider string            `gorm:"column:cloud_provider;type:varchar(50)" json:"cloud_provider"`
	Region        string            `gorm:"column:region;type:varchar(50)" json:"region"`
	InstanceType  string            `gorm:"column:instance_type;type:varchar(50)" json:"instance_type"`
	Status        string            `gorm:"column:status;type:varchar(20)" json:"status"`
	Remark        string            `gorm:"column:remark;type:text" json:"remark"`
}

func (CMDBAsset) TableName() string { return "cmdb_assets" }

