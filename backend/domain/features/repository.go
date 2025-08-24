package features

// Repository defines the interface for feature data operations
type Repository interface {
	Create(feature *Feature) error
	GetByID(id int, userID *int) (*Feature, error)
	GetAll(page, perPage int, userID *int) ([]Feature, int, error)
	GetByCreatedBy(userID int) ([]Feature, error)
	Update(id int, title, description *string) error
	Delete(id int) error
	FeatureExists(id int) (bool, error)
}