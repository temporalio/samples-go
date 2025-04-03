package queue

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// SampleWorkflowWithResourcePool 使用资源池的示例工作流
func SampleWorkflowWithResourcePool(
	ctx workflow.Context,
	resourceID string,
	processTime time.Duration,
	cancelable bool,
) error {
	currentWorkflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	logger := workflow.GetLogger(ctx)
	logger.Info("workflow start",
		"workflowID", currentWorkflowID,
		"resourceID", resourceID,
		"cancelable", cancelable)

	// 创建资源池
	resourcePool := NewResourcePool(currentWorkflowID, "ResourcePoolDemo")

	// 尝试获取资源，如果指定了可取消，则在等待队列中可以被取消
	releaseFunc, resourceInfo, err := resourcePool.AcquireResource(ctx, resourceID, 30*time.Minute, cancelable)
	if err != nil {
		logger.Error("failed to acquire resource", "Error", err)
		return err
	}

	logger.Info("successfully acquired resource, start processing",
		"resourceID", resourceInfo.GetID(),
		"resourceIndex", resourceInfo.GetIndex(),
		"acquiredTime", resourceInfo.GetAcquiredTime())

	// 模拟处理过程
	logger.Info("start critical operation")
	if processTime <= 0 {
		processTime = 10 * time.Second // 默认处理时间
	}
	_ = workflow.Sleep(ctx, processTime)
	logger.Info("critical operation completed")

	// 释放资源
	err = releaseFunc()
	if err != nil {
		logger.Error("failed to release resource", "Error", err)
		return err
	}

	logger.Info("resource released, workflow completed")
	return nil
}

// CustomResourcePoolInitializer 自定义资源池初始化器示例
type CustomResourcePoolInitializer struct {
	Prefix     string                 // 资源ID前缀
	ExtraProps map[string]interface{} // 额外属性
}

// Initialize 实现自定义资源池初始化
func (ci *CustomResourcePoolInitializer) Initialize(
	ctx workflow.Context,
	resourceID string,
	totalResources int,
) *ResourcePoolState {
	logger := workflow.GetLogger(ctx)
	logger.Info("using custom resource pool initializer",
		"resourceID", resourceID,
		"prefix", ci.Prefix,
		"totalResources", totalResources)

	// 创建资源池状态
	state := &ResourcePoolState{
		ResourcePool:   make([]InternalResource, totalResources),
		AvailableCount: totalResources,
		WaitingQueue:   []ResourceRequest{},
		ActiveRequests: make(map[string]int),
	}

	// 添加自定义属性到元数据
	baseMetadata := map[string]interface{}{
		"initTime": workflow.Now(ctx),
	}

	// 合并额外属性
	for k, v := range ci.ExtraProps {
		baseMetadata[k] = v
	}

	// 初始化资源池
	for i := 0; i < totalResources; i++ {
		// 为每个资源复制基础元数据
		metadata := make(map[string]interface{})
		for k, v := range baseMetadata {
			metadata[k] = v
		}

		// 添加资源特定属性
		metadata["poolIndex"] = i

		// 创建资源
		state.ResourcePool[i] = InternalResource{
			Info: ResourceInfo{
				ResourceID:    fmt.Sprintf("%s:%s-%d", ci.Prefix, resourceID, i),
				Metadata:      metadata,
				ResourceIndex: i,
			},
			Available: true,
		}
	}

	return state
}

// ExpandPool 自定义资源池扩展实现
func (ci *CustomResourcePoolInitializer) ExpandPool(
	ctx workflow.Context,
	state *ResourcePoolState,
	currentSize int,
	newSize int,
	resourceID string,
) int {
	logger := workflow.GetLogger(ctx)
	logger.Info("using custom expander to expand resource pool",
		"from", currentSize,
		"to", newSize)

	// 创建基础元数据
	baseMetadata := map[string]interface{}{
		"expandTime": workflow.Now(ctx),
	}

	// 合并额外属性
	for k, v := range ci.ExtraProps {
		baseMetadata[k] = v
	}

	// 扩展资源池
	for i := currentSize; i < newSize; i++ {
		// 为每个资源复制基础元数据
		metadata := make(map[string]interface{})
		for k, v := range baseMetadata {
			metadata[k] = v
		}

		// 添加资源特定属性
		metadata["poolIndex"] = i
		metadata["expanded"] = true

		// 创建资源
		state.ResourcePool = append(state.ResourcePool, InternalResource{
			Info: ResourceInfo{
				ResourceID:    fmt.Sprintf("%s:%s-%d", ci.Prefix, resourceID, i),
				Metadata:      metadata,
				ResourceIndex: i,
			},
			Available: true,
		})
	}

	// 返回新增的资源数量
	return newSize - currentSize
}

