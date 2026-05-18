package challenge

// ChallengePhase represents the lifecycle phase of a challenge.
type ChallengePhase string

const (
	PhaseWaiting   ChallengePhase = "Waiting"
	PhaseVerified  ChallengePhase = "Verified"
	PhaseExpired   ChallengePhase = "Expired"
	PhaseFailed    ChallengePhase = "Failed"
	PhaseRejected  ChallengePhase = "Rejected"
)

// StateMachine manages challenge phase transitions.
type StateMachine struct {
	Current     ChallengePhase
	Attempts    int
	MaxAttempts int
}

// NewStateMachine creates a new state machine for a challenge.
func NewStateMachine(phase string, attempts, maxAttempts int) *StateMachine {
	return &StateMachine{
		Current:     ChallengePhase(phase),
		Attempts:    attempts,
		MaxAttempts: maxAttempts,
	}
}

// CanTransitionTo checks if a transition to the target phase is valid.
func (m *StateMachine) CanTransitionTo(target ChallengePhase) bool {
	switch m.Current {
	case PhaseWaiting:
		return target == PhaseVerified || target == PhaseExpired || target == PhaseFailed || target == PhaseRejected
	case PhaseVerified, PhaseExpired, PhaseFailed, PhaseRejected:
		return false // Terminal states
	default:
		return false
	}
}

// Transition attempts to move to the target phase.
func (m *StateMachine) Transition(target ChallengePhase) bool {
	if !m.CanTransitionTo(target) {
		return false
	}
	m.Current = target
	return true
}

// RecordAttempt increments the attempt counter and may transition to Failed.
func (m *StateMachine) RecordAttempt() {
	m.Attempts++
	if m.Attempts >= m.MaxAttempts {
		m.Current = PhaseFailed
	}
}

// IsTerminal returns true if the state machine is in a final state.
func (m *StateMachine) IsTerminal() bool {
	return m.Current == PhaseVerified ||
		m.Current == PhaseExpired ||
		m.Current == PhaseFailed ||
		m.Current == PhaseRejected
}
