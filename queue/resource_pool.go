package queue

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// 资源池相关的信号常量
const (
	// ResourceAcquiredSignalName 资源获取信号通道名称
	ResourceAcquiredSignalName = "resource-acquired-signal"
	// RequestResourceSignalName 请求资源信号通道名称
	RequestResourceSignalName = "request-resource-signal"
	// ResourceReleasedSignalName 资源释放信号通道名称
	ResourceReleasedSignalName = "resource-released-signal"
	// UpdateResourcePoolSignalName 更新资源池大小信号
	UpdateResourcePoolSignalName = "update-resource-pool-signal"
	// CancelRequestSignalName 取消资源请求信号
	CancelRequestSignalName = "cancel-request-signal"
	// RequestCancelledSignalName 请求已取消信号
	RequestCancelledSignalName = "request-cancelled-signal"

	// 资源池查询相关常量
	// GetResourcePoolStatusQuery 获取资源池状态查询名称
	GetResourcePoolStatusQuery = "get-resource-pool-status-query"
	// GetResourceAllocationQuery 获取资源分配情况查询名称
	GetResourceAllocationQuery = "get-resource-allocation-query"

	// ClientContextKey 客户端上下文键
	ClientContextKey ContextKey = "Client"

	// InitializerContextKey 初始化器上下文键
	InitializerContextKey ContextKey = "ResourcePoolInitializer"
)

// 资源定义
type Resource interface {
	// GetID 获取资源ID
	GetID() string
	// GetMetadata 获取资源元数据
	GetMetadata() map[string]interface{}
	// GetAcquiredTime 获取资源获取时间
	GetAcquiredTime() time.Time
	// GetIndex 获取资源索引
	GetIndex() int
}

// ResourcePool 资源池接口
type ResourcePoolInterface interface {
	// AcquireResource 获取资源
	AcquireResource(ctx workflow.Context, resourceID string, releaseTimeout time.Duration, cancelable bool) (ReleaseResourceFunc, Resource, error)
	// GetNamespace 获取资源池命名空间
	GetNamespace() string
	// GetRequesterID 获取请求方ID
	GetRequesterID() string
	// QueryResourcePoolStatus 查询资源池状态
	QueryResourcePoolStatus(ctx workflow.Context) (*ResourcePoolStatus, error)
	// QueryResourceAllocation 查询资源分配详情
	QueryResourceAllocation(ctx workflow.Context) (*ResourceAllocation, error)
}

// ResourcePoolInitializer 资源池初始化器接口
type ResourcePoolInitializer interface {
	// Initialize 初始化资源池
	Initialize(ctx workflow.Context, resourceID string, totalResources int) *ResourcePoolState

	// ExpandPool 扩展资源池
	ExpandPool(ctx workflow.Context, state *ResourcePoolState, currentSize int, newSize int, resourceID string) int

	// ShrinkPool 缩减资源池
	ShrinkPool(ctx workflow.Context, state *ResourcePoolState, currentSize int, newSize int) int
}

// 类型定义
type (
	// ContextKey 上下文键类型
	ContextKey string

	// ReleaseResourceFunc 释放资源的函数类型
	ReleaseResourceFunc func() error

	// InternalResource 内部资源结构
	InternalResource struct {
		Info      ResourceInfo // 资源信息
		Available bool         // 是否可用
	}

	// ResourceRequest 资源请求结构
	ResourceRequest struct {
		WorkflowID string // 请求工作流ID
		Priority   int    // 优先级，数值越低优先级越高
	}

	// ResourceResponse 资源请求响应结构
	ResourceResponse struct {
		ResourceChannelName string       // 资源通道名称
		ResourceInfo        ResourceInfo // 资源详细信息
	}

	// ResourcePoolState 资源池状态
	ResourcePoolState struct {
		ResourcePool   []InternalResource // 资源池
		AvailableCount int                // 可用资源数
		WaitingQueue   []ResourceRequest  // 等待队列
		ActiveRequests map[string]int     // 工作流ID -> 资源索引
	}

	// UpdateResourcePoolRequest 更新资源池请求
	UpdateResourcePoolRequest struct {
		NewSize int // 新的资源池大小
	}

	// CancelRequestCommand 取消资源请求命令
	CancelRequestCommand struct {
		WorkflowID string // 要取消的工作流ID
	}

	// ResourcePoolStatus 资源池状态查询结果
	ResourcePoolStatus struct {
		ResourceID     string // 资源ID
		TotalResources int    // 总资源数
		AvailableCount int    // 可用资源数
		WaitingCount   int    // 等待队列长度
		AllocatedCount int    // 已分配资源数
	}

	// ResourceAllocation 资源分配详情
	ResourceAllocation struct {
		Resources      []ResourceDetail     // 资源详情列表
		WaitingQueue   []WaitingRequestInfo // 等待队列信息
		AllocatedCount int                  // 已分配资源数
	}

	// ResourceDetail 资源详情
	ResourceDetail struct {
		ResourceID    string                 // 资源ID
		ResourceIndex int                    // 资源索引
		Available     bool                   // 是否可用
		Metadata      map[string]interface{} // 资源元数据
		AcquiredTime  time.Time              // 资源获取时间，仅当资源被占用时有效
		AssignedTo    string                 // 分配给的工作流ID，仅当资源被占用时有效
	}

	// WaitingRequestInfo 等待队列中的请求信息
	WaitingRequestInfo struct {
		WorkflowID    string // 工作流ID
		Priority      int    // 优先级
		QueuePosition int    // 在队列中的位置
	}
)

