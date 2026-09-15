package screen

type Cursor struct {
	Row int
	Col int
}

func (c Cursor) Equals(other Cursor) bool {
	return c.Row == other.Row && c.Col == other.Col
}

/* // adjustCursorForDeletion updates the cursor position to reflect the deleted text within the buffer.
func (c *Cursor) AdjustForDeletion(deleteStart, deleteEnd Cursor) {
	// If the cursor is before the deletion row, no adjustment is needed.
	if c.Row < deleteStart.Row {
		return
	}

	// If the cursor is on the same row as the deletion start:
	if c.Row == deleteStart.Row {
		// If the deletion is after the cursor column, leave the cursor unchanged.
		if c.Col <= deleteStart.Col {
			return
		}
		// If the deletion affects the cursor's position within the same row, move it to the start of deletion.
		c.Col = deleteStart.Col
		return
	}

	// If the cursor is after the deletion row range:
	if c.Row > deleteEnd.Row {
		// Adjust the row index to account for the rows deleted.
		c.Row -= deleteEnd.Row - deleteStart.Row
		return
	}

	// If the cursor is within the deleted range, move it to the start of the deletion.
	c.Row = deleteStart.Row
	c.Col = deleteStart.Col
}

// adjustCursorForInsertion updates the cursor position to reflect the inserted text within the buffer.
func (c *Cursor) AdjustForInsertion(insertStart, insertEnd Cursor) {
	// If the cursor is before the insertion row, no adjustment is needed.
	if c.Row < insertStart.Row {
		return
	}

	// Calculate the number of rows added by the insertion.
	rowOffset := insertEnd.Row - insertStart.Row

	// If the insertion is on the same row as the cursor:
	if c.Row == insertStart.Row {
		// If the insertion is before the cursor in the same row, adjust the cursor column.
		if c.Col >= insertStart.Col {
			// If the insertion doesn't span multiple rows, adjust only the column.
			if rowOffset == 0 {
				c.Col += insertEnd.Col - insertStart.Col
			} else {
				// Adjust column to the end of the insertion in the current row, then add row offset.
				c.Col = c.Col - insertStart.Col + insertEnd.Col
				c.Row += rowOffset
			}
		}
		return
	}

	// If the cursor is after the insertion row(s), increment the row index.
	c.Row += rowOffset
}
*/
