package decoder

type AuthorizedCategory string

const (
	AUTHORIZED_CATEGORY_LANGGENIUS AuthorizedCategory = "langgenius"
	AUTHORIZED_CATEGORY_PARTNER    AuthorizedCategory = "partner"
	AUTHORIZED_CATEGORY_COMMUNITY  AuthorizedCategory = "community"
)

type Verification struct {
	AuthorizedCategory AuthorizedCategory `json:"authorized_category"`
}

func DefaultVerification() *Verification {
	return &Verification{
		AuthorizedCategory: AUTHORIZED_CATEGORY_LANGGENIUS,
	}
}
