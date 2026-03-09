package model

type Menu struct {
	BaseModel
	Name      string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Path      string `gorm:"column:path;type:varchar(255);not null" json:"path"`
	Component string `gorm:"column:component;type:varchar(255)" json:"component"`
	Icon      string `gorm:"column:icon;type:varchar(64)" json:"icon"`
	ParentID  *uint  `gorm:"column:parent_id" json:"parent_id"`
	Order     int    `gorm:"column:order;type:int;default:0;not null" json:"order"`
	Hidden    bool   `gorm:"column:hidden;type:tinyint(1);default:0;not null" json:"hidden"`
}

func (Menu) TableName() string { return "menus" }

