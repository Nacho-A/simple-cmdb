package controller

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/gin-gonic/gin"

	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
)

type CreateAPIKeyReq struct {
	Name  string `json:"name" binding:"required"`
	Scope string `json:"scope" binding:"required,oneof=read write"`
}

type APIKeyResp struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"`
	Scope     string `json:"scope"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) APIKeyList(c *gin.Context) {
	var keys []model.APIKey
	if err := h.DB.Order("id desc").Find(&keys).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	list := make([]APIKeyResp, 0, len(keys))
	for _, k := range keys {
		list = append(list, APIKeyResp{
			ID:        k.ID,
			Name:      k.Name,
			Scope:     k.Scope,
			Status:    k.Status,
			CreatedAt: k.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	utils.OK(c, gin.H{"list": list, "total": len(list)})
}

func (h *Handler) APIKeyCreate(c *gin.Context) {
	var req CreateAPIKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	var exist model.APIKey
	if err := h.DB.Where("name = ?", req.Name).First(&exist).Error; err == nil {
		utils.Fail(c, 400, "名称已存在")
		return
	}

	rawKey := generateAPIKey()
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey := model.APIKey{
		Name:    req.Name,
		KeyHash: keyHash,
		Scope:   req.Scope,
		Status:  1,
	}

	if err := h.DB.Create(&apiKey).Error; err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}

	utils.OK(c, APIKeyResp{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       rawKey,
		Scope:     apiKey.Scope,
		Status:    apiKey.Status,
		CreatedAt: apiKey.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *Handler) APIKeyDelete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.Fail(c, 400, "参数错误")
		return
	}

	result := h.DB.Delete(&model.APIKey{}, id)
	if result.Error != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		utils.Fail(c, 404, "不存在")
		return
	}

	utils.OK(c, nil)
}

func generateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("sk_live_%s", hex.EncodeToString(b))
}

func (h *Handler) ServiceIPs(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		utils.Fail(c, 400, "参数错误")
		return
	}

	var assets []model.CMDBAsset
	if err := h.DB.Where("service_name = ?", serviceName).Find(&assets).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	ips := make([]string, 0)
	for _, a := range assets {
		if a.PrivateIP != "" {
			ips = append(ips, a.PrivateIP)
		}
	}

	utils.OK(c, gin.H{
		"service": serviceName,
		"ips":     ips,
	})
}
