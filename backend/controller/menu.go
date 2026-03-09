package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
)

type MenuCreateReq struct {
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Component string `json:"component"`
	Icon      string `json:"icon"`
	ParentID  *uint  `json:"parent_id"`
	Order     int    `json:"order"`
	Hidden    bool   `json:"hidden"`
}

type MenuUpdateReq struct {
	Name      *string `json:"name"`
	Path      *string `json:"path"`
	Component *string `json:"component"`
	Icon      *string `json:"icon"`
	ParentID  *uint   `json:"parent_id"`
	Order     *int    `json:"order"`
	Hidden    *bool   `json:"hidden"`
}

func (h *Handler) MenuList(c *gin.Context) {
	var menus []model.Menu
	if err := h.DB.Order("`order` asc, id asc").Find(&menus).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OK(c, gin.H{"items": buildMenuTree(menus)})
}

func (h *Handler) MenuCreate(c *gin.Context) {
	var req MenuCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}
	m := model.Menu{
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		ParentID:  req.ParentID,
		Order:     req.Order,
		Hidden:    req.Hidden,
	}
	if err := h.DB.Create(&m).Error; err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}
	utils.OK(c, gin.H{"id": m.ID})
}

func (h *Handler) MenuUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var req MenuUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Path != nil {
		updates["path"] = *req.Path
	}
	if req.Component != nil {
		updates["component"] = *req.Component
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.Order != nil {
		updates["order"] = *req.Order
	}
	if req.Hidden != nil {
		updates["hidden"] = *req.Hidden
	}
	if len(updates) == 0 {
		utils.OK(c, gin.H{})
		return
	}
	if err := h.DB.Model(&model.Menu{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.Fail(c, 500, "更新失败")
		return
	}
	utils.OK(c, gin.H{})
}

func (h *Handler) MenuDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var cnt int64
	if err := h.DB.Model(&model.Menu{}).Where("parent_id = ?", id).Count(&cnt).Error; err != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	if cnt > 0 {
		utils.Fail(c, 500, "请先删除子菜单")
		return
	}
	if err := h.DB.Delete(&model.Menu{}, id).Error; err != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	utils.OK(c, gin.H{})
}