// ResourceInfo 资源信息结构，实现Resource接口
type ResourceInfo struct {
	ResourceID    string                 // 资源ID
	Metadata      map[string]interface{} // 资源元数据
	AcquiredTime  time.Time              // 资源获取时间
	ResourceIndex int                    // 资源在池中的索引
}

// Resource接口实现
func (ri *ResourceInfo) GetID() string {
	return ri.ResourceID
}

func (ri *ResourceInfo) GetMetadata() map[string]interface{} {
	return ri.Metadata
}

func (ri *ResourceInfo) GetAcquiredTime() time.Time {
	return ri.AcquiredTime
}

func (ri *ResourceInfo) GetIndex() int {
	return ri.ResourceIndex
}

// IsDefaultInitializer 检查是否为默认初始化器
func IsDefaultInitializer(initializer ResourcePoolInitializer) bool {
	_, ok := initializer.(*DefaultResourcePoolInitializer)
	return ok
}

// ResourcePoolClient 资源池客户端结构
type ResourcePoolClient struct {
	RequesterID   string // 请求方工作流ID
	PoolNamespace string // 资源池命名空间
	PoolExecution *workflow.Execution
	Initializer   ResourcePoolInitializer // 资源池初始化器
}

// NewResourcePool 初始化资源池客户端
func NewResourcePool(requesterID string, poolNamespace string) *ResourcePoolClient {
	return &ResourcePoolClient{
		RequesterID:   requesterID,
		PoolNamespace: poolNamespace,
		Initializer:   &DefaultResourcePoolInitializer{},
	}
}

// NewResourcePoolWithInitializer 使用自定义初始化器创建资源池客户端
func NewResourcePoolWithInitializer(
	requesterID string,
	poolNamespace string,
	initializer ResourcePoolInitializer,
) *ResourcePoolClient {
	return &ResourcePoolClient{
		RequesterID:   requesterID,
		PoolNamespace: poolNamespace,
		Initializer:   initializer,
	}
}

// GetNamespace 获取资源池命名空间
func (rp *ResourcePoolClient) GetNamespace() string {
	return rp.PoolNamespace
}

// GetRequesterID 获取请求方ID
func (rp *ResourcePoolClient) GetRequesterID() string {
	return rp.RequesterID
}

// AcquireResource 获取资源，支持资源请求取消
func (rp *ResourcePoolClient) AcquireResource(ctx workflow.Context,
	resourceID string, releaseTimeout time.Duration, cancelable bool,
) (ReleaseResourceFunc, Resource, error) {
	// 设置本地活动选项
	activityCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	})

	// 获取当前workflow的信息用于日志
	logger := workflow.GetLogger(ctx)
	logger.Info("尝试获取资源", "workflowID", rp.RequesterID, "resourceID", resourceID)

	var resourceResponse ResourceResponse
	var execution workflow.Execution

	// 如果使用自定义初始化器，则使用带初始化器的活动
	if rp.Initializer != nil && !IsDefaultInitializer(rp.Initializer) {
		// 发送信号并启动自定义资源池workflow
		err := workflow.ExecuteLocalActivity(activityCtx,
			SignalWithStartCustomResourcePoolWorkflowActivity, rp.PoolNamespace,
			resourceID, rp.RequesterID, 0, releaseTimeout,
			ResourcePoolWorkflowWithInitializer,
			rp.PoolNamespace, resourceID, 1, releaseTimeout, rp.Initializer).Get(ctx, &execution)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// 使用默认初始化器
		err := workflow.ExecuteLocalActivity(activityCtx,
			SignalWithStartResourcePoolWorkflowActivity, rp.PoolNamespace,
			resourceID, rp.RequesterID, 0, releaseTimeout).Get(ctx, &execution)
		if err != nil {
			return nil, nil, err
		}
	}

	// 保存执行信息
	rp.PoolExecution = &execution

	// 如果请求可取消，提供等待与取消机制
	if cancelable {
		// 等待资源获取或取消信号
		resourceCh := workflow.GetSignalChannel(ctx, ResourceAcquiredSignalName)
		cancelCh := workflow.GetSignalChannel(ctx, RequestCancelledSignalName)

		// 使用选择器等待任一信号
		selector := workflow.NewSelector(ctx)
		received := false

		// 处理资源获取信号
		selector.AddReceive(resourceCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &resourceResponse)
			received = true
		})

		// 处理取消信号
		selector.AddReceive(cancelCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			logger.Info("资源请求被取消", "workflowID", rp.RequesterID)
			received = false // 标记为未收到资源
		})

		// 等待任一信号
		selector.Select(ctx)

		// 如果未收到资源获取信号，表示取消了
		if !received {
			return nil, nil, fmt.Errorf("资源请求已被取消")
		}
	} else {
		// 普通不可取消的获取，直接等待获取信号
		workflow.GetSignalChannel(ctx, ResourceAcquiredSignalName).
			Receive(ctx, &resourceResponse)
	}

	logger.Info("成功获取资源", "workflowID", rp.RequesterID, "resourceID", resourceID,
		"resourceInfo", resourceResponse.ResourceInfo)

	// 创建资源释放函数
	releaseFunc := func() error {
		logger.Info("释放资源", "workflowID", rp.RequesterID, "resourceID", resourceID)
		return workflow.SignalExternalWorkflow(ctx, execution.ID, execution.RunID,
			resourceResponse.ResourceChannelName, "releaseResource").Get(ctx, nil)
	}
	return releaseFunc, &resourceResponse.ResourceInfo, nil
}

