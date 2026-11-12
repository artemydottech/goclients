package repository

import "context"

func (r *ClientRepository) GetClientTags(ctx context.Context, id int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT tag FROM client_tags WHERE client_id=? ORDER BY tag", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := []string{}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *ClientRepository) ReplaceClientTags(ctx context.Context, id int, tags []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "DELETE FROM client_tags WHERE client_id=?", id); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, "INSERT INTO client_tags(client_id,tag) VALUES(?,?)", id, tag); err != nil {
			return err
		}
	}
	return tx.Commit()
}
