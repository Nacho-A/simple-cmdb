package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
)

type AssetUpsertReq struct {
	ServiceName   string            `json:"service_name"`
	PrivateIP     string            `json:"private_ip"`
	PublicIP      string            `json:"public_ip"`
	Labels        map[string]string `json:"labels"`
	Tags          string            `json:"tags"`
	Owner         string            `json:"owner"`
	CloudProvider string            `json:"cloud_provider"`
	Region        string            `json:"region"`
	InstanceType  string            `json:"instance_type"`
	Status        string            `json:"status"`
	Remark        string            `json:"remark"`
}

type AssetBatchDeleteReq struct {
	IDs []uint `json:"ids" binding:"required"`
}

func (h *Handler) AssetList(c *gin.Context) {
	page, pageSize, offset, limit := utils.GetPage(c)
	q := strings.TrimSpace(c.Query("q"))
	serviceName := strings.TrimSpace(c.Query("service_name"))
	ip := strings.TrimSpace(c.Query("ip"))
	owner := strings.TrimSpace(c.Query("owner"))
	tags := strings.TrimSpace(c.Query("tags"))
	label := strings.TrimSpace(c.Query("label"))

	dbq := h.DB.Model(&model.CMDBAsset{})
	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where(
			"service_name like ? or private_ip like ? or owner like ? or tags like ? or cast(labels as char) like ?",
			like, like, like, like, like,
		)
	}
	if serviceName != "" {
		like := "%" + serviceName + "%"
		dbq = dbq.Where("service_name like ?", like)
	}
	if ip != "" {
		like := "%" + ip + "%"
		dbq = dbq.Where("private_ip like ? or public_ip like ?", like, like)
	}
	if owner != "" {
		like := "%" + owner + "%"
		dbq = dbq.Where("owner like ?", like)
	}
	if tags != "" {
		like := "%" + tags + "%"
		dbq = dbq.Where("tags like ?", like)
	}
	if label != "" {
		like := "%" + label + "%"
		dbq = dbq.Where("cast(labels as char) like ?", like)
	}

	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	var items []model.CMDBAsset
	if err := dbq.Order("id desc").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	utils.OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) AssetGet(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var a model.CMDBAsset
	if err := h.DB.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 500, "资产不存在")
			return
		}
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OK(c, a)
}

func (h *Handler) AssetCreate(c *gin.Context) {
	var req AssetUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}
	req.ServiceName = strings.TrimSpace(req.ServiceName)
	if req.ServiceName == "" {
		utils.Fail(c, 500, "service_name必填")
		return
	}

	labels := datatypes.JSONMap{}
	for k, v := range req.Labels {
		if strings.TrimSpace(k) == "" {
			continue
		}
		labels[k] = v
	}

	a := model.CMDBAsset{
		ServiceName:   req.ServiceName,
		PrivateIP:     req.PrivateIP,
		PublicIP:      req.PublicIP,
		Labels:        labels,
		Tags:          req.Tags,
		Owner:         req.Owner,
		CloudProvider: req.CloudProvider,
		Region:        req.Region,
		InstanceType:  req.InstanceType,
		Status:        req.Status,
		Remark:        req.Remark,
	}
	if err := h.DB.Create(&a).Error; err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}
	utils.OK(c, gin.H{"id": a.ID})
}

func (h *Handler) AssetUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	var req AssetUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}
	req.ServiceName = strings.TrimSpace(req.ServiceName)
	if req.ServiceName == "" {
		utils.Fail(c, 500, "service_name必填")
		return
	}

	labels := datatypes.JSONMap{}
	for k, v := range req.Labels {
		if strings.TrimSpace(k) == "" {
			continue
		}
		labels[k] = v
	}

	updates := map[string]interface{}{
		"service_name":   req.ServiceName,
		"private_ip":     req.PrivateIP,
		"public_ip":      req.PublicIP,
		"labels":         labels,
		"tags":           req.Tags,
		"owner":          req.Owner,
		"cloud_provider": req.CloudProvider,
		"region":         req.Region,
		"instance_type":  req.InstanceType,
		"status":         req.Status,
		"remark":         req.Remark,
	}
	if err := h.DB.Model(&model.CMDBAsset{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.Fail(c, 500, "更新失败")
		return
	}
	utils.OK(c, gin.H{})
}

