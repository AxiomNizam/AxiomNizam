package runtime

import (
	"context"
	"fmt"
	"log"
	"sync"

	"example.com/axiomnizam/internal/apiserver"
	"example.com/axiomnizam/internal/controllers"
	"example.com/axiomnizam/internal/workqueue"
)

// ControllerManager manages all controllers in the system
type ControllerManager struct {
	mu          sync.RWMutex
	controllers map[string]ManagedController
	store       *apiserver.ResourceStore
	leaderElect bool
	namespace   string
}

// ManagedController wraps a controller with metadata
type ManagedController struct {
	name       string
	controller *controllers.ResourceController
	running    bool
}

// NewControllerManager creates a new controller manager
func NewControllerManager(store *apiserver.ResourceStore, leaderElect bool, namespace string) *ControllerManager {
	return &ControllerManager{
		controllers: make(map[string]ManagedController),
		store:       store,
		leaderElect: leaderElect,
		namespace:   namespace,
	}
}

// RegisterWorkloadController registers the workload controller
func (cm *ControllerManager) RegisterWorkloadController() error {
	workQueue := workqueue.NewSimpleQueue(nil)
	reconciler := controllers.NewWorkloadReconciler(cm.store)

	controller := controllers.NewResourceController(
		"workload",
		workQueue,
		cm.store,
		reconciler,
		3, // max concurrent
	)

	cm.mu.Lock()
	cm.controllers["workload"] = ManagedController{
		name:       "workload",
		controller: controller,
		running:    false,
	}
	cm.mu.Unlock()

	log.Println("Registered workload controller")
	return nil
}

// RegisterPipelineController registers the pipeline controller
func (cm *ControllerManager) RegisterPipelineController() error {
	workQueue := workqueue.NewSimpleQueue(nil)
	reconciler := controllers.NewPipelineReconciler(cm.store)

	controller := controllers.NewResourceController(
		"pipeline",
		workQueue,
		cm.store,
		reconciler,
		3, // max concurrent
	)

	cm.mu.Lock()
	cm.controllers["pipeline"] = ManagedController{
		name:       "pipeline",
		controller: controller,
		running:    false,
	}
	cm.mu.Unlock()

	log.Println("Registered pipeline controller")
	return nil
}

// RegisterScheduleController registers the schedule controller
func (cm *ControllerManager) RegisterScheduleController() error {
	workQueue := workqueue.NewSimpleQueue(nil)
	reconciler := controllers.NewScheduleReconciler(cm.store)

	controller := controllers.NewResourceController(
		"schedule",
		workQueue,
		cm.store,
		reconciler,
		1, // max concurrent
	)

	cm.mu.Lock()
	cm.controllers["schedule"] = ManagedController{
		name:       "schedule",
		controller: controller,
		running:    false,
	}
	cm.mu.Unlock()

	log.Println("Registered schedule controller")
	return nil
}

// Start starts all registered controllers
func (cm *ControllerManager) Start(ctx context.Context) error {
	cm.mu.Lock()
	controllers := make([]ManagedController, 0, len(cm.controllers))
	for _, ctrl := range cm.controllers {
		controllers = append(controllers, ctrl)
	}
	cm.mu.Unlock()

	if len(controllers) == 0 {
		return fmt.Errorf("no controllers registered")
	}

	log.Printf("Starting %d controllers", len(controllers))

	// Start all controllers
	var wg sync.WaitGroup
	var errs []error
	var errMu sync.Mutex

	for _, managed := range controllers {
		wg.Add(1)
		go func(m ManagedController) {
			defer wg.Done()

			cm.mu.Lock()
			managed.running = true
			updated := managed
			cm.controllers[updated.name] = updated
			cm.mu.Unlock()

			if err := m.controller.Start(ctx); err != nil {
				errMu.Lock()
				errs = append(errs, err)
				errMu.Unlock()

				cm.mu.Lock()
				managed.running = false
				updated := managed
				cm.controllers[updated.name] = updated
				cm.mu.Unlock()
			}
		}(managed)
	}

	// Wait for all to finish
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("controllers failed: %v", errs)
	}

	return nil
}

