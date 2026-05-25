package anonymization

import "example.com/axiomnizam/internal/anonymization/models"

// Re-export domain Resource types from models subpackage.
type AnonymizationPolicyResource = models.AnonymizationPolicyResource
type AnonymizationPolicySpec = models.AnonymizationPolicySpec
type AnonymizationPolicyResourceStatus = models.AnonymizationPolicyResourceStatus
type AnonymRule = models.AnonymRule
type PolicyScope = models.PolicyScope
type MaskTechnique = models.MaskTechnique

const (
	AnonymizationPolicyKind       = models.AnonymizationPolicyKind
	AnonymizationPolicyAPIVersion = models.AnonymizationPolicyAPIVersion
)

const (
	MaskHash       = models.MaskHash
	MaskRedact     = models.MaskRedact
	MaskPartial    = models.MaskPartial
	MaskTokenize   = models.MaskTokenize
	MaskNoise      = models.MaskNoise
	MaskGeneralize = models.MaskGeneralize
	MaskSynthetic  = models.MaskSynthetic
	MaskShuffle    = models.MaskShuffle
)
