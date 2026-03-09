package model

type User struct {
	BaseModel
	Username string `gorm:"column:username;type:varchar(64);uniqueIndex;not null" json:"username"`
	Password string `gorm:"column:password;type:varchar(255);not null" json:"-"`
	Nickname string `gorm:"column:nickname;type:varchar(64)" json:"nickname"`
	Email    string `gorm:"column:email;type:varchar(128)" json:"email"`
	Status   int    `gorm:"column:status;type:tinyint;default:1;not null" json:"status"`
	Roles    []Role `gorm:"many2many:user_roles;joinForeignKey:UserID;JoinReferences:RoleID" json:"roles,omitempty"`
}

func (User) TableName() string { return "users" }

