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

type RoleCreateReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type RoleUpdateReq struct {
	Description *string `json:"description"`
}

type RoleSaveMenusReq struct {
	MenuIDs []uint `json:"menu_ids" binding:"required"`
}

func (h *Handler) RoleList(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	var roles []model.Role
	dbq := h.DB.Model(&model.Role{})
	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("name like ? or description like ?", like, like)
	}
	if err := dbq.Order("id asc").Find(&roles).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OK(c, gin.H{"items": roles})
}

func (h *Handler) RoleCreate(c *gin.Context) {
	var req RoleCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var cnt int64
	if err := h.DB.Model(&model.Role{}).Where("name = ?", req.Name).Count(&cnt).Error; err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}
	if cnt > 0 {
		utils.Fail(c, 500, "角色名已存在")
		return
	}
	r := model.Role{Name: req.Name, Description: req.Description}
	if err := h.DB.Create(&r).Error; err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}
	utils.OK(c, gin.H{"id": r.ID})
}

func (h *Handler) RoleUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var req RoleUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}
	updates := map[string]interface{}{}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) == 0 {
		utils.OK(c, gin.H{})
		return
	}
	if err := h.DB.Model(&model.Role{}).Where("id = ?", uint(id)).Updates(updates).Error; err != nil {
		utils.Fail(c, 500, "更新失败")
		return
	}
	utils.OK(c, gin.H{})
}

func (h *Handler) RoleDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	if err := h.DB.Delete(&model.Role{}, uint(id)).Error; err != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	utils.OK(c, gin.H{})
}

func (h *Handler) RoleGetMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var role model.Role
	if err := h.DB.Preload("Menus").First(&role, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 500, "角色不存在")
			return
		}
		utils.Fail(c, 500, "查询失败")
		return
	}
	menuIDs := make([]uint, 0, len(role.Menus))
	for _, m := range role.Menus {
		menuIDs = append(menuIDs, m.ID)
	}
	utils.OK(c, gin.H{"menu_ids": menuIDs})
}

func (h *Handler) RoleSaveMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var req RoleSaveMenusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}

	var role model.Role
	if err := h.DB.First(&role, uint(id)).Error; err != nil {
		utils.Fail(c, 500, "角色不存在")
		return
	}

	var menus []model.Menu
	if len(req.MenuIDs) > 0 {
		if err := h.DB.Where("id in ?", req.MenuIDs).Find(&menus).Error; err != nil {
			utils.Fail(c, 500, "保存失败")
			return
		}
		if len(menus) != len(req.MenuIDs) {
			utils.Fail(c, 500, "存在无效菜单")
			return
		}
	}

	if err := h.DB.Model(&role).Association("Menus").Replace(&menus); err != nil {
		utils.Fail(c, 500, "保存失败")
		return
	}
	utils.OK(c, gin.H{})
}
