package db

func (s *storage) ListTags() ([]Tag, error) {
	rows, err := s.pg.Query("SELECT id, name FROM tags")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tags := make([]Tag, 0)

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *storage) CreateTag(tag Tag) (*Tag, error) {
	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		RETURNING id
	`

	if err := s.pg.QueryRow(query, tag.Name).Scan(&tag.ID); err != nil {
		return nil, err
	}

	return &tag, nil
}

func (s *storage) DeleteTag(id int64) error {
	query := `
		DELETE FROM tags
		WHERE id = $1
	`

	if _, err := s.pg.Exec(query, id); err != nil {
		return err
	}

	return nil
}
