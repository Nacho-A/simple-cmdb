package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
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

	// ==================== 查询条件（优化后的写法，更清晰且防注入） ====================
	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where(
			h.DB.Where("service_name LIKE ?", like).
				Or("private_ip LIKE ?", like).
				Or("owner LIKE ?", like).
				Or("tags LIKE ?", like).
				Or("CAST(labels AS CHAR) LIKE ?", like),
		)
	}
	if serviceName != "" {
		dbq = dbq.Where("service_name LIKE ?", "%"+serviceName+"%")
	}
	if ip != "" {
		like := "%" + ip + "%"
		dbq = dbq.Where("private_ip LIKE ? OR public_ip LIKE ?", like, like)
	}
	if owner != "" {
		dbq = dbq.Where("owner LIKE ?", "%"+owner+"%")
	}
	if tags != "" {
		dbq = dbq.Where("tags LIKE ?", "%"+tags+"%")
	}
	if label != "" {
		dbq = dbq.Where("CAST(labels AS CHAR) LIKE ?", "%"+label+"%")
	}

	// ==================== 查询数据（建议根据实际数据量调整或改成分页导出） ====================
	var items []model.CMDBAsset
	if err := dbq.Order("id desc").Limit(5000).Find(&items).Error; err != nil {
		utils.Fail(c, 500, "查询数据失败")
		return
	}

	// ==================== 生成 Excel（核心：改为内存 Buffer） ====================
	f := excelize.NewFile()
	sheetName := "Assets"               // 推荐使用有意义的 sheet 名
	f.SetSheetName("Sheet1", sheetName) // 把默认的 Sheet1 改名

	// 表头
	headers := []string{
		"ID", "服务名称", "私网IP", "公网IP", "标签(labels)", "tags", "负责人",
		"云供应商", "地域", "实例规格", "状态", "备注", "创建时间", "更新时间",
	}
	for i, h1 := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, h1)
	}

	// 数据行
	for r, a := range items {
		row := r + 2

		labelsStr := ""
		if a.Labels != nil {
			if b, err := json.Marshal(a.Labels); err == nil {
				labelsStr = string(b)
			} else {
				labelsStr = "[序列化失败]"
			}
		}

		values := []interface{}{
			a.ID, a.ServiceName, a.PrivateIP, a.PublicIP, labelsStr, a.Tags, a.Owner,
			a.CloudProvider, a.Region, a.InstanceType, a.Status, a.Remark,
			a.CreatedAt.Format(time.RFC3339), a.UpdatedAt.Format(time.RFC3339),
		}

		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			_ = f.SetCellValue(sheetName, cell, v)
		}
	}

	// ==================== 直接写入内存 Buffer（关键改动） ====================
	buf, err := f.WriteToBuffer()
	if err != nil {
		utils.Fail(c, 500, "生成 Excel 文件失败")
		return
	}
	if buf.Len() == 0 {
		utils.Fail(c, 500, "导出数据为空")
		return
	}

	// ==================== 设置响应头（UTF-8 中文文件名兼容） ====================
	filename := fmt.Sprintf("cmdb_assets_%s.xlsx", time.Now().Format("20060102_150405"))
	encodedFilename := url.QueryEscape(filename)

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename*=UTF-8''`+encodedFilename)
	c.Header("Content-Length", strconv.Itoa(buf.Len()))

	// ==================== 返回文件流 ====================
	c.DataFromReader(
		http.StatusOK,
		int64(buf.Len()),
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		buf,
		nil,
	)
}

type ServiceItem struct {
	ServiceName string   `json:"service_name"`
	PrivateIPs  []string `json:"private_ips"`
	PublicIPs   []string `json:"public_ips"`
	AssetCount  int      `json:"asset_count"`
}

func (h *Handler) ServiceList(c *gin.Context) {
	page, pageSize, offset, limit := utils.GetPage(c)
	serviceName := strings.TrimSpace(c.Query("service_name"))
	privateIP := strings.TrimSpace(c.Query("private_ip"))
	publicIP := strings.TrimSpace(c.Query("public_ip"))

	h.Log.Info("ServiceList query",
		zap.String("service_name", serviceName),
		zap.String("private_ip", privateIP),
		zap.String("public_ip", publicIP),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	dbq := h.DB.Model(&model.CMDBAsset{})
	if serviceName != "" {
		dbq = dbq.Where("service_name LIKE ?", "%"+serviceName+"%")
	}
	if privateIP != "" {
		dbq = dbq.Where("private_ip LIKE ?", "%"+privateIP+"%")
	}
	if publicIP != "" {
		dbq = dbq.Where("public_ip LIKE ?", "%"+publicIP+"%")
	}

	type serviceRow struct {
		ServiceName string
		PrivateIP   string
		PublicIP    string
	}

	var rows []serviceRow
	if err := dbq.Select("service_name, private_ip, public_ip").Find(&rows).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	serviceMap := make(map[string]*ServiceItem)
	for _, row := range rows {
		if row.ServiceName == "" {
			continue
		}
		item, ok := serviceMap[row.ServiceName]
		if !ok {
			item = &ServiceItem{
				ServiceName: row.ServiceName,
				PrivateIPs:  []string{},
				PublicIPs:   []string{},
			}
			serviceMap[row.ServiceName] = item
		}
		item.AssetCount++
		if row.PrivateIP != "" && !contains(item.PrivateIPs, row.PrivateIP) {
			item.PrivateIPs = append(item.PrivateIPs, row.PrivateIP)
		}
		if row.PublicIP != "" && !contains(item.PublicIPs, row.PublicIP) {
			item.PublicIPs = append(item.PublicIPs, row.PublicIP)
		}
	}

	items := make([]ServiceItem, 0, len(serviceMap))
	for _, v := range serviceMap {
		items = append(items, *v)
	}

	total := len(items)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	utils.OK(c, gin.H{
		"items":     items[start:end],
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
