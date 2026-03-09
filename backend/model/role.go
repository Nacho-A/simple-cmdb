package model

type Role struct {
	BaseModel
	Name        string `gorm:"column:name;type:varchar(64);uniqueIndex;not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(255)" json:"description"`
	Menus       []Menu `gorm:"many2many:role_menus;joinForeignKey:RoleID;JoinReferences:MenuID" json:"menus,omitempty"`
}

func (Role) TableName() string { return "roles" }