// ResizePool 调整资源池大小
func (rp *ResourcePoolClient) ResizePool(ctx workflow.Context, newSize int) error {
	if rp.PoolExecution == nil {
		return fmt.Errorf("资源池未初始化")
	}

	return UpdateResourcePool(ctx, *rp.PoolExecution, newSize)
}

// CancelRequest 取消资源请求
func (rp *ResourcePoolClient) CancelRequest(ctx workflow.Context, workflowID string) error {
	if rp.PoolExecution == nil {
		return fmt.Errorf("资源池未初始化")
	}

	return CancelResourceRequest(ctx, *rp.PoolExecution, workflowID)
}

// ResourcePoolWorkflowImpl 资源池工作流结构体
type ResourcePoolWorkflowImpl struct {
	Namespace      string                  // 命名空间
	ResourceID     string                  // 资源ID
	ReleaseTimeout time.Duration           // 资源释放超时时间
	Initializer    ResourcePoolInitializer // 资源池初始化器
	State          *ResourcePoolState      // 资源池状态
}

// NewResourcePoolWorkflow 创建新的资源池工作流
func NewResourcePoolWorkflow(
	ctx workflow.Context,
	namespace string,
	resourceID string,
	totalResources int,
	releaseTimeout time.Duration,
	initializer ResourcePoolInitializer,
) *ResourcePoolWorkflowImpl {
	if initializer == nil {
		initializer = &DefaultResourcePoolInitializer{}
	}

	rpw := &ResourcePoolWorkflowImpl{
		Namespace:      namespace,
		ResourceID:     resourceID,
		ReleaseTimeout: releaseTimeout,
		Initializer:    initializer,
	}

	// 将初始化器存入上下文中
	ctx = workflow.WithValue(ctx, InitializerContextKey, initializer)

	// 使用初始化器创建资源池状态
	if totalResources <= 0 {
		totalResources = 1 // 默认至少有一个资源
	}

	rpw.State = initializer.Initialize(ctx, resourceID, totalResources)
	return rpw
}

// Run 运行资源池工作流
func (rpw *ResourcePoolWorkflowImpl) Run(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	workflowInfo := workflow.GetInfo(ctx)
	logger.Info("资源池工作流启动",
		"workflowID", workflowInfo.WorkflowExecution.ID,
		"resourceID", rpw.ResourceID)

	// 设置信号处理通道
	requestResourceCh := workflow.GetSignalChannel(ctx, RequestResourceSignalName)
	updateResourcePoolCh := workflow.GetSignalChannel(ctx, UpdateResourcePoolSignalName)
	cancelRequestCh := workflow.GetSignalChannel(ctx, CancelRequestSignalName)

	// 注册查询处理器
	if err := workflow.SetQueryHandler(ctx, GetResourcePoolStatusQuery, rpw.handleResourcePoolStatusQuery); err != nil {
		logger.Error("注册资源池状态查询处理器失败", "Error", err)
		return err
	}

	if err := workflow.SetQueryHandler(ctx, GetResourceAllocationQuery, rpw.handleResourceAllocationQuery); err != nil {
		logger.Error("注册资源分配详情查询处理器失败", "Error", err)
		return err
	}

	// 无限循环处理资源请求
	for {
		// 使用选择器处理不同的事件
		selector := workflow.NewSelector(ctx)

		// 处理资源请求信号
		selector.AddReceive(requestResourceCh, func(c workflow.ReceiveChannel, _ bool) {
			var request ResourceRequest
			c.Receive(ctx, &request)
			rpw.HandleResourceRequest(ctx, request)
		})

		// 处理资源池大小更新信号
		selector.AddReceive(updateResourcePoolCh, func(c workflow.ReceiveChannel, _ bool) {
			var updateRequest UpdateResourcePoolRequest
			c.Receive(ctx, &updateRequest)
			rpw.HandlePoolSizeUpdate(ctx, updateRequest)
		})

		// 处理取消资源请求信号
		selector.AddReceive(cancelRequestCh, func(c workflow.ReceiveChannel, _ bool) {
			var cancelCmd CancelRequestCommand
			c.Receive(ctx, &cancelCmd)
			rpw.HandleCancelRequest(ctx, cancelCmd)
		})

		// 执行选择器
		selector.Select(ctx)
	}
}

// HandleResourceRequest 处理资源请求
func (rpw *ResourcePoolWorkflowImpl) HandleResourceRequest(
	ctx workflow.Context,
	request ResourceRequest,
) {
	logger := workflow.GetLogger(ctx)
	logger.Info("收到资源请求",
		"requesterID", request.WorkflowID,
		"availableCount", rpw.State.AvailableCount)

	if rpw.State.AvailableCount > 0 {
		// 有可用资源，查找可用资源索引
		resourceIndex := rpw.FindAvailableResource()
		if resourceIndex == -1 {
			logger.Error("资源计数与实际可用资源不一致",
				"availableCount", rpw.State.AvailableCount)
			return
		}

		// 分配资源
		rpw.AllocateResource(ctx, resourceIndex, request)
	} else {
		// 没有可用资源，加入等待队列
		logger.Info("资源不可用，添加到等待队列", "requesterID", request.WorkflowID)
		rpw.State.WaitingQueue = append(rpw.State.WaitingQueue, request)
	}
}

