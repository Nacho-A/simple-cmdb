package controller

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
)

type UserCreateReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Status   *int   `json:"status"`
}

type UserUpdateReq struct {
	Password *string `json:"password"`
	Nickname *string `json:"nickname"`
	Email    *string `json:"email"`
	Status   *int    `json:"status"`
}

type UserBindRolesReq struct {
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

func (h *Handler) UserList(c *gin.Context) {
	page, pageSize, offset, limit := utils.GetPage(c)
	q := strings.TrimSpace(c.Query("q"))

	var total int64
	dbq := h.DB.Model(&model.User{})
	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("username like ? or nickname like ? or email like ?", like, like, like)
	}
	if err := dbq.Count(&total).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	var users []model.User
	if err := dbq.Preload("Roles").Order("id desc").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	items := make([]UserInfoResp, 0, len(users))
	for _, u := range users {
		rs := make([]string, 0, len(u.Roles))
		for _, r := range u.Roles {
			rs = append(rs, r.Name)
		}
		items = append(items, UserInfoResp{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			Status:   u.Status,
			Roles:    rs,
		})
	}

	utils.OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) UserCreate(c *gin.Context) {
	var req UserCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}

	var cnt int64
	if err := h.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&cnt).Error; err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}
	if cnt > 0 {
		utils.Fail(c, 500, "用户名已存在")
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	u := model.User{
		Username: req.Username,
		Password: hash,
		Nickname: req.Nickname,
		Email:    req.Email,
		Status:   status,
	}
	if err := h.DB.Create(&u).Error; err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}

	utils.OK(c, gin.H{"id": u.ID})
}

func (h *Handler) UserUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}

	var req UserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}

	var u model.User
	if err := h.DB.First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 500, "用户不存在")
			return
		}
		utils.Fail(c, 500, "更新失败")
		return
	}

	updates := map[string]interface{}{}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Password != nil && *req.Password != "" {
		hash, err := utils.HashPassword(*req.Password)
		if err != nil {
			utils.Fail(c, 500, "更新失败")
			return
		}
		updates["password"] = hash
	}

	if len(updates) == 0 {
		utils.OK(c, gin.H{})
		return
	}

	if err := h.DB.Model(&model.User{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
		utils.Fail(c, 500, "更新失败")
		return
	}

	utils.OK(c, gin.H{})
}

func (h *Handler) UserDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	if err := h.DB.Delete(&model.User{}, id).Error; err != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	utils.OK(c, gin.H{})
}

func (h *Handler) UserBindRoles(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var req UserBindRolesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}

	var u model.User
	if err := h.DB.Preload("Roles").First(&u, id).Error; err != nil {
		utils.Fail(c, 500, "用户不存在")
		return
	}

	var roles []model.Role
	if err := h.DB.Where("id in ?", req.RoleIDs).Find(&roles).Error; err != nil {
		utils.Fail(c, 500, "绑定失败")
		return
	}
	if len(roles) != len(req.RoleIDs) {
		utils.Fail(c, 500, "存在无效角色")
		return
	}

	if err := h.DB.Model(&u).Association("Roles").Replace(&roles); err != nil {
		utils.Fail(c, 500, "绑定失败")
		return
	}

	utils.OK(c, gin.H{})
}

