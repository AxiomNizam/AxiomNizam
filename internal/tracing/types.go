package tracing

import "example.com/axiomnizam/internal/tracing/models"

// Re-exported resource types from models/.
type TracingConfigResource = models.TracingConfigResource
type TracingConfigSpec = models.TracingConfigSpec
type TracingConfigResourceStatus = models.TracingConfigResourceStatus

// Re-exported constants.
const TracingConfigKind = models.TracingConfigKind
const TracingConfigAPIVersion = models.TracingConfigAPIVersion