// ShrinkPool 自定义资源池缩减实现
func (ci *CustomResourcePoolInitializer) ShrinkPool(
	ctx workflow.Context,
	state *ResourcePoolState,
	currentSize int,
	newSize int,
) int {
	logger := workflow.GetLogger(ctx)
	logger.Info("using custom shrinker to shrink resource pool",
		"from", currentSize,
		"to", newSize)

	// 优先缩减具有'expanded'标记的资源
	availableReduceCount := 0
	reducedIndices := make([]int, 0, currentSize-newSize)

	// 第一轮：寻找并处理标记为expanded的可用资源
	for i := currentSize - 1; i >= 0 && availableReduceCount < currentSize-newSize; i-- {
		if state.ResourcePool[i].Available {
			metadata := state.ResourcePool[i].Info.Metadata
			if expanded, ok := metadata["expanded"].(bool); ok && expanded {
				// 标记为不可用
				state.ResourcePool[i].Available = false
				availableReduceCount++
				reducedIndices = append(reducedIndices, i)
			}
		}
	}

	// 第二轮：如果还需要缩减更多资源，处理剩余的可用资源
	if availableReduceCount < currentSize-newSize {
		for i := currentSize - 1; i >= 0 && availableReduceCount < currentSize-newSize; i-- {
			// 检查此索引是否已经被处理过
			alreadyProcessed := false
			for _, idx := range reducedIndices {
				if idx == i {
					alreadyProcessed = true
					break
				}
			}

			if !alreadyProcessed && state.ResourcePool[i].Available {
				// 标记为不可用
				state.ResourcePool[i].Available = false
				availableReduceCount++
			}
		}
	}

	// 裁剪资源池大小
	if len(state.ActiveRequests) == 0 {
		// 如果没有活动请求，可以直接缩减资源池
		state.ResourcePool = state.ResourcePool[:newSize]
		logger.Info("resource pool physical size reduced", "new size", len(state.ResourcePool))
	} else {
		logger.Info("resource pool has active requests, will reduce physical size after requests complete")
	}

	// 返回减少的可用资源数
	return availableReduceCount
}

// SampleWorkflowWithCustomResourcePool 使用自定义资源池的示例工作流
func SampleWorkflowWithCustomResourcePool(
	ctx workflow.Context,
	resourceID string,
	processTime time.Duration,
	cancelable bool,
) error {
	currentWorkflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	logger := workflow.GetLogger(ctx)
	logger.Info("custom resource pool workflow start",
		"workflowID", currentWorkflowID,
		"resourceID", resourceID,
		"cancelable", cancelable)

	// 创建自定义初始化器
	customInitializer := &CustomResourcePoolInitializer{
		Prefix: "custom",
		ExtraProps: map[string]interface{}{
			"creator": currentWorkflowID,
			"purpose": "demonstration",
		},
	}

	// 创建带自定义初始化器的资源池
	resourcePool := NewResourcePoolWithInitializer(
		currentWorkflowID,
		"CustomResourcePoolDemo",
		customInitializer,
	)

	// 尝试获取资源
	releaseFunc, resourceInfo, err := resourcePool.AcquireResource(ctx, resourceID, 30*time.Minute, cancelable)
	if err != nil {
		logger.Error("failed to acquire resource", "Error", err)
		return err
	}

	logger.Info("successfully acquired custom resource",
		"resourceID", resourceInfo.GetID(),
		"resourceIndex", resourceInfo.GetIndex(),
		"metadata", resourceInfo.GetMetadata(),
		"acquiredTime", resourceInfo.GetAcquiredTime())

	// 模拟处理过程
	logger.Info("start critical operation")
	if processTime <= 0 {
		processTime = 10 * time.Second // 默认处理时间
	}
	_ = workflow.Sleep(ctx, processTime)
	logger.Info("critical operation completed")

	// 释放资源
	err = releaseFunc()
	if err != nil {
		logger.Error("failed to release resource", "Error", err)
		return err
	}

	logger.Info("resource released, custom resource pool workflow completed")
	return nil
}