// FindAvailableResource 查找可用资源
func (rpw *ResourcePoolWorkflowImpl) FindAvailableResource() int {
	for i, res := range rpw.State.ResourcePool {
		if res.Available {
			return i
		}
	}
	return -1
}

// AllocateResource 分配资源
func (rpw *ResourcePoolWorkflowImpl) AllocateResource(
	ctx workflow.Context,
	resourceIndex int,
	request ResourceRequest,
) {
	// 标记资源为已占用
	rpw.State.ResourcePool[resourceIndex].Available = false
	rpw.State.ResourcePool[resourceIndex].Info.AcquiredTime = workflow.Now(ctx)
	rpw.State.AvailableCount--
	rpw.State.ActiveRequests[request.WorkflowID] = resourceIndex

	// 生成唯一的资源通道名称 - 不再使用SideEffect
	resourceChannelName := generateResourceChannelName(request.WorkflowID)

	// 准备资源响应
	response := ResourceResponse{
		ResourceChannelName: resourceChannelName,
		ResourceInfo:        rpw.State.ResourcePool[resourceIndex].Info,
	}

	// 通知请求方已获取资源
	err := workflow.SignalExternalWorkflow(ctx, request.WorkflowID, "",
		ResourceAcquiredSignalName, response).Get(ctx, nil)
	if err != nil {
		logger := workflow.GetLogger(ctx)
		logger.Info("通知请求方失败，资源将被重新释放", "Error", err)
		rpw.State.ResourcePool[resourceIndex].Available = true
		rpw.State.AvailableCount++
		delete(rpw.State.ActiveRequests, request.WorkflowID)
		return
	}

	// 创建资源释放通道
	releaseChannel := workflow.GetSignalChannel(ctx, resourceChannelName)

	// 在新协程中处理资源释放和超时
	workflow.Go(ctx, func(childCtx workflow.Context) {
		childSelector := workflow.NewSelector(childCtx)

		// 添加超时处理
		if rpw.ReleaseTimeout > 0 {
			childSelector.AddFuture(workflow.NewTimer(childCtx, rpw.ReleaseTimeout), func(f workflow.Future) {
				rpw.HandleResourceRelease(childCtx, resourceIndex, request.WorkflowID, "timeout")
			})
		}

		// 添加资源释放信号处理
		childSelector.AddReceive(releaseChannel, func(c workflow.ReceiveChannel, more bool) {
			var ack string
			c.Receive(childCtx, &ack)
			rpw.HandleResourceRelease(childCtx, resourceIndex, request.WorkflowID, "release")
		})

		// 等待任一事件发生
		childSelector.Select(childCtx)
	})
}

// HandleResourceRelease 处理资源释放
func (rpw *ResourcePoolWorkflowImpl) HandleResourceRelease(
	ctx workflow.Context,
	resourceIndex int,
	requesterID string,
	reason string,
) {
	logger := workflow.GetLogger(ctx)
	if reason == "timeout" {
		logger.Info("资源使用超时，自动释放", "requesterID", requesterID)
	} else {
		logger.Info("资源已被释放", "requesterID", requesterID)
	}

	// 释放资源
	rpw.State.ResourcePool[resourceIndex].Available = true
	rpw.State.AvailableCount++
	delete(rpw.State.ActiveRequests, requesterID)

	// 处理等待队列中的下一个请求
	if len(rpw.State.WaitingQueue) > 0 {
		rpw.ProcessNextRequest(ctx)
	}
}

// HandlePoolSizeUpdate 处理资源池大小更新
func (rpw *ResourcePoolWorkflowImpl) HandlePoolSizeUpdate(
	ctx workflow.Context,
	updateRequest UpdateResourcePoolRequest,
) {
	logger := workflow.GetLogger(ctx)
	logger.Info("收到资源池大小更新请求",
		"当前大小", len(rpw.State.ResourcePool),
		"新大小", updateRequest.NewSize,
		"可用资源", rpw.State.AvailableCount)

	if updateRequest.NewSize <= 0 {
		logger.Info("资源池大小必须大于0，忽略本次更新")
		return
	}

	currentSize := len(rpw.State.ResourcePool)
	// 处理扩容
	if updateRequest.NewSize > currentSize {
		rpw.ExpandResourcePool(ctx, updateRequest.NewSize)
	} else if updateRequest.NewSize < currentSize {
		rpw.ShrinkResourcePool(ctx, updateRequest.NewSize)
	}
}

// ExpandResourcePool 扩展资源池
func (rpw *ResourcePoolWorkflowImpl) ExpandResourcePool(
	ctx workflow.Context,
	newSize int,
) {
	logger := workflow.GetLogger(ctx)
	currentSize := len(rpw.State.ResourcePool)

	// 使用初始化器扩展资源池
	addedCount := rpw.Initializer.ExpandPool(ctx, rpw.State, currentSize, newSize, rpw.ResourceID)

	// 更新可用资源计数
	rpw.State.AvailableCount += addedCount

	logger.Info("资源池扩大",
		"新增资源数", addedCount,
		"当前可用资源", rpw.State.AvailableCount)

	// 处理等待队列中的请求
	rpw.ProcessWaitingRequests(ctx)
}

