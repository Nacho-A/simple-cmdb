package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cursor-cmdb-backend/middleware"
	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
)

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserInfoResp struct {
	ID       uint     `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Email    string   `json:"email"`
	Status   int      `json:"status"`
	Roles    []string `json:"roles"`
}

type MenuNode struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Component string     `json:"component"`
	Icon      string     `json:"icon"`
	ParentID  *uint      `json:"parent_id"`
	Order     int        `json:"order"`
	Hidden    bool       `json:"hidden"`
	Children  []MenuNode `json:"children,omitempty"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}

	var u model.User
	if err := h.DB.Preload("Roles").Where("username = ?", req.Username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 401, "用户名或密码错误")
			return
		}
		utils.Fail(c, 500, "服务器错误")
		return
	}

	if u.Status != 1 {
		utils.Fail(c, 403, "用户已被禁用")
		return
	}

	if !utils.CheckPassword(u.Password, req.Password) {
		utils.Fail(c, 401, "用户名或密码错误")
		return
	}

	roles := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}

	token, err := middleware.SignToken(h.Cfg.JWT.Secret, h.Cfg.JWT.Issuer, h.Cfg.JWT.Audience, h.Cfg.JWT.ExpireH, u.ID, u.Username, roles)
	if err != nil {
		utils.Fail(c, 500, "生成token失败")
		return
	}

	utils.OK(c, gin.H{
		"token": token,
		"userInfo": UserInfoResp{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			Status:   u.Status,
			Roles:    roles,
		},
	})
}

func (h *Handler) Me(c *gin.Context) {
	uidV, _ := c.Get(middleware.CtxUserID)
	uid, _ := uidV.(uint)

	var u model.User
	if err := h.DB.Preload("Roles").First(&u, uid).Error; err != nil {
		utils.Fail(c, 401, "未登录")
		return
	}

	roleNames := make([]string, 0, len(u.Roles))
	roleIDs := make([]uint, 0, len(u.Roles))
	for _, r := range u.Roles {
		roleNames = append(roleNames, r.Name)
		roleIDs = append(roleIDs, r.ID)
	}

	menus, err := h.getMenusByRoleIDs(roleIDs)
	if err != nil {
		utils.Fail(c, 500, "获取菜单失败")
		return
	}

	utils.OK(c, gin.H{
		"userInfo": UserInfoResp{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			Status:   u.Status,
			Roles:    roleNames,
		},
		"menus": buildMenuTree(menus),
	})
}

func (h *Handler) getMenusByRoleIDs(roleIDs []uint) ([]model.Menu, error) {
	if len(roleIDs) == 0 {
		return []model.Menu{}, nil
	}
	var menus []model.Menu
	err := h.DB.Table("menus").
		Select("distinct menus.*").
		Joins("join role_menus on role_menus.menu_id = menus.id").
		Where("role_menus.role_id in ?", roleIDs).
		Order("menus.`order` asc, menus.id asc").
		Find(&menus).Error
	return menus, err
}

func buildMenuTree(menus []model.Menu) []MenuNode {
	nodes := make([]*MenuNode, 0, len(menus))
	id2node := make(map[uint]*MenuNode, len(menus))

	for _, m := range menus {
		n := &MenuNode{
			ID:        m.ID,
			Name:      m.Name,
			Path:      m.Path,
			Component: m.Component,
			Icon:      m.Icon,
			ParentID:  m.ParentID,
			Order:     m.Order,
			Hidden:    m.Hidden,
			Children:  []MenuNode{},
		}
		nodes = append(nodes, n)
		id2node[n.ID] = n
	}

	roots := make([]*MenuNode, 0)
	for _, n := range nodes {
		if n.ParentID == nil || *n.ParentID == 0 {
			roots = append(roots, n)
			continue
		}
		if p, ok := id2node[*n.ParentID]; ok {
			p.Children = append(p.Children, *n)
		} else {
			roots = append(roots, n)
		}
	}

	out := make([]MenuNode, 0, len(roots))
	for _, r := range roots {
		out = append(out, *r)
	}
	return out
}