func (h *Handler) AssetDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	if err := h.DB.Delete(&model.CMDBAsset{}, id).Error; err != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	utils.OK(c, gin.H{})
}

func (h *Handler) AssetBatchDelete(c *gin.Context) {
	var req AssetBatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.OK(c, gin.H{})
		return
	}
	if err := h.DB.Where("id in ?", req.IDs).Delete(&model.CMDBAsset{}).Error; err != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	utils.OK(c, gin.H{})
}

func (h *Handler) AssetExportExcel(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	serviceName := strings.TrimSpace(c.Query("service_name"))
	ip := strings.TrimSpace(c.Query("ip"))
	owner := strings.TrimSpace(c.Query("owner"))
	tags := strings.TrimSpace(c.Query("tags"))
	label := strings.TrimSpace(c.Query("label"))
	dbq := h.DB.Model(&model.CMDBAsset{})
	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where(
			"service_name like ? or private_ip like ? or owner like ? or tags like ? or cast(labels as char) like ?",
			like, like, like, like, like,
		)
	}
	if serviceName != "" {
		like := "%" + serviceName + "%"
		dbq = dbq.Where("service_name like ?", like)
	}
	if ip != "" {
		like := "%" + ip + "%"
		dbq = dbq.Where("private_ip like ? or public_ip like ?", like, like)
	}
	if owner != "" {
		like := "%" + owner + "%"
		dbq = dbq.Where("owner like ?", like)
	}
	if tags != "" {
		like := "%" + tags + "%"
		dbq = dbq.Where("tags like ?", like)
	}
	if label != "" {
		like := "%" + label + "%"
		dbq = dbq.Where("cast(labels as char) like ?", like)
	}
	var items []model.CMDBAsset
	if err := dbq.Order("id desc").Limit(5000).Find(&items).Error; err != nil {
		utils.Fail(c, 500, "导出失败")
		return
	}

	f := excelize.NewFile()
	sheets := f.GetSheetList()
	sheetName := sheets[0]
	if len(sheets) == 0 {
		_, _ = f.NewSheet("Sheet1")
		sheetName = "Sheet1"
	}
	headers := []string{
		"ID", "服务名称", "私网IP", "公网IP", "标签(labels)", "tags", "负责人", "云供应商", "地域", "实例规格", "状态", "备注", "创建时间", "更新时间",
	}
	for i, h1 := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, h1)
	}
	for r, a := range items {
		row := r + 2
		labelsStr := ""
		if a.Labels != nil {
			if b, err := json.Marshal(a.Labels); err == nil {
				labelsStr = string(b)
			}
		}
		values := []interface{}{
			a.ID, a.ServiceName, a.PrivateIP, a.PublicIP, labelsStr, a.Tags, a.Owner, a.CloudProvider, a.Region, a.InstanceType, a.Status, a.Remark,
			a.CreatedAt.Format(time.RFC3339), a.UpdatedAt.Format(time.RFC3339),
		}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			_ = f.SetCellValue(sheetName, cell, v)
		}
	}

	var body []byte
	if tmpFile, err := os.CreateTemp("", "cmdb_export_*.xlsx"); err != nil {
		utils.Fail(c, 500, "导出失败")
		return
	} else {
		tmpPath := tmpFile.Name()
		_ = tmpFile.Close()
		defer os.Remove(tmpPath)
		if err := f.SaveAs(tmpPath); err != nil {
			utils.Fail(c, 500, "导出失败")
			return
		}
		body, _ = os.ReadFile(tmpPath)
	}
	if len(body) == 0 {
		utils.Fail(c, 500, "导出失败")
		return
	}

	filename := fmt.Sprintf("cmdb_assets_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(len(body)))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", body)
}