// SampleWorkflowWithResourcePoolResizing 演示资源池动态扩展缩减的示例工作流
func SampleWorkflowWithResourcePoolResizing(
	ctx workflow.Context,
	resourceID string,
	initialSize int,
	expandSize int,
	shrinkSize int,
) error {
	currentWorkflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	logger := workflow.GetLogger(ctx)
	logger.Info("resource pool resizing demonstration workflow start",
		"workflowID", currentWorkflowID,
		"resourceID", resourceID,
		"initialSize", initialSize,
		"expandSize", expandSize,
		"shrinkSize", shrinkSize)

	if initialSize <= 0 {
		initialSize = 2
	}
	if expandSize <= initialSize {
		expandSize = initialSize + 2
	}
	if shrinkSize >= initialSize || shrinkSize <= 0 {
		shrinkSize = initialSize - 1
	}

	// 创建自定义初始化器
	customInitializer := &CustomResourcePoolInitializer{
		Prefix: "dynamic",
		ExtraProps: map[string]interface{}{
			"creator": currentWorkflowID,
			"purpose": "resize-demonstration",
		},
	}

	// 创建带自定义初始化器的资源池
	resourcePool := NewResourcePoolWithInitializer(
		currentWorkflowID,
		"DynamicResourcePoolDemo",
		customInitializer,
	)

	// 获取初始的两个资源用于演示
	logger.Info("start acquiring initial resources")
	var resources []Resource
	var releaseFuncs []ReleaseResourceFunc

	// 获取多个资源
	for i := 0; i < initialSize; i++ {
		releaseFunc, resource, err := resourcePool.AcquireResource(ctx, resourceID, 30*time.Minute, false)
		if err != nil {
			logger.Error("failed to acquire resource", "Error", err)
			return err
		}

		resources = append(resources, resource)
		releaseFuncs = append(releaseFuncs, releaseFunc)

		logger.Info("acquired initial resource",
			"index", i,
			"resourceID", resource.GetID(),
			"metadata", resource.GetMetadata())
	}

	// 模拟处理一些时间
	_ = workflow.Sleep(ctx, 2*time.Second)

	// 扩展资源池
	logger.Info("start expanding resource pool", "new size", expandSize)
	err := resourcePool.ResizePool(ctx, expandSize)
	if err != nil {
		logger.Error("failed to expand resource pool", "Error", err)
		return err
	}

	// 等待扩展生效
	_ = workflow.Sleep(ctx, 2*time.Second)

	// 获取新增的资源
	logger.Info("start acquiring expanded resources")
	for i := 0; i < expandSize-initialSize; i++ {
		releaseFunc, resource, err := resourcePool.AcquireResource(ctx, resourceID, 30*time.Minute, false)
		if err != nil {
			logger.Error("failed to acquire expanded resource", "Error", err)
			return err
		}

		resources = append(resources, resource)
		releaseFuncs = append(releaseFuncs, releaseFunc)

		logger.Info("acquired expanded resource",
			"index", initialSize+i,
			"resourceID", resource.GetID(),
			"metadata", resource.GetMetadata())
	}

	// 释放一些资源，为缩减做准备
	for i := 0; i < expandSize-shrinkSize; i++ {
		logger.Info("release resource to prepare for shrink",
			"index", i,
			"resourceID", resources[i].GetID())

		err := releaseFuncs[i]()
		if err != nil {
			logger.Error("failed to release resource", "Error", err)
		}
	}

	// 等待释放生效
	_ = workflow.Sleep(ctx, 2*time.Second)

	// 缩减资源池
	logger.Info("start shrinking resource pool", "new size", shrinkSize)
	err = resourcePool.ResizePool(ctx, shrinkSize)
	if err != nil {
		logger.Error("failed to shrink resource pool", "Error", err)
	}

	// 等待缩减生效
	_ = workflow.Sleep(ctx, 2*time.Second)

	// 释放剩余资源
	for i := expandSize - shrinkSize; i < len(releaseFuncs); i++ {
		logger.Info("release remaining resources",
			"index", i,
			"resourceID", resources[i].GetID())

		err := releaseFuncs[i]()
		if err != nil {
			logger.Error("failed to release resource", "Error", err)
		}
	}

	logger.Info("resource pool resizing demonstration workflow completed")
	return nil
}