// GetController returns a controller by name
func (cm *ControllerManager) GetController(name string) *controllers.ResourceController {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if managed, ok := cm.controllers[name]; ok {
		return managed.controller
	}
	return nil
}

// Status returns the status of all controllers
func (cm *ControllerManager) Status() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := make(map[string]interface{})

	for name, managed := range cm.controllers {
		status[name] = map[string]interface{}{
			"running": managed.running,
		}
	}

	return status
}

// Runtime orchestrates the entire system
type Runtime struct {
	apiServer     *apiserver.APIServer
	store         *apiserver.ResourceStore
	controllerMgr *ControllerManager
	mu            sync.Mutex
	running       bool
	version       string
}

// NewRuntime creates a new runtime
func NewRuntime(version string) *Runtime {
	store := apiserver.NewResourceStore()
	apiServer := apiserver.NewAPIServer(store)

	return &Runtime{
		apiServer:     apiServer,
		store:         store,
		controllerMgr: NewControllerManager(store, false, "default"),
		version:       version,
		running:       false,
	}
}

// Initialize sets up all components
func (r *Runtime) Initialize(ctx context.Context) error {
	log.Printf("Initializing runtime version %s", r.version)

	// Register all controllers
	if err := r.controllerMgr.RegisterWorkloadController(); err != nil {
		return err
	}
	if err := r.controllerMgr.RegisterPipelineController(); err != nil {
		return err
	}
	if err := r.controllerMgr.RegisterScheduleController(); err != nil {
		return err
	}

	return nil
}

// Start starts the runtime
func (r *Runtime) Start(ctx context.Context, apiAddr string) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("runtime already running")
	}
	r.running = true
	r.mu.Unlock()

	log.Printf("Starting runtime on %s", apiAddr)

	// Start API server in background
	go func() {
		if err := r.apiServer.Run(apiAddr); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	// Give API server time to start
	// In production, would use health checks

	// Start controller manager
	if err := r.controllerMgr.Start(ctx); err != nil {
		return err
	}

	return nil
}

// Stop stops the runtime
func (r *Runtime) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return fmt.Errorf("runtime not running")
	}

	r.running = false
	log.Println("Runtime stopped")

	return nil
}

// GetStore returns the resource store
func (r *Runtime) GetStore() *apiserver.ResourceStore {
	return r.store
}

// GetAPIServer returns the API server
func (r *Runtime) GetAPIServer() *apiserver.APIServer {
	return r.apiServer
}

// GetControllerManager returns the controller manager
func (r *Runtime) GetControllerManager() *ControllerManager {
	return r.controllerMgr
}

// Status returns runtime status
func (r *Runtime) Status() map[string]interface{} {
	return map[string]interface{}{
		"version":     r.version,
		"running":     r.running,
		"controllers": r.controllerMgr.Status(),
	}
}

// Probe represents a liveness/readiness probe
type Probe interface {
	Check(ctx context.Context) error
}

// LivenessProbe checks if runtime is alive
type LivenessProbe struct {
	runtime *Runtime
}

// NewLivenessProbe creates a liveness probe
func NewLivenessProbe(runtime *Runtime) *LivenessProbe {
	return &LivenessProbe{runtime: runtime}
}

// Check checks if runtime is alive
func (lp *LivenessProbe) Check(ctx context.Context) error {
	if !lp.runtime.running {
		return fmt.Errorf("runtime not running")
	}
	return nil
}

// ReadinessProbe checks if runtime is ready to serve
type ReadinessProbe struct {
	runtime *Runtime
}

// NewReadinessProbe creates a readiness probe
func NewReadinessProbe(runtime *Runtime) *ReadinessProbe {
	return &ReadinessProbe{runtime: runtime}
}

// Check checks if runtime is ready
func (rp *ReadinessProbe) Check(ctx context.Context) error {
	if !rp.runtime.running {
		return fmt.Errorf("runtime not ready")
	}

	// Check if all controllers are running
	status := rp.runtime.controllerMgr.Status()
	for name, ctrl := range status {
		if ctrlMap, ok := ctrl.(map[string]interface{}); ok {
			if running, ok := ctrlMap["running"].(bool); ok && !running {
				return fmt.Errorf("controller %s not running", name)
			}
		}
	}

	return nil
}
