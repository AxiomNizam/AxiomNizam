package contracts

// ServiceRegistry provides a lookup mechanism for all Gatekeeper services.
// Unlike Provider (which is a concrete struct), this is an interface
// for use by middleware and handlers that need dynamic service access.
type ServiceRegistry interface {
	GetEnrollmentService() EnrollmentService
	GetChallengeService() ChallengeService
	GetFactorService() FactorService
}
