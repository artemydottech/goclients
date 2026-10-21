package models

const MaxPageLimit = 200

type Page struct {
	Limit  int
	Offset int
}

func (p Page) Validate() error {
	if p.Limit < 0 {
		return Invalid("limit cannot be negative")
	}

	if p.Limit > MaxPageLimit {
		return Invalid("limit cannot exceed %d", MaxPageLimit)
	}

	if p.Offset < 0 {
		return Invalid("offset cannot be negative")
	}

	return nil
}

func (p Page) Unlimited() bool {
	return p.Limit == 0 && p.Offset == 0
}