// ShrinkResourcePool 缩减资源池
func (rpw *ResourcePoolWorkflowImpl) ShrinkResourcePool(
	ctx workflow.Context,
	newSize int,
) {
	logger := workflow.GetLogger(ctx)
	currentSize := len(rpw.State.ResourcePool)

	// 使用初始化器缩减资源池
	reducedCount := rpw.Initializer.ShrinkPool(ctx, rpw.State, currentSize, newSize)

	// 更新可用计数
	if rpw.State.AvailableCount >= reducedCount {
		rpw.State.AvailableCount -= reducedCount
		logger.Info("资源池缩小",
			"减少资源数", reducedCount,
			"当前可用资源", rpw.State.AvailableCount)
	} else {
		logger.Info("资源计数不一致，将进行修正",
			"oldCount", rpw.State.AvailableCount,
			"newCount", 0)
		rpw.State.AvailableCount = 0
	}
}

// HandleCancelRequest 处理取消资源请求
func (rpw *ResourcePoolWorkflowImpl) HandleCancelRequest(
	ctx workflow.Context,
	cancelCmd CancelRequestCommand,
) {
	logger := workflow.GetLogger(ctx)
	logger.Info("收到取消资源请求命令", "workflowID", cancelCmd.WorkflowID)

	// 检查是否在等待队列中
	for i, req := range rpw.State.WaitingQueue {
		if req.WorkflowID == cancelCmd.WorkflowID {
			// 从等待队列中移除
			rpw.State.WaitingQueue = append(rpw.State.WaitingQueue[:i], rpw.State.WaitingQueue[i+1:]...)
			logger.Info("已从等待队列中移除工作流", "workflowID", cancelCmd.WorkflowID)

			// 通知工作流资源请求已取消
			err := workflow.SignalExternalWorkflow(ctx, cancelCmd.WorkflowID, "",
				RequestCancelledSignalName, nil).Get(ctx, nil)
			if err != nil {
				logger.Info("通知工作流取消失败", "Error", err)
			}

			return
		}
	}

	logger.Info("工作流不在等待队列中", "workflowID", cancelCmd.WorkflowID)
}

// ProcessNextRequest 处理等待队列中的下一个请求
func (rpw *ResourcePoolWorkflowImpl) ProcessNextRequest(
	ctx workflow.Context,
) {
	if len(rpw.State.WaitingQueue) == 0 || rpw.State.AvailableCount <= 0 {
		return
	}

	logger := workflow.GetLogger(ctx)

	// 查找可用资源
	resourceIndex := rpw.FindAvailableResource()
	if resourceIndex == -1 {
		logger.Error("资源计数与实际可用资源不一致",
			"availableCount", rpw.State.AvailableCount)
		return
	}

	// 取出下一个请求
	nextRequest := rpw.State.WaitingQueue[0]
	rpw.State.WaitingQueue = rpw.State.WaitingQueue[1:]

	// 分配资源
	rpw.AllocateResource(ctx, resourceIndex, nextRequest)
}

// ProcessWaitingRequests 处理所有等待中的请求
func (rpw *ResourcePoolWorkflowImpl) ProcessWaitingRequests(
	ctx workflow.Context,
) {
	// 循环处理等待队列中的请求，直到没有可用资源或队列为空
	for rpw.State.AvailableCount > 0 && len(rpw.State.WaitingQueue) > 0 {
		// 处理下一个请求
		rpw.ProcessNextRequest(ctx)
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("等待队列处理完成",
		"剩余可用资源", rpw.State.AvailableCount,
		"等待队列长度", len(rpw.State.WaitingQueue))
}

// ResourcePoolWorkflow 资源池工作流，管理资源的分配和释放
func ResourcePoolWorkflow(
	ctx workflow.Context,
	namespace string,
	resourceID string,
	totalResources int,
	releaseTimeout time.Duration,
) error {
	return ResourcePoolWorkflowWithInitializer(
		ctx,
		namespace,
		resourceID,
		totalResources,
		releaseTimeout,
		&DefaultResourcePoolInitializer{},
	)
}

// ResourcePoolWorkflowWithInitializer 带初始化器的资源池工作流
func ResourcePoolWorkflowWithInitializer(
	ctx workflow.Context,
	namespace string,
	resourceID string,
	totalResources int,
	releaseTimeout time.Duration,
	initializer ResourcePoolInitializer,
) error {
	// 创建工作流结构体并运行
	rpw := NewResourcePoolWorkflow(
		ctx,
		namespace,
		resourceID,
		totalResources,
		releaseTimeout,
		initializer,
	)

	return rpw.Run(ctx)
}

// 以下为现有的外部接口，保持兼容性

// SignalWithStartResourcePoolWorkflowActivity 发送信号并启动资源池工作流
func SignalWithStartResourcePoolWorkflowActivity(
	ctx context.Context,
	namespace string,
	resourceID string,
	requesterWorkflowID string,
	priority int,
	releaseTimeout time.Duration,
) (*workflow.Execution, error) {
	c, ok := ctx.Value(ClientContextKey).(client.Client)
	if !ok || c == nil {
		return nil, fmt.Errorf("无法从上下文中获取有效的 Temporal 客户端")
	}

	// 生成资源池工作流的唯一ID
	workflowID := fmt.Sprintf(
		"%s:%s:%s",
		"resource-pool",
		namespace,
		resourceID,
	)

	// 配置工作流选项
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "resource-pool",
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}

	// 创建资源请求
	request := ResourceRequest{
		WorkflowID: requesterWorkflowID,
		Priority:   priority,
	}

	// 发送信号并启动工作流
	wr, err := c.SignalWithStartWorkflow(
		ctx, workflowID, RequestResourceSignalName, request,
		workflowOptions, ResourcePoolWorkflow, namespace, resourceID, 1, releaseTimeout)
	if err != nil {
		activity.GetLogger(ctx).Error("无法发送信号和启动工作流", "Error", err)
		return nil, err
	}

	activity.GetLogger(ctx).Info("信号发送和工作流启动成功",
		"WorkflowID", wr.GetID(),
		"RunID", wr.GetRunID())

	return &workflow.Execution{
		ID:    wr.GetID(),
		RunID: wr.GetRunID(),
	}, nil
}

