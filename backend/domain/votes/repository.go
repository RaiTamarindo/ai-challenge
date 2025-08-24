package votes

// Repository defines the interface for vote data operations
type Repository interface {
	AddVote(userID, featureID int) error
	RemoveVote(userID, featureID int) error
	HasUserVoted(userID, featureID int) (bool, error)
	GetUserVotes(userID int) ([]Vote, error)
}