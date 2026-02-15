package proxy

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Instance ComfyUI 实例
type Instance struct {
	ID        string    `json:"id"`         // 实例 ID（如 comfyui-1）
	URL       string    `json:"url"`        // 实例 URL（如 http://comfyui-1:8188）
	GPUID     int       `json:"gpu_id"`     // GPU ID
	QueueSize int       `json:"queue_size"` // 当前队列长度
	Status    string    `json:"status"`     // healthy/unhealthy
	LastCheck time.Time `json:"last_check"` // 最后健康检查时间
}

// InstancePool 实例池
type InstancePool struct {
	instances []*Instance
	mu        sync.RWMutex
}

// NewInstancePool 创建实例池
func NewInstancePool(urls []string) *InstancePool {
	pool := &InstancePool{
		instances: make([]*Instance, 0, len(urls)),
	}

	for i, url := range urls {
		pool.instances = append(pool.instances, &Instance{
			ID:        fmt.Sprintf("comfyui-%d", i+1),
			URL:       url,
			GPUID:     i,
			QueueSize: 0,
			Status:    "healthy",
			LastCheck: time.Now(),
		})
	}

	return pool
}

// GetHealthyInstances 获取所有健康的实例
func (p *InstancePool) GetHealthyInstances() []*Instance {
	p.mu.RLock()
	defer p.mu.RUnlock()

	healthy := make([]*Instance, 0)
	for _, inst := range p.instances {
		if inst.Status == "healthy" {
			healthy = append(healthy, inst)
		}
	}
	return healthy
}

// SelectInstance 选择实例（简单轮询）
func (p *InstancePool) SelectInstance() *Instance {
	healthy := p.GetHealthyInstances()
	if len(healthy) == 0 {
		return nil
	}

	// 选择队列最短的实例
	minQueue := healthy[0]
	for _, inst := range healthy {
		if inst.QueueSize < minQueue.QueueSize {
			minQueue = inst
		}
	}

	return minQueue
}

// UpdateQueueSize 更新实例队列长度
func (p *InstancePool) UpdateQueueSize(instanceID string, queueSize int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, inst := range p.instances {
		if inst.ID == instanceID {
			inst.QueueSize = queueSize
			break
		}
	}
}

// UpdateStatus 更新实例状态
func (p *InstancePool) UpdateStatus(instanceID string, status string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, inst := range p.instances {
		if inst.ID == instanceID {
			inst.Status = status
			inst.LastCheck = time.Now()
			break
		}
	}
}

// GetInstance 根据 ID 获取实例
func (p *InstancePool) GetInstance(instanceID string) *Instance {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, inst := range p.instances {
		if inst.ID == instanceID {
			return inst
		}
	}
	return nil
}

// GetAllInstances 获取所有实例
func (p *InstancePool) GetAllInstances() []*Instance {
	p.mu.RLock()
	defer p.mu.RUnlock()

	instances := make([]*Instance, len(p.instances))
	copy(instances, p.instances)
	return instances
}

// HealthCheck 健康检查
func (p *InstancePool) HealthCheck() {
	for _, inst := range p.GetAllInstances() {
		go func(instance *Instance) {
			// 检查实例是否可访问
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(instance.URL + "/system_stats")

			if err != nil || resp.StatusCode != 200 {
				p.UpdateStatus(instance.ID, "unhealthy")
			} else {
				p.UpdateStatus(instance.ID, "healthy")
				resp.Body.Close()
			}
		}(inst)
	}
}

// StartHealthCheck 启动定期健康检查
func (p *InstancePool) StartHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			p.HealthCheck()
		}
	}()
}