// SignalWithStartCustomResourcePoolWorkflowActivity 发送信号并启动自定义资源池工作流
func SignalWithStartCustomResourcePoolWorkflowActivity(
	ctx context.Context,
	namespace string,
	resourceID string,
	requesterWorkflowID string,
	priority int,
	releaseTimeout time.Duration,
	workflowFunc interface{},
	workflowArgs ...interface{},
) (*workflow.Execution, error) {
	c, ok := ctx.Value(ClientContextKey).(client.Client)
	if !ok || c == nil {
		return nil, fmt.Errorf("无法从上下文中获取有效的 Temporal 客户端")
	}

	// 生成资源池工作流的唯一ID
	workflowID := fmt.Sprintf(
		"%s:%s:%s",
		"resource-pool",
		namespace,
		resourceID,
	)

	// 配置工作流选项
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "resource-pool",
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}

	// 创建资源请求
	request := ResourceRequest{
		WorkflowID: requesterWorkflowID,
		Priority:   priority,
	}

	// 发送信号并启动工作流
	wr, err := c.SignalWithStartWorkflow(
		ctx, workflowID, RequestResourceSignalName, request,
		workflowOptions, workflowFunc, workflowArgs...)
	if err != nil {
		activity.GetLogger(ctx).Error("无法发送信号和启动自定义工作流", "Error", err)
		return nil, err
	}

	activity.GetLogger(ctx).Info("信号发送和自定义工作流启动成功",
		"WorkflowID", wr.GetID(),
		"RunID", wr.GetRunID())

	return &workflow.Execution{
		ID:    wr.GetID(),
		RunID: wr.GetRunID(),
	}, nil
}

// generateResourceChannelName 生成资源通道名称
func generateResourceChannelName(requesterWorkflowID string) string {
	return fmt.Sprintf("resource-channel-%s", requesterWorkflowID)
}

// UpdateResourcePool 更新资源池大小
func UpdateResourcePool(
	ctx workflow.Context,
	execution workflow.Execution,
	newSize int,
) error {
	logger := workflow.GetLogger(ctx)

	// 创建更新请求
	updateRequest := UpdateResourcePoolRequest{
		NewSize: newSize,
	}

	// 发送更新信号
	err := workflow.SignalExternalWorkflow(ctx, execution.ID, execution.RunID,
		UpdateResourcePoolSignalName, updateRequest).Get(ctx, nil)
	if err != nil {
		logger.Error("更新资源池失败", "Error", err)
		return err
	}

	logger.Info("已发送资源池更新请求", "newSize", newSize)
	return nil
}

// CancelResourceRequest 取消在等待队列中的资源请求
func CancelResourceRequest(
	ctx workflow.Context,
	execution workflow.Execution,
	workflowID string,
) error {
	logger := workflow.GetLogger(ctx)

	// 创建取消命令
	cancelCmd := CancelRequestCommand{
		WorkflowID: workflowID,
	}

	// 发送取消信号
	err := workflow.SignalExternalWorkflow(ctx, execution.ID, execution.RunID,
		CancelRequestSignalName, cancelCmd).Get(ctx, nil)
	if err != nil {
		logger.Error("取消资源请求失败", "Error", err)
		return err
	}

	logger.Info("已发送取消资源请求命令", "workflowID", workflowID)
	return nil
}

// ================================================
// 默认资源池初始化器
// ================================================

// DefaultResourcePoolInitializer 默认资源池初始化器
type DefaultResourcePoolInitializer struct{}

// Initialize 初始化默认资源池
func (di *DefaultResourcePoolInitializer) Initialize(
	ctx workflow.Context,
	resourceID string,
	totalResources int,
) *ResourcePoolState {
	// 创建资源池状态
	state := &ResourcePoolState{
		ResourcePool:   make([]InternalResource, totalResources),
		AvailableCount: totalResources,
		WaitingQueue:   []ResourceRequest{},
		ActiveRequests: make(map[string]int),
	}

	// 初始化资源池
	for i := 0; i < totalResources; i++ {
		state.ResourcePool[i] = InternalResource{
			Info: ResourceInfo{
				ResourceID:    fmt.Sprintf("%s-%d", resourceID, i),
				Metadata:      map[string]interface{}{"poolIndex": i},
				ResourceIndex: i,
			},
			Available: true,
		}
	}

	return state
}

// ExpandPool 默认的资源池扩展实现
func (di *DefaultResourcePoolInitializer) ExpandPool(
	ctx workflow.Context,
	state *ResourcePoolState,
	currentSize int,
	newSize int,
	resourceID string,
) int {
	// 扩展资源池
	for i := currentSize; i < newSize; i++ {
		state.ResourcePool = append(state.ResourcePool, InternalResource{
			Info: ResourceInfo{
				ResourceID:    fmt.Sprintf("%s-%d", resourceID, i),
				Metadata:      map[string]interface{}{"poolIndex": i},
				ResourceIndex: i,
			},
			Available: true,
		})
	}

	// 返回新增的可用资源数量
	return newSize - currentSize
}

