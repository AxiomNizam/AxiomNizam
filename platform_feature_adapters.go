package main

import (
	"fmt"
	"strings"
	"time"

	"example.com/axiomnizam/internal/export"
	"example.com/axiomnizam/internal/lineage"
	"example.com/axiomnizam/internal/rbac"
	"example.com/axiomnizam/internal/tracing"
)

// exportManagerAdapter bridges handler interface to in-memory export manager.
type exportManagerAdapter struct {
	base *export.InMemoryExportManager
}

func (a *exportManagerAdapter) SubmitExport(job *export.ExportJob) (*export.ExportJob, error) {
	return a.base.SubmitExport(job)
}

func (a *exportManagerAdapter) GetExport(id string) (*export.ExportJob, error) {
	return a.base.GetExport(id)
}

func (a *exportManagerAdapter) ListExports(tenantID, status, format string) ([]*export.ExportJob, error) {
	items, err := a.base.ListExports(tenantID)
	if err != nil {
		return nil, err
	}

	filtered := make([]*export.ExportJob, 0, len(items))
	for _, item := range items {
		if status != "" && !strings.EqualFold(string(item.Status), status) {
			continue
		}
		if format != "" && !strings.EqualFold(string(item.Format), format) {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered, nil
}

func (a *exportManagerAdapter) CancelExport(id string) error {
	return a.base.CancelExport(id)
}

func (a *exportManagerAdapter) CreateTemplate(template *export.ExportTemplate) (*export.ExportTemplate, error) {
	return a.base.CreateTemplate(template)
}

func (a *exportManagerAdapter) ListTemplates(tenantID string) ([]*export.ExportTemplate, error) {
	return a.base.ListTemplates(tenantID)
}

// rbacManagerAdapter bridges handler interface to in-memory RBAC manager.
type rbacManagerAdapter struct {
	base *rbac.InMemoryRBACManager
}

func (a *rbacManagerAdapter) CreateRole(role *rbac.Role) (*rbac.Role, error) {
	return a.base.CreateRole(role)
}

func (a *rbacManagerAdapter) GetRole(id string) (*rbac.Role, error) {
	return a.base.GetRole(id)
}

func (a *rbacManagerAdapter) ListRoles(tenantID string) ([]*rbac.Role, error) {
	return a.base.ListRoles(tenantID)
}

func (a *rbacManagerAdapter) UpdateRole(role *rbac.Role) (*rbac.Role, error) {
	return a.base.UpdateRole(role)
}

func (a *rbacManagerAdapter) DeleteRole(id string) error {
	return a.base.DeleteRole(id)
}

func (a *rbacManagerAdapter) CreateRoleBinding(binding *rbac.RoleBinding) (*rbac.RoleBinding, error) {
	return a.base.CreateRoleBinding(binding)
}

func (a *rbacManagerAdapter) ListRoleBindings(tenantID, principalID string) ([]*rbac.RoleBinding, error) {
	items, err := a.base.ListRoleBindings("", principalID)
	if err != nil {
		return nil, err
	}

	filtered := make([]*rbac.RoleBinding, 0, len(items))
	for _, item := range items {
		if tenantID != "" && item.TenantID != tenantID {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered, nil
}

func (a *rbacManagerAdapter) DeleteRoleBinding(id string) error {
	return a.base.DeleteRoleBinding(id)
}

func (a *rbacManagerAdapter) CheckPermission(req *rbac.PermissionCheck) (*rbac.PermissionCheckResult, error) {
	allowed, err := a.base.CheckPermission(req.PrincipalID, req.Resource, req.Action)
	if err != nil {
		return nil, err
	}

	reason := "no matching permission"
	if allowed {
		reason = "allowed by role binding"
	}

	return &rbac.PermissionCheckResult{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}

func (a *rbacManagerAdapter) ListPermissions(tenantID, resource string) ([]*rbac.Permission, error) {
	items, err := a.base.ListPermissions("")
	if err != nil {
		return nil, err
	}

	filtered := make([]*rbac.Permission, 0, len(items))
	for _, item := range items {
		if tenantID != "" && item.TenantID != tenantID {
			continue
		}
		if resource != "" && item.Resource != resource {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered, nil
}

func (a *rbacManagerAdapter) CreateAccessRequest(req *rbac.AccessRequest) (*rbac.AccessRequest, error) {
	return a.base.CreateAccessRequest(req)
}

func (a *rbacManagerAdapter) ApproveAccessRequest(id, approvedBy string) (*rbac.AccessRequest, error) {
	if err := a.base.ApproveAccessRequest(id, approvedBy); err != nil {
		return nil, err
	}
	return a.base.GetAccessRequest(id)
}

// lineageManagerAdapter bridges handler interface to in-memory lineage manager.
type lineageManagerAdapter struct {
	base *lineage.InMemoryLineageManager
}

func (a *lineageManagerAdapter) GetNode(id string) (*lineage.LineageNode, error) {
	return a.base.GetNode(id)
}

func (a *lineageManagerAdapter) ListNodes(tenantID, resourceType string) ([]*lineage.LineageNode, error) {
	items, err := a.base.ListNodes("")
	if err != nil {
		return nil, err
	}

	filtered := make([]*lineage.LineageNode, 0, len(items))
	for _, item := range items {
		if tenantID != "" && item.TenantID != tenantID {
			continue
		}
		if resourceType != "" && item.ResourceType != resourceType {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered, nil
}

func (a *lineageManagerAdapter) GetLineageGraph(resourceType, resourceID string) (*lineage.LineageGraph, error) {
	return a.base.BuildGraph(resourceID, resourceType, 6)
}

func (a *lineageManagerAdapter) GetUpstreamLineage(resourceType, resourceID string) ([]*lineage.LineagePath, error) {
	nodes, err := a.base.GetUpstream(resourceID, 6)
	if err != nil {
		return nil, err
	}

	paths := make([]*lineage.LineagePath, 0, len(nodes))
	for _, node := range nodes {
		paths = append(paths, &lineage.LineagePath{
			ID:           fmt.Sprintf("path-upstream-%s-%s", node.ID, resourceID),
			SourceNodeID: node.ID,
			TargetNodeID: resourceID,
			Path:         []string{node.ID, resourceID},
			Length:       2,
			IsActive:     true,
		})
	}

	return paths, nil
}

func (a *lineageManagerAdapter) GetDownstreamLineage(resourceType, resourceID string) ([]*lineage.LineagePath, error) {
	nodes, err := a.base.GetDownstream(resourceID, 6)
	if err != nil {
		return nil, err
	}

	paths := make([]*lineage.LineagePath, 0, len(nodes))
	for _, node := range nodes {
		paths = append(paths, &lineage.LineagePath{
			ID:           fmt.Sprintf("path-downstream-%s-%s", resourceID, node.ID),
			SourceNodeID: resourceID,
			TargetNodeID: node.ID,
			Path:         []string{resourceID, node.ID},
			Length:       2,
			IsActive:     true,
		})
	}

	return paths, nil
}

func (a *lineageManagerAdapter) GetImpactAnalysis(resourceType, resourceID string) (*lineage.ImpactAnalysis, error) {
	impacts, err := a.base.AnalyzeImpact(resourceID)
	if err != nil {
		return nil, err
	}
	if len(impacts) == 0 {
		return &lineage.ImpactAnalysis{SourceNodeID: resourceID, EstimatedImpact: "low"}, nil
	}
	return impacts[0], nil
}

func (a *lineageManagerAdapter) GetColumnLineage(sourceCol, targetCol string) (*lineage.ColumnLineage, error) {
	nodeID := ""
	column := sourceCol
	if strings.Contains(sourceCol, ".") {
		parts := strings.SplitN(sourceCol, ".", 2)
		nodeID = parts[0]
		column = parts[1]
	}
	if nodeID == "" && strings.Contains(targetCol, ".") {
		parts := strings.SplitN(targetCol, ".", 2)
		nodeID = parts[0]
	}
	if nodeID == "" {
		return nil, fmt.Errorf("sourceColumn or targetColumn must include node identifier")
	}

	items, err := a.base.GetColumnLineage(nodeID, column)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("lineage not found")
	}

	for _, item := range items {
		if targetCol == "" || item.TargetColumn == targetCol {
			return item, nil
		}
	}

	return items[0], nil
}

func (a *lineageManagerAdapter) TraceDataFlow(sourceID, targetID string) (*lineage.LineagePath, error) {
	nodes, err := a.base.GetDownstream(sourceID, 8)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if node.ID == targetID {
			return &lineage.LineagePath{
				ID:           fmt.Sprintf("trace-%s-%s", sourceID, targetID),
				SourceNodeID: sourceID,
				TargetNodeID: targetID,
				Path:         []string{sourceID, targetID},
				Length:       2,
				IsActive:     true,
			}, nil
		}
	}
	return nil, fmt.Errorf("no data flow path found")
}

func (a *lineageManagerAdapter) GetStatistics(tenantID string) (*lineage.LineageStatistics, error) {
	stats, err := a.base.GetLineageStatistics()
	if err != nil {
		return nil, err
	}
	stats.TenantID = tenantID
	stats.LastUpdated = time.Now()
	return stats, nil
}

// tracingManagerAdapter bridges handler interface to in-memory tracing manager.
type tracingManagerAdapter struct {
	base *tracing.InMemoryTracingManager
}

func (a *tracingManagerAdapter) GetTrace(traceID string) (*tracing.Trace, error) {
	return a.base.GetTrace(traceID)
}

func firstServiceName(trace *tracing.Trace) string {
	if trace == nil || len(trace.Services) == 0 {
		return ""
	}
	return trace.Services[0]
}

func firstOperationName(trace *tracing.Trace) string {
	if trace == nil {
		return ""
	}
	if trace.Root.OperationName != "" {
		return trace.Root.OperationName
	}
	if len(trace.Spans) > 0 {
		return trace.Spans[0].OperationName
	}
	return ""
}

func (a *tracingManagerAdapter) SearchTraces(req *tracing.TraceSearchRequest) ([]*tracing.TraceSearchResult, error) {
	items, err := a.base.SearchTraces(req)
	if err != nil {
		return nil, err
	}

	results := make([]*tracing.TraceSearchResult, 0, len(items))
	for _, item := range items {
		results = append(results, &tracing.TraceSearchResult{
			TraceID:    item.ID,
			Service:    firstServiceName(item),
			Operation:  firstOperationName(item),
			StartTime:  item.StartTime,
			Duration:   item.Duration,
			SpanCount:  item.TotalSpans,
			ErrorCount: item.ErrorSpans,
			Status:     item.Status,
		})
	}
	return results, nil
}

func (a *tracingManagerAdapter) GetSpan(spanID string) (*tracing.Span, error) {
	return a.base.GetSpan(spanID)
}

func (a *tracingManagerAdapter) GetServiceMap(tenantID string) (*tracing.ServiceMap, error) {
	depsMap, err := a.base.GetServiceMap()
	if err != nil {
		return nil, err
	}

	svcSet := make(map[string]bool)
	deps := make([]tracing.DependencyMetrics, 0)
	for source, sourceDeps := range depsMap {
		svcSet[source] = true
		for _, dep := range sourceDeps {
			svcSet[dep.Destination] = true
			deps = append(deps, *dep)
		}
	}

	services := make([]tracing.ServiceInfo, 0, len(svcSet))
	for name := range svcSet {
		services = append(services, tracing.ServiceInfo{
			Name:     name,
			Status:   "healthy",
			LastSeen: time.Now(),
		})
	}

	return &tracing.ServiceMap{
		TenantID:     tenantID,
		Timestamp:    time.Now(),
		Services:     services,
		Dependencies: deps,
	}, nil
}

func (a *tracingManagerAdapter) GetServiceMetrics(service string) (*tracing.TraceMetrics, error) {
	return a.base.GetServiceMetrics(service)
}

func (a *tracingManagerAdapter) GetOperationMetrics(service, operation string) (*tracing.SpanMetrics, error) {
	return a.base.GetOperationMetrics(service, operation)
}

func (a *tracingManagerAdapter) ListServices(tenantID string) ([]*tracing.ServiceInfo, error) {
	serviceMap, err := a.GetServiceMap(tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]*tracing.ServiceInfo, 0, len(serviceMap.Services))
	for i := range serviceMap.Services {
		svc := serviceMap.Services[i]
		items = append(items, &svc)
	}
	return items, nil
}

func (a *tracingManagerAdapter) GetErrorAnalysis(tenantID, service string) (*tracing.ErrorAnalysis, error) {
	items, err := a.base.AnalyzeErrors(service)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return &tracing.ErrorAnalysis{
			TenantID:  tenantID,
			ErrorType: "none",
			Count:     0,
		}, nil
	}

	analysis := items[0]
	analysis.TenantID = tenantID
	return analysis, nil
}
