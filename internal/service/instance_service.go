package service

import (
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/proxy"

	"gorm.io/gorm"
)

// InstanceService 实例管理服务
type InstanceService struct {
	db   *gorm.DB
	pool *proxy.InstancePool
}

// NewInstanceService 创建实例服务
func NewInstanceService(db *gorm.DB, pool *proxy.InstancePool) *InstanceService {
	return &InstanceService{
		db:   db,
		pool: pool,
	}
}

// SetPool 设置实例池引用
func (s *InstanceService) SetPool(pool *proxy.InstancePool) {
	s.pool = pool
}

// LoadInstances 从数据库加载实例配置
func (s *InstanceService) LoadInstances() ([]*proxy.Instance, error) {
	var dbInstances []models.ComfyInstance
	if err := s.db.Where("status = ? AND pool = ?", "active", "shared").
		Order("priority DESC, id ASC").
		Find(&dbInstances).Error; err != nil {
		return nil, err
	}

	instances := make([]*proxy.Instance, 0, len(dbInstances))
	for _, dbInst := range dbInstances {
		instances = append(instances, &proxy.Instance{
			ID:        dbInst.Name,
			URL:       dbInst.URL,
			GPUID:     dbInst.GPUID,
			QueueSize: 0,
			Status:    "healthy",
		})
	}

	return instances, nil
}

// ReloadPool 重新加载实例池
func (s *InstanceService) ReloadPool() error {
	instances, err := s.LoadInstances()
	if err != nil {
		return err
	}

	// 更新实例池
	s.pool.ReplaceInstances(instances)

	return nil
}

// AddInstance 添加实例
func (s *InstanceService) AddInstance(inst *models.ComfyInstance) error {
	if err := s.db.Create(inst).Error; err != nil {
		return err
	}

	// 如果是 shared 池且状态为 active，立即加载到实例池
	if inst.Pool == "shared" && inst.Status == "active" {
		s.ReloadPool()
	}

	return nil
}

// UpdateInstance 更新实例
func (s *InstanceService) UpdateInstance(inst *models.ComfyInstance) error {
	if err := s.db.Save(inst).Error; err != nil {
		return err
	}

	// 重新加载实例池
	s.ReloadPool()

	return nil
}

// DeleteInstance 删除实例
func (s *InstanceService) DeleteInstance(id uint) error {
	if err := s.db.Delete(&models.ComfyInstance{}, id).Error; err != nil {
		return err
	}

	// 重新加载实例池
	s.ReloadPool()

	return nil
}

// GetAllInstances 获取所有实例
func (s *InstanceService) GetAllInstances() ([]models.ComfyInstance, error) {
	var instances []models.ComfyInstance
	if err := s.db.Order("pool ASC, priority DESC, id ASC").Find(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

// GetInstanceByID 根据 ID 获取实例
func (s *InstanceService) GetInstanceByID(id uint) (*models.ComfyInstance, error) {
	var instance models.ComfyInstance
	if err := s.db.First(&instance, id).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}