// ShrinkPool 默认的资源池缩减实现
func (di *DefaultResourcePoolInitializer) ShrinkPool(
	ctx workflow.Context,
	state *ResourcePoolState,
	currentSize int,
	newSize int,
) int {
	logger := workflow.GetLogger(ctx)

	// 计算需要减少的资源数量
	reduceCount := currentSize - newSize
	if reduceCount <= 0 {
		return 0 // 没有需要缩减的资源
	}

	// 处理缩容，先统计可用资源数量
	availableCount := 0
	for i := currentSize - 1; i >= 0; i-- {
		if state.ResourcePool[i].Available {
			availableCount++
		}
	}

	logger.Info("开始资源池缩容",
		"当前资源数", currentSize,
		"目标资源数", newSize,
		"可用资源数", availableCount,
		"已使用资源数", currentSize-availableCount)

	// 确定可以立即缩减的资源数量
	immediateReduceCount := min(availableCount, reduceCount)

	// 标记要缩减的资源
	availableReduceCount := 0
	pendingReduceIndexes := make([]int, 0, reduceCount-immediateReduceCount)

	// 从后向前缩减可用资源
	for i := currentSize - 1; i >= 0 && availableReduceCount < immediateReduceCount; i-- {
		if state.ResourcePool[i].Available {
			// 标记为不可用的资源
			state.ResourcePool[i].Available = false
			availableReduceCount++
		} else if availableReduceCount+len(pendingReduceIndexes) < reduceCount {
			// 记录需要延迟缩减的资源索引
			pendingReduceIndexes = append(pendingReduceIndexes, i)
		}
	}

	logger.Info("资源池缩容进展",
		"立即缩减资源数", availableReduceCount,
		"待延迟缩减资源数", len(pendingReduceIndexes))

	// 如果有需要延迟缩减的资源，则启动观察协程
	if len(pendingReduceIndexes) > 0 {
		workflow.Go(ctx, func(ctx workflow.Context) {
			di.monitorPendingResourcesForShrink(ctx, state, pendingReduceIndexes)
		})
	}

	// 返回已经减少的可用资源数
	return availableReduceCount
}

// monitorPendingResourcesForShrink 监控延迟缩减的资源
func (di *DefaultResourcePoolInitializer) monitorPendingResourcesForShrink(
	ctx workflow.Context,
	state *ResourcePoolState,
	pendingIndexes []int,
) {
	logger := workflow.GetLogger(ctx)
	logger.Info("启动延迟缩容监控", "待缩减资源数", len(pendingIndexes))

	// 创建待观察资源的映射，方便查找
	pendingMap := make(map[int]bool)
	for _, idx := range pendingIndexes {
		pendingMap[idx] = true
	}

	// 每隔一段时间检查一次资源状态
	for len(pendingMap) > 0 {
		// 等待一段时间再检查
		_ = workflow.Sleep(ctx, 30*time.Second)

		// 检查是否有资源已经释放
		for idx := range pendingMap {
			if idx < len(state.ResourcePool) && state.ResourcePool[idx].Available {
				// 资源已释放，可以缩减
				state.ResourcePool[idx].Available = false
				logger.Info("延迟缩容：资源已释放，标记为不可用", "resourceIndex", idx)
				delete(pendingMap, idx)

				// 更新可用资源计数
				if state.AvailableCount > 0 {
					state.AvailableCount--
				}
			}
		}

		// 记录剩余待缩减资源数
		if len(pendingMap) > 0 {
			logger.Info("延迟缩容进行中", "剩余待缩减资源数", len(pendingMap))
		}
	}

	logger.Info("延迟缩容完成，所有标记的资源都已缩减")
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// QueryResourcePoolStatus 查询资源池状态
func (rp *ResourcePoolClient) QueryResourcePoolStatus(ctx workflow.Context) (*ResourcePoolStatus, error) {
	if rp.PoolExecution == nil {
		return nil, fmt.Errorf("资源池未初始化")
	}

	// 使用查询接口获取资源池状态
	logger := workflow.GetLogger(ctx)
	logger.Info("查询资源池状态", "resourcePoolID", rp.PoolExecution.ID)

	// 设置本地活动选项
	activityCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
	})

	var result ResourcePoolStatus
	err := workflow.ExecuteLocalActivity(activityCtx, QueryResourcePoolStatusActivity,
		rp.PoolExecution.ID, rp.PoolExecution.RunID).Get(ctx, &result)
	if err != nil {
		logger.Error("查询资源池状态失败", "Error", err)
		return nil, err
	}

	return &result, nil
}

// QueryResourceAllocation 查询资源分配详情
func (rp *ResourcePoolClient) QueryResourceAllocation(ctx workflow.Context) (*ResourceAllocation, error) {
	if rp.PoolExecution == nil {
		return nil, fmt.Errorf("资源池未初始化")
	}

	// 使用查询接口获取资源分配详情
	logger := workflow.GetLogger(ctx)
	logger.Info("查询资源分配详情", "resourcePoolID", rp.PoolExecution.ID)

	// 设置本地活动选项
	activityCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
	})

	var result ResourceAllocation
	err := workflow.ExecuteLocalActivity(activityCtx, QueryResourceAllocationActivity,
		rp.PoolExecution.ID, rp.PoolExecution.RunID).Get(ctx, &result)
	if err != nil {
		logger.Error("查询资源分配详情失败", "Error", err)
		return nil, err
	}

	return &result, nil
}

// QueryResourcePoolStatus 查询资源池状态
func QueryResourcePoolStatus(
	ctx workflow.Context,
	execution workflow.Execution,
) (*ResourcePoolStatus, error) {
	logger := workflow.GetLogger(ctx)

	// 设置本地活动选项
	activityCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
	})

	var result ResourcePoolStatus
	err := workflow.ExecuteLocalActivity(activityCtx, QueryResourcePoolStatusActivity,
		execution.ID, execution.RunID).Get(ctx, &result)
	if err != nil {
		logger.Error("查询资源池状态失败", "Error", err)
		return nil, err
	}

	return &result, nil
}

// QueryResourceAllocation 查询资源分配详情
func QueryResourceAllocation(
	ctx workflow.Context,
	execution workflow.Execution,
) (*ResourceAllocation, error) {
	logger := workflow.GetLogger(ctx)

	// 设置本地活动选项
	activityCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
	})

	var result ResourceAllocation
	err := workflow.ExecuteLocalActivity(activityCtx, QueryResourceAllocationActivity,
		execution.ID, execution.RunID).Get(ctx, &result)
	if err != nil {
		logger.Error("查询资源分配详情失败", "Error", err)
		return nil, err
	}

	return &result, nil
}

// QueryResourcePoolStatusActivity 查询资源池状态活动
func QueryResourcePoolStatusActivity(
	ctx context.Context,
	workflowID string,
	runID string,
) (ResourcePoolStatus, error) {
	c, ok := ctx.Value(ClientContextKey).(client.Client)
	if !ok || c == nil {
		return ResourcePoolStatus{}, fmt.Errorf("无法从上下文中获取有效的 Temporal 客户端")
	}

	// 使用客户端查询工作流
	resp, err := c.QueryWorkflow(ctx, workflowID, runID, GetResourcePoolStatusQuery)
	if err != nil {
		activity.GetLogger(ctx).Error("查询资源池状态失败", "Error", err)
		return ResourcePoolStatus{}, err
	}

	var result ResourcePoolStatus
	if err := resp.Get(&result); err != nil {
		activity.GetLogger(ctx).Error("解析查询结果失败", "Error", err)
		return ResourcePoolStatus{}, err
	}

	return result, nil
}

// QueryResourceAllocationActivity 查询资源分配详情活动
func QueryResourceAllocationActivity(
	ctx context.Context,
	workflowID string,
	runID string,
) (ResourceAllocation, error) {
	c, ok := ctx.Value(ClientContextKey).(client.Client)
	if !ok || c == nil {
		return ResourceAllocation{}, fmt.Errorf("无法从上下文中获取有效的 Temporal 客户端")
	}

	// 使用客户端查询工作流
	resp, err := c.QueryWorkflow(ctx, workflowID, runID, GetResourceAllocationQuery)
	if err != nil {
		activity.GetLogger(ctx).Error("查询资源分配详情失败", "Error", err)
		return ResourceAllocation{}, err
	}

	var result ResourceAllocation
	if err := resp.Get(&result); err != nil {
		activity.GetLogger(ctx).Error("解析查询结果失败", "Error", err)
		return ResourceAllocation{}, err
	}

	return result, nil
}

// handleResourcePoolStatusQuery 处理资源池状态查询
func (rpw *ResourcePoolWorkflowImpl) handleResourcePoolStatusQuery() (ResourcePoolStatus, error) {
	totalResources := len(rpw.State.ResourcePool)
	allocatedCount := totalResources - rpw.State.AvailableCount

	return ResourcePoolStatus{
		ResourceID:     rpw.ResourceID,
		TotalResources: totalResources,
		AvailableCount: rpw.State.AvailableCount,
		WaitingCount:   len(rpw.State.WaitingQueue),
		AllocatedCount: allocatedCount,
	}, nil
}

// handleResourceAllocationQuery 处理资源分配详情查询
func (rpw *ResourcePoolWorkflowImpl) handleResourceAllocationQuery() (ResourceAllocation, error) {
	// 构建资源详情列表
	resources := make([]ResourceDetail, len(rpw.State.ResourcePool))

	// 填充资源详情
	for i, res := range rpw.State.ResourcePool {
		resourceDetail := ResourceDetail{
			ResourceID:    res.Info.ResourceID,
			ResourceIndex: res.Info.ResourceIndex,
			Available:     res.Available,
			Metadata:      res.Info.Metadata,
		}

		// 如果资源不可用，查找它被分配给的工作流
		if !res.Available {
			for workflowID, resIndex := range rpw.State.ActiveRequests {
				if resIndex == i {
					resourceDetail.AssignedTo = workflowID
					resourceDetail.AcquiredTime = res.Info.AcquiredTime
					break
				}
			}
		}

		resources[i] = resourceDetail
	}

	// 构建等待队列信息
	waitingQueue := make([]WaitingRequestInfo, len(rpw.State.WaitingQueue))
	for i, req := range rpw.State.WaitingQueue {
		waitingQueue[i] = WaitingRequestInfo{
			WorkflowID:    req.WorkflowID,
			Priority:      req.Priority,
			QueuePosition: i + 1,
		}
	}

	return ResourceAllocation{
		Resources:      resources,
		WaitingQueue:   waitingQueue,
		AllocatedCount: len(rpw.State.ActiveRequests),
	}, nil
}
